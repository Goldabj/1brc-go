package brc

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

// Convert a log file of readings into a map of measurements
func ProcessLogFile(file *os.File) (map[string]*Measurement, error) {
	chunkSize := 10000000

	// go process to push lines into a string queue
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines) // built in splitter to move iterator until next /n char
	count := int64(0)

	lines := make([]string, 0, chunkSize)
	resultsQueue := make(chan map[string]*Measurement, 10000)

	for scanner.Scan() {
		count++

		line := scanner.Text()
		lines = append(lines, line)
		if count%int64(chunkSize) == 0 {
			wg.Add(1)
			go processChunk(lines, resultsQueue)
			lines = make([]string, 0, chunkSize)
		}

		if count%100000000 == 0 {
			log.Printf("Processed %v lines", count)
		}
	}
	// just in case, finish any left overs
	wg.Add(1)
	go processChunk(lines, resultsQueue)

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

func processChunk(lines []string, resultQueue chan map[string]*Measurement) {
	defer wg.Done()
	log.Printf("Starting new worker process")

	results := make(map[string]*Measurement)
	for _, line := range lines {
		measurement, city, err := lineToMeasurement(line)
		if err != nil {
			panic(err)
		}
		combineMeasurements(city, measurement, results)
	}
	resultQueue <- results
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

// TODO: look up mmaping again

// TODO: profile to see largest time spends
// TODO: time a simple program to see how fast it is just to read the file with a scanner
// TODO: time a single program to see how fast it is to mmmap the file, then read each line
