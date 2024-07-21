package brc

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
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
	chunkSize := len(data) / workers
	log.Printf("Chunk Size: %v", chunkSize)

	chunksQueue := make(chan []byte, workers)
	resultsQueue := make(chan map[string]Measurement, workers)

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

func chunkWorker(dataQueue chan []byte, resultQueue chan map[string]Measurement) {
	defer wg.Done()
	for data := range dataQueue {
		results := make(map[string]Measurement, 256)
		lineStart := 0

		lineEnd := 0
		for lineEnd < len(data)-1 {
			lineLen := bytes.IndexByte(data[lineStart:], '\n')
			if lineLen == -1 {
				break
			}

			line := string(data[lineStart:(lineStart + lineLen)])

			measurement, city, err := lineToMeasurement(line)
			if err != nil {
				log.Fatalf("errored: %v", err.Error())
			}
			combineMeasurements(city, measurement, results)

			lineStart += lineLen + 1
		}
		resultQueue <- results
	}
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
// returns teh Measurement and city.
func lineToMeasurement(line string) (Measurement, string, error) {
	splitIdx := strings.IndexByte(line, ';')
	if splitIdx == -1 {
		return Measurement{}, "", fmt.Errorf("line split produced more than 2 splits:  %v", line)
	}

	city := line[:splitIdx]
	measureString := line[splitIdx+1:]

	decimalIdx := strings.IndexByte(measureString, '.')
	measure, err := strconv.ParseInt(measureString[:decimalIdx], 10, 64)
	if err != nil {
		return Measurement{}, "", fmt.Errorf("measure is not a number:  %v", measureString)
	}

	decimal, err := strconv.ParseInt(measureString[decimalIdx+1:], 10, 64)
	if err != nil {
		return Measurement{}, "", fmt.Errorf("decimal is not a number:  %v", measureString)
	}

	total := measure*10 + decimal // we know decimal is a number 0 - 9

	measurement := Measurement{
		minShifted: total,
		maxShifted: total,
		sumShifted: total,
		Count:      1,
	}
	return measurement, city, nil
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

// TODO: lets try different worker sizes to see if theat helps lower the scheduler and GC times
// TODO: try to set the go GC at a large default min space size.
// TODO: Add tracing to see Heap size over time
// TODO: what is runtime.asyncPreempt?

// TODO: looks like map accesses are the slowest part right now (with strings). So maybe we change it to an idx style lookup?
// TODO: SIMD -- For string splitting (finding \n and other chars)
