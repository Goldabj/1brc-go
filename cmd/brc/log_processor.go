package brc

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Convert a log file of readings into a map of measurements
func ProcessLogFile(file *os.File) (map[string]*Measurement, error) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines) // built in splitter to move iterator until next /n char

	m := make(map[string]*Measurement)

	count := 0
	for {
		read := scanner.Scan()
		count++

		if !read {
			break
		}

		line := scanner.Text()
		measurement, city, err := lineToMeasurement(line)
		if err != nil {
			return nil, err
		}

		combineMeasurements(city, measurement, m)

		if count%100000000 == 0 {
			log.Printf("Processed %v lines", count)
		}

	}

	return m, nil
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
