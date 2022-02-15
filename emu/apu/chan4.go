package apu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	DIVISORS = map[uint8]int{
		0x0: 8,
		0x1: 16,
		0x2: 32,
		0x3: 48,
		0x4: 64,
		0x5: 80,
		0x6: 96,
		0x7: 112,
	}
)

type Channel4 struct {
	freqTimer int

	envVol    uint8
	envPeriod uint8
	envDir    uint8
	envTimer  int
	currVol   int

	lfsr         uint16
	shiftAmount  uint8
	counterWidth uint8
	divisorCode  uint8

	nr43Upper     uint8
	triggerBit    uint8
	length        uint8
	lengthTimer   int
	lengthEnabled uint8

	leftOn  uint8
	rightOn uint8
	left    int
	right   int
	enabled bool
}

func NewChannel4() *Channel4 {
	return &Channel4{}
}

func (c *Channel4) update() {
	var sample int
	c.freqTimer--
	if c.freqTimer <= 0 {
		c.freqTimer = int(DIVISORS[c.divisorCode] << c.shiftAmount)
		xorResult := (c.lfsr & 1) ^ ((c.lfsr & 2) >> 1)
		c.lfsr = (c.lfsr >> 1) | (xorResult << 14)

		if c.counterWidth == 1 {
			temp := int32(c.lfsr)
			temp &= ^0x40
			c.lfsr = uint16(temp)
			c.lfsr |= xorResult << 6
		}

		if c.enabled && (c.lfsr&1) == 0 {
			sample = c.currVol
		} else {
			sample = 0
		}

		if c.leftOn == 1 {
			c.left = sample
		}
		if c.rightOn == 1 {
			c.right = sample
		}
	}

}

func (c *Channel4) writeByte(reg uint8, val uint8) {
	switch reg {

	case NR41:
		c.length = val
		c.lengthTimer = int(64 - (c.length & 0x3F))

	case NR42:
		c.envVol = val >> 4
		c.envDir = bits.Value(val, 3)
		c.envPeriod = val & 0x7

	case NR43:
		c.shiftAmount = val >> 4
		c.counterWidth = bits.Value(val, 3)
		c.divisorCode = val & 0x7

	case NR44:
		c.nr43Upper = val & 0x7
		c.triggerBit = bits.Value(val, 7)
		c.lengthEnabled = bits.Value(val, 6)
		if c.triggerBit == 1 {
			c.trigger()
		}
	}
}

func (c *Channel4) readByte(reg uint8) uint8 {
	switch reg {

	case NR41:
		return c.length

	case NR42:
		return (c.envVol << 4) | (c.envDir << 3) | c.envPeriod

	case NR43:
		return (c.shiftAmount << 4) | (c.counterWidth << 3) | c.divisorCode

	case NR44:
		return (c.triggerBit << 7) | (c.lengthEnabled << 6) | c.nr43Upper
	}
	return 0x00
}

func (c *Channel4) clockEnvelope() {
	if c.envPeriod != 0 {
		c.envTimer--

		if c.envTimer <= 0 {
			c.envTimer = int(c.envPeriod)

			if c.currVol < 0xF && c.envDir == 1 {
				c.currVol++
			} else if c.currVol > 0x0 && c.envDir == 0 {
				c.currVol--
			}
		}
	}
}

func (c *Channel4) clockLength() {
	if c.lengthEnabled == 1 && c.lengthTimer > 0 {
		c.lengthTimer--
		if c.lengthTimer == 0 {
			c.enabled = false
		}
	}
}

func (c *Channel4) trigger() {
	c.envTimer = int(c.envPeriod)
	c.currVol = int(c.envVol)
	c.enabled = true
	c.lfsr = 0x7FFF
	if c.lengthTimer == 0 {
		c.lengthTimer = 64
	}
}
