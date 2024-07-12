package brc

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Convert a log file of readings into a map of measurements
func ProcessLogFile(file *os.File) (map[string]*Measurement, error) {
	coreCount := runtime.NumCPU()
	log.Printf("Core Count: %v", coreCount)

	// go process to push lines into a string queue
	linesQueue := make(chan string, 100000000)
	go readLines(file, linesQueue)

	// process each chunk in its own go routine
	resultsQueue := make(chan map[string]*Measurement, coreCount)
	for i := 0; i < coreCount-1; i++ {
		go lineWorker(linesQueue, resultsQueue)
	}

	// collect results from go routines
	m := make(map[string]*Measurement)

	// TODO: instead of this for loop with counts, we could (and maybe should) use some sort of wait group.
	for i := 0; i < coreCount-1; i++ {
		chunkResults := <-resultsQueue
		log.Printf("Received chunk results from a process")

		for city, measure := range chunkResults {
			combineMeasurements(city, measure, m)
		}
	}
	log.Print("All Done")
	return m, nil
}

// Read the liens of the file and push the lines int a queue for workers to process
func readLines(file *os.File, linesQueue chan string) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines) // built in splitter to move iterator until next /n char
	count := int64(0)
	for scanner.Scan() {
		count++

		line := scanner.Text()
		linesQueue <- line // push line into line queue

		if count%100000000 == 0 {
			log.Printf("Processed %v lines", count)
		}
	}
	close(linesQueue) // close to mark the file as done reading
}

// A worker to process lines as they are added to the queue
func lineWorker(linesQueue chan string, resultsChan chan map[string]*Measurement) {
	results := make(map[string]*Measurement)
	count := 0
	for {
		line, ok := <-linesQueue
		if !ok {
			log.Printf("No more lines to process ending")
			break
		}

		count++
		measurement, city, err := lineToMeasurement(line)
		if err != nil {
			panic(err)
		}
		combineMeasurements(city, measurement, results)
	}

	// publish results
	log.Printf("This worker processed %v lines", count)
	resultsChan <- results
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
// TODO: For attempt 3, lets just read lines into memory (chuncking) and processing the lines in another go routine
// TODO: time a single program to see how fast it is to mmmap the file, then read each line
