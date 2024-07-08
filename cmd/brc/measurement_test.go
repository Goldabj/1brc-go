package brc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeHappyPath(t *testing.T) {
	measure1 := Measurement{
		Min:   2.0,
		Max:   12.0,
		Sum:   24,
		Count: 4,
	}

	measure2 := Measurement{
		Min:   1,
		Max:   11,
		Sum:   12,
		Count: 2,
	}

	measure1.Merge(&measure2)

	assert.Equal(t, 1.0, measure1.Min)
	assert.Equal(t, 12.0, measure1.Max)
	assert.Equal(t, 36.0, measure1.Sum)
	assert.Equal(t, int64(6), measure1.Count)
}

func TestMergeWithFloatingPoints(t *testing.T) {
	measure1 := Measurement{
		Min:   12.2,
		Max:   12.2,
		Sum:   12.2,
		Count: 1,
	}

	measure2 := Measurement{
		Min:   4.32,
		Max:   4.32,
		Sum:   4.32,
		Count: 2,
	}

	measure1.Merge(&measure2)

	assert.Equal(t, 4.32, measure1.Min)
	assert.Equal(t, 12.2, measure1.Max)
	assert.Equal(t, 16.52, measure1.Sum)
	assert.Equal(t, int64(3), measure1.Count)
}
