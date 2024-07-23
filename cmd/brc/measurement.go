package brc

import (
	"math"
)

type Measurement struct {
	minShifted int64
	maxShifted int64
	sumShifted int64
	Count      int64
}

func (m *Measurement) Merge(other Measurement) {
	m.minShifted = min(m.minShifted, other.minShifted)
	m.maxShifted = max(m.maxShifted, other.maxShifted)
	m.sumShifted = m.sumShifted + other.sumShifted
	m.Count = m.Count + other.Count
}

func (m *Measurement) Min() float64 {
	return float64(m.minShifted) / 10.0
}

func (m *Measurement) Max() float64 {
	return float64(m.maxShifted) / 10.0
}

func (m *Measurement) Sum() float64 {
	return float64(m.sumShifted) / 10.0
}

func (m *Measurement) Avg() float64 {
	avg := m.Sum() / float64(m.Count)
	return math.Round(avg*100) / 100.0 // round to 2 decimals
}
