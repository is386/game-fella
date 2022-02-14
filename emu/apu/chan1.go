package apu

import (
	"github.com/is386/GoBoy/emu/bits"
)

type Channel1 struct {
	freqLowBits  uint16
	freqHighBits uint16
	freqTimer    int

	duty         uint8
	dutyPosition uint8

	envVol    uint8
	envPeriod uint8
	envDir    uint8
	envTimer  int
	currVol   int

	sweepPeriod  uint8
	sweepDir     uint8
	sweepShift   uint8
	sweepTimer   int
	sweepEnabled bool
	shadowFreq   int

	triggerBit    uint8
	length        uint8
	lengthTimer   int
	lengthEnabled uint8

	leftOn  uint8
	rightOn uint8
	left    uint16
	right   uint16
	enabled bool
}

func NewChannel1() *Channel1 {
	return &Channel1{}
}

func (c *Channel1) update() {
	var sample uint16
	c.freqTimer--
	if c.freqTimer <= 0 {
		freq := int((uint16(c.freqHighBits) << 8) | uint16(c.freqLowBits))
		c.freqTimer = (2048 - freq) * 4
		c.dutyPosition += 1
		c.dutyPosition &= 7
	}
	if c.enabled {
		sample = uint16(c.currVol * DUTY_TABLE[c.duty][c.dutyPosition])
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

func (c *Channel1) writeByte(reg uint8, val uint8) {
	switch reg {

	case NR10:
		c.sweepPeriod = val >> 4
		c.sweepDir = bits.Value(val, 3)
		c.sweepShift = val & 0x7

	case NR11:
		c.duty = val >> 6
		c.length = val & 0x3F
		c.lengthTimer = int(64 - c.length)

	case NR12:
		c.envVol = val >> 4
		c.envDir = bits.Value(val, 3)
		c.envPeriod = val & 0x7

	case NR13:
		c.freqLowBits = uint16(val)

	case NR14:
		c.freqHighBits = uint16(val) & 0x7
		c.triggerBit = bits.Value(val, 7)
		c.lengthEnabled = bits.Value(val, 6)
		if c.triggerBit == 1 {
			c.trigger()
		}
	}
}

func (c *Channel1) readByte(reg uint8) uint8 {
	switch reg {

	case NR10:
		return (c.sweepPeriod << 4) | (c.sweepDir << 3) | c.sweepShift

	case NR11:
		return (c.duty << 6) | uint8(c.length)

	case NR12:
		return (c.envVol << 4) | (c.envDir << 3) | c.envPeriod

	case NR13:
		return uint8(c.freqLowBits)

	case NR14:
		return (c.triggerBit << 7) | (c.lengthEnabled << 6) | uint8(c.freqHighBits)
	}
	return 0x00
}

func (c *Channel1) clockEnvelope() {
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

func (c *Channel1) clockLength() {
	if c.lengthEnabled == 1 && c.lengthTimer > 0 {
		c.lengthTimer--
		if c.lengthTimer == 0 {
			c.enabled = false
		}
	}
}

func (c *Channel1) clockSweep() {
	c.sweepTimer--
	if c.sweepTimer <= 0 {
		if c.sweepPeriod > 0 {
			c.sweepTimer = int(c.sweepPeriod)
		} else {
			c.sweepTimer = 8
		}

		if c.sweepEnabled && c.sweepPeriod > 0 {
			newFreq := c.calculateFreq()
			if newFreq <= 2047 && c.sweepShift > 0 {
				c.freqLowBits = uint16(newFreq & 0xFF)
				c.freqHighBits = uint16((newFreq >> 8) & 7)
				c.shadowFreq = int(newFreq)
				c.calculateFreq()
			}
		}
	}
}

func (c *Channel1) calculateFreq() int {
	newFreq := c.shadowFreq >> c.sweepShift
	if c.sweepDir == 1 {
		newFreq = c.shadowFreq - newFreq
	} else {
		newFreq = c.shadowFreq + newFreq
	}

	if newFreq > 2047 {
		c.enabled = false
	}
	return newFreq
}

func (c *Channel1) trigger() {
	c.envTimer = int(c.envPeriod)
	c.currVol = int(c.envVol)
	c.enabled = true
	if c.lengthTimer == 0 {
		c.lengthTimer = 64
	}

	c.sweepEnabled = true
	c.shadowFreq = int((uint16(c.freqHighBits) << 8) | uint16(c.freqLowBits))
	c.sweepTimer = int(c.sweepPeriod)
	if c.sweepTimer == 0 {
		c.sweepTimer = 8
	}
	c.sweepEnabled = c.sweepPeriod > 0 || c.sweepShift > 0
	if c.sweepShift > 0 {
		c.calculateFreq()
	}
}
