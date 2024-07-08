package brc

import "math"

type Measurement struct {
	Min   float64
	Max   float64
	Sum   float64
	Count int64
}

func (m *Measurement) Merge(other *Measurement) {
	m.Min = min(m.Min, other.Min)
	m.Max = max(m.Max, other.Max)
	m.Sum = math.Round((m.Sum+other.Sum)*100) / 100 // round to 2 decimals
	m.Count = m.Count + other.Count
}
