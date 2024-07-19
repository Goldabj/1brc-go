package brc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombineMeasurementsHappyPath(t *testing.T) {
	m1 := make(map[string]*Measurement)

	measure1 := Measurement{
		Min:   1,
		Max:   10,
		Sum:   11,
		Count: 2,
	}
	m1["city"] = &measure1

	measure2 := Measurement{
		Min:   4,
		Max:   13,
		Sum:   12,
		Count: 3,
	}

	combineMeasurements("city", &measure2, m1)

	merged := m1["city"]
	assert.Equal(t, 1.0, merged.Min)
	assert.Equal(t, 13.0, merged.Max)
	assert.Equal(t, 23.0, merged.Sum)
	assert.Equal(t, int64(5), merged.Count)
}

func TestLineToMeasurementHappyPath(t *testing.T) {
	line := "mycity;23.0"

	measure, city, error := lineToMeasurement(line)
	if error != nil {
		assert.Fail(t, error.Error())
	}

	assert.Equal(t, "mycity", city)
	assert.Equal(t, 23.0, measure.Max)
	assert.Equal(t, 23.0, measure.Min)
	assert.Equal(t, 23.0, measure.Sum)
	assert.Equal(t, int64(1), measure.Count)
}

func TestLineToMeasurementWithFloatingPointNumber(t *testing.T) {
	line := "mycity;22.3"

	measure, city, error := lineToMeasurement(line)
	if error != nil {
		assert.Fail(t, error.Error())
	}

	assert.Equal(t, "mycity", city)
	assert.Equal(t, 22.3, measure.Max)
	assert.Equal(t, 22.3, measure.Min)
	assert.Equal(t, 22.3, measure.Sum)
	assert.Equal(t, int64(1), measure.Count)
}

func TestProcessLogHappyPath(t *testing.T) {
	file, err := os.Open("../../test/data/small-log.txt")
	if err != nil {
		t.Error(err.Error())
	}
	defer file.Close()

	measurementMap, error := ProcessLogFile(file)
	if error != nil {
		t.Error(error.Error())
		return
	}

	// Palermo
	measurement := measurementMap["Palermo"]
	assert.Equal(t, 7.3, measurement.Sum)
	assert.Equal(t, int64(3), measurement.Count)
	assert.Equal(t, 1.1, measurement.Min)
	assert.Equal(t, 3.1, measurement.Max)

	// Harbin
	measurement = measurementMap["Harbin"]
	assert.Equal(t, 21.3, measurement.Sum)
	assert.Equal(t, int64(3), measurement.Count)
	assert.Equal(t, 1.1, measurement.Min)
	assert.Equal(t, 10.1, measurement.Max)

	// Tallinn
	measurement = measurementMap["Tallinn"]
	assert.Equal(t, 46.3, measurement.Sum)
	assert.Equal(t, int64(3), measurement.Count)
	assert.Equal(t, 12.1, measurement.Min)
	assert.Equal(t, 17.1, measurement.Max)

	// Banjul
	measurement = measurementMap["Banjul"]
	assert.Equal(t, 55.0, measurement.Sum)
	assert.Equal(t, int64(3), measurement.Count)
	assert.Equal(t, 5.0, measurement.Min)
	assert.Equal(t, 25.0, measurement.Max)

	// Boston
	measurement = measurementMap["Boston"]
	assert.Equal(t, 12.0, measurement.Sum)
	assert.Equal(t, int64(3), measurement.Count)
	assert.Equal(t, 4.0, measurement.Min)
	assert.Equal(t, 4.0, measurement.Max)
}

// BenchMark Testing
func BenchmarkProcessor(b *testing.B) {
	file, err := os.Open("../../test/data/measurements-1m.txt")
	if err != nil {
		b.Error(err.Error())
	}
	defer file.Close()

	for i := 0; i < b.N; i++ {
		ProcessLogFile(file)
	}
}
