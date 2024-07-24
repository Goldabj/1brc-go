package brc

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/dolthub/swiss"
)

var wg sync.WaitGroup

// Convert a log file of readings into a map of measurements
func ProcessLogFile(file *os.File) (*swiss.Map[string, Measurement], error) {
	workers := runtime.NumCPU() - 1
	chunkSize := 64 * 1024 * 1024

	chunksQueue := make(chan []byte)
	resultsQueue := make(chan *swiss.Map[string, Measurement], 12)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go chunkWorker(chunksQueue, resultsQueue)
	}

	// go process to close the results queue to signal all the consumers are done
	go func() {
		buf := make([]byte, chunkSize)
		leftover := make([]byte, 0, chunkSize)
		for {
			readTotal, err := file.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				panic(err)
			}
			buf = buf[:readTotal]

			lastNewLineIndex := bytes.LastIndex(buf, []byte{'\n'})

			toSend := append(leftover, buf[:lastNewLineIndex+1]...)
			leftover = make([]byte, len(buf[lastNewLineIndex+1:]))
			copy(leftover, buf[lastNewLineIndex+1:])

			chunksQueue <- toSend

		}
		close(chunksQueue)

		// wait for workers to complete
		wg.Wait()
		close(resultsQueue)
	}()

	// collect and reduce results from go routines
	m := swiss.NewMap[string, Measurement](512)

	for {
		chunkResults, ok := <-resultsQueue
		if !ok {
			break
		}

		chunkResults.Iter(func(city string, v Measurement) (stop bool) {
			combineMeasurements(city, v, m)
			return false // continue
		})
	}
	log.Print("All Done")

	return m, nil
}

func chunkWorker(chunksQueue <-chan []byte, resultQueue chan<- *swiss.Map[string, Measurement]) {
	defer wg.Done()
	results := swiss.NewMap[string, Measurement](256)
	for data := range chunksQueue {
		for i := 0; i < len(data); {
			measure, city, bytesRead, err := lineToMeasure(data[i:])
			if err != nil {
				log.Fatal("Failed to create measure", err)
			}

			combineMeasurements(city, measure, results)
			i += bytesRead
		}
	}
	resultQueue <- results
}

// converts a line in the format: "city;xx.x\n" into a measurement object.
// It stops when the first \n char is reached.
// returns the Measurement, city, and bytes read
func lineToMeasure(line []byte) (Measurement, string, int, error) {
	for idx, char := range line {
		switch char {
		case ';':
			city := line[:idx]
			sample, bytesRead := bytesToInt(line[idx+1:])
			measurement := Measurement{
				minShifted: sample,
				maxShifted: sample,
				sumShifted: sample,
				Count:      1,
			}
			return measurement, string(city), idx + bytesRead + 2, nil
		}
	}
	return Measurement{}, "", 0, errors.New("failed to parse line")

}

// Merge the new measurement into the map of measurements.
func combineMeasurements(city string, newMeasurement Measurement, m *swiss.Map[string, Measurement]) {
	if currentMeasure, found := m.Get(city); found {
		currentMeasure.Merge(newMeasurement)
		m.Put(city, currentMeasure)
	} else {
		m.Put(city, newMeasurement)
	}
}

// takes a []byte array representing a string such as "-23.3" and returns the
// number multiplied by 10 in as a int64 (ex 233)
// returns the number, and bytes read
func bytesToInt(measure []byte) (int64, int) {
	negative := false
	index := 0
	if measure[index] == '-' {
		index++
		negative = true
	}
	temp := int64(measure[index]-'0') * 10 // parse first digit
	index++
	if measure[index] != '.' {
		temp = temp*10 + int64(measure[index]-'0')*10 // parse optional second digit
		index++
	}
	index++                             // skip '.'
	temp += int64(measure[index] - '0') // parse decimal digit
	index++
	if negative {
		temp = -temp
	}
	return temp, index
}

// TODO: Instead of storing the city as the key to our map (which needs to have its hash re-computed often) it would be much more
// efficient to use an integer as the key.
