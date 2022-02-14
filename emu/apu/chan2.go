package apu

import (
	"math"

	"github.com/is386/GoBoy/emu/bits"
)

var (
	DUTY_TABLE = map[byte]float64{
		0: -0.25,
		1: -0.5,
		2: 0,
		3: 0.5,
	}
)

type Channel2 struct {
	freq         float64
	freqLowBits  uint8
	freqHighBits uint8

	timer     float64
	amplitude float64
	duration  int
	len       uint8

	duty uint8

	envVol       uint8
	envTime      uint8
	envSteps     uint8
	envStepsInit uint8
	envSweep     uint8
	envSamples   uint8
	envDir       uint8

	initial        uint8
	durationSelect uint8

	leftOn  uint8
	rightOn uint8
	left    uint16
	right   uint16
}

func NewChannel2() *Channel2 {
	return &Channel2{}
}

func (c *Channel2) update() {
	var sample uint16
	step := c.freq * 2 * math.Pi / float64(SAMPLE_RATE)
	c.timer += step

	if (c.duration == -1 || c.duration > 0) && c.envStepsInit > 0 {
		sample = uint16(square(c.timer, DUTY_TABLE[c.duty]))
		if c.duration > 0 {
			c.duration--
		}
	}

	c.envelope()
	if c.leftOn == 1 {
		c.left = sample
	}
	if c.rightOn == 1 {
		c.right = sample
	}
}

func (c *Channel2) writeByte(reg uint8, val uint8) {
	switch reg {

	case NR21:
		c.duty = val >> 6
		c.len = val & 0x3F

	case NR22:
		c.envVol = val >> 4
		c.envDir = bits.Value(val, 3)
		c.envSweep = val & 0x7
		c.envSamples = c.envSweep * uint8(SAMPLE_RATE/64)

	case NR23:
		c.freqLowBits = val
		freq := (uint16(c.freqHighBits) << 8) | uint16(c.freqLowBits)
		c.freq = 131072 / (2048 - float64(freq))

	case NR24:
		c.freqHighBits = val & 0x7
		freq := (uint16(c.freqHighBits) << 8) | uint16(c.freqLowBits)
		c.freq = 131072 / (2048 - float64(freq))

		c.initial = bits.Value(val, 7)
		c.durationSelect = bits.Value(val, 6)

		if c.initial != 0 {
			if c.len == 0 {
				c.len = 64
			}
			c.duration = -1
			if c.durationSelect != 0 {
				c.duration = int(float64(c.len)*float64(1)/64) * SAMPLE_RATE
			}
			c.amplitude = 1
			c.envSteps = c.envVol
			c.envStepsInit = c.envVol
		}
	}
}

func (c *Channel2) readByte(reg uint8) uint8 {
	switch reg {

	case NR21:
		return (c.duty << 6) | c.len

	case NR22:
		return (c.envVol << 4) | (c.envDir << 3) | c.envSweep

	case NR23:
		return c.freqLowBits

	case NR24:
		return (c.initial << 7) | (c.durationSelect << 6) | c.freqHighBits
	}
	return 0x00
}

func (c *Channel2) envelope() {
	if c.envSamples > 0 {
		c.envTime += 1
		if c.envSteps > 0 && c.envTime >= c.envSamples {
			c.envTime -= c.envSamples
			c.envSteps--
			if c.envSteps == 0 {
				c.amplitude = 0
			} else if c.envDir == 1 {
				c.amplitude = 1 - float64(c.envSteps)/float64(c.envStepsInit)
			} else {
				c.amplitude = float64(c.envSteps) / float64(c.envStepsInit)
			}
		}
	}
}

func square(t float64, bound float64) float64 {
	if math.Sin(t) <= bound {
		return 255.0
	}
	return 0
}
