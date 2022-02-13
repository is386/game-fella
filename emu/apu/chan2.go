package apu

import (
	"math"
)

type Channel2 struct {
	self  *Channel2
	freq  float64
	ticks float64
}

func NewChannel2(freq float64) Channel2 {
	return Channel2{freq: freq}
}

func square(t float64, bound float64) float64 {
	if math.Sin(t) <= bound {
		return 1
	}
	return 0
}

func (c Channel2) Stream(samples [][2]float64) (int, bool) {
	for i := range samples {
		c.self.ticks += c.freq / SAMPLE_RATE
		samples[i][0] = square((c.self.ticks * 2 * math.Pi), 0.5)
		samples[i][1] = square((c.self.ticks * 2 * math.Pi), 0.5)
	}
	return len(samples), true
}

func (c Channel2) Err() error {
	return nil
}
