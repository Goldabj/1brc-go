package brc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeHappyPath(t *testing.T) {
	measure1 := Measurement{
		minShifted: 20,
		maxShifted: 120,
		sumShifted: 240,
		Count:      4,
	}

	measure2 := Measurement{
		minShifted: 10,
		maxShifted: 110,
		sumShifted: 120,
		Count:      2,
	}

	err := measure1.Merge(measure2)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, 1.0, measure1.Min())
	assert.Equal(t, 12.0, measure1.Max())
	assert.Equal(t, 36.0, measure1.Sum())
	assert.Equal(t, int64(6), measure1.Count)
	assert.Equal(t, 6.0, measure1.Avg())
}

func TestMergeWithFloatingPoints(t *testing.T) {
	measure1 := Measurement{
		minShifted: 122,
		maxShifted: 122,
		sumShifted: 122,
		Count:      1,
	}

	measure2 := Measurement{
		minShifted: 43,
		maxShifted: 43,
		sumShifted: 43,
		Count:      2,
	}

	err := measure1.Merge(measure2)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, 4.3, measure1.Min())
	assert.Equal(t, 12.2, measure1.Max())
	assert.Equal(t, 16.5, measure1.Sum())
	assert.Equal(t, int64(3), measure1.Count)
	assert.Equal(t, 5.5, measure1.Avg())
}
