package brc

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/edsrzf/mmap-go"
)

var wg sync.WaitGroup

// Convert a log file of readings into a map of measurements
func ProcessLogFile(file *os.File) (map[string]*Measurement, error) {
	// mmap a file
	data, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}
	//nolint:all
	defer data.Unmap()
	log.Printf("Data Size: %v", len(data))

	resultsQueue := make(chan map[string]*Measurement, 100)

	workers := 8
	chunkSize := len(data) / workers
	log.Printf("Chunk Size: %v", chunkSize)
	start := 0
	end := -1
	for end < len(data)-1 { // -1 since the last char is an extra new line
		start = end + 1
		end = seekToNewLine(data, end+chunkSize)
		wg.Add(1)
		go processChunk(data[start:end+1], resultsQueue)
	}

	// go process to close the results queue to signal all the consumers are done
	go func() {
		wg.Wait()
		close(resultsQueue)
	}()

	// collect results from go routines
	m := make(map[string]*Measurement)

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

func processChunk(data []byte, resultQueue chan map[string]*Measurement) {
	defer wg.Done()
	log.Printf("Starting new worker process with %v data", len(data))

	results := make(map[string]*Measurement)
	lineStart := 0
	for i := 1; i < len(data); i++ {
		if data[i] != '\n' {
			continue
		}

		line := string(data[lineStart:i])

		measurement, city, err := lineToMeasurement(line)
		if err != nil {
			log.Fatalf("errored: %v", err.Error())
		}
		combineMeasurements(city, measurement, results)

		lineStart = i + 1
	}
	resultQueue <- results
}

// returns the index of the first /n char after the start position in the byte array
func seekToNewLine(data []byte, start int) int {
	for i := start; i < len(data); i++ {
		if data[i] == '\n' {
			return i
		}
	}
	return len(data) - 1
}

// converts a line in the format: "city;xx.x" into a measurement object
// returns teh Measurement and city.
func lineToMeasurement(line string) (*Measurement, string, error) {
	splits := strings.Split(line, ";")
	if len(splits) != 2 {
		return nil, "", fmt.Errorf("line split produced more than 2 splits:  %v", line)
	}
	city := splits[0]
	measureString := splits[1]
	measure, err := strconv.ParseFloat(measureString, 64)
	if err != nil {
		return nil, "", fmt.Errorf("measure is not a number:  %v", measureString)
	}

	measurement := Measurement{
		Min:   measure,
		Max:   measure,
		Sum:   measure,
		Count: 1,
	}
	return &measurement, city, nil
}

// Merge the new measurement into the map of measurements.
func combineMeasurements(city string, newMeasurement *Measurement, m map[string]*Measurement) {
	if currentMeasure, found := m[city]; found {
		currentMeasure.Merge(newMeasurement)
	} else {
		m[city] = newMeasurement
	}
}

// TODO: profile to see largest time spends
