package brc

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/edsrzf/mmap-go"
)

var wg sync.WaitGroup

// Convert a log file of readings into a map of measurements
func ProcessLogFile(file *os.File) (map[string]Measurement, error) {
	// mmap a file
	data, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}
	//nolint:all
	defer data.Unmap()
	log.Printf("Data Size: %v", len(data))

	workers := runtime.NumCPU()
	chunkSize := 64 * 1024 * 1024
	chunks := len(data) / chunkSize

	chunksQueue := make(chan []byte, chunks)
	resultsQueue := make(chan map[string]Measurement, chunks)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go chunkWorker(chunksQueue, resultsQueue)
	}

	// send chunks aligned on \n char to workers
	start := 0
	end := -1
	for end < len(data)-1 { // -1 since the last char is an extra new line
		start = end + 1
		end = seekToNewLine(data, end+chunkSize)
		chunksQueue <- data[start : end+1]
	}
	close(chunksQueue)

	// go process to close the results queue to signal all the consumers are done
	go func() {
		wg.Wait()
		close(resultsQueue)
	}()

	// collect and reduce results from go routines
	m := make(map[string]Measurement, 512)

	for {
		chunkResults, ok := <-resultsQueue
		if !ok {
			break
		}
		log.Printf("Received chunk results from a process")

		for city, measure := range chunkResults {
			combineMeasurements(city, measure, m)
		}
	}
	log.Print("All Done")
	return m, nil
}

func chunkWorker(chunksQueue <-chan []byte, resultQueue chan<- map[string]Measurement) {
	defer wg.Done()
	results := make(map[string]Measurement, 256)
	for data := range chunksQueue {
		lineStart := 0

		lineEnd := 0
		for lineEnd < len(data)-1 {
			lineLen := bytes.IndexByte(data[lineStart:], '\n')
			if lineLen == -1 {
				break
			}

			line := data[lineStart:(lineStart + lineLen)]

			measurement, city, err := lineToMeasurement(line)
			if err != nil {
				log.Fatalf("errored: %v", err.Error())
			}
			combineMeasurements(city, measurement, results)

			lineStart += lineLen + 1
		}
	}
	resultQueue <- results
}

// returns the index of the first /n char after the start position in the byte array
func seekToNewLine(data []byte, start int) int {
	if start > len(data) {
		return len(data) - 1
	}
	idx := bytes.IndexByte(data[start:], '\n')
	if idx == -1 {
		return len(data) - 1
	}
	return idx + start
}

// converts a line in the format: "city;xx.x" into a measurement object
// returns the Measurement and city.
func lineToMeasurement(line []byte) (Measurement, string, error) {
	splitIdx := bytes.IndexByte(line, ';')
	if splitIdx == -1 {
		return Measurement{}, "", fmt.Errorf("line split produced more than 2 splits:  %v", line)
	}

	city := line[:splitIdx]
	measureString := line[splitIdx+1:]

	measure := bytesToInt(measureString)

	measurement := Measurement{
		minShifted: measure,
		maxShifted: measure,
		sumShifted: measure,
		Count:      1,
	}
	return measurement, string(city), nil
}

// Merge the new measurement into the map of measurements.
func combineMeasurements(city string, newMeasurement Measurement, m map[string]Measurement) {
	if currentMeasure, found := m[city]; found {
		currentMeasure.Merge(newMeasurement)
		m[city] = currentMeasure
	} else {
		m[city] = newMeasurement
	}
}

// takes a []byte array representing a string such as "-23.3" and returns the
// number multiplied by 10 in as a int64 (ex 233)
func bytesToInt(measure []byte) int64 {
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
	if negative {
		temp = -temp
	}
	return temp
}
