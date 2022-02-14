package apu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	DUTY_TABLE = [4][8]uint8{
		{0, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 1, 1, 1},
		{0, 1, 1, 1, 1, 1, 1, 0},
	}
)

type Channel2 struct {
	freqLowBits  uint8
	freqHighBits uint8
	freqTimer    float64

	duty         uint8
	dutyPosition uint8

	envVol      uint8
	period      uint8
	envDir      uint8
	periodTimer uint8
	currVol     uint8

	triggerBit    uint8
	length        uint8
	lengthTimer   float64
	lengthEnabled uint8

	leftOn  uint8
	rightOn uint8
	left    uint16
	right   uint16
	enabled bool
}

func NewChannel2() *Channel2 {
	return &Channel2{}
}

func (c *Channel2) update() {
	var sample uint16
	c.freqTimer--
	if c.freqTimer <= 0 {
		freq := float64((uint16(c.freqHighBits) << 8) | uint16(c.freqLowBits))
		c.freqTimer = (2048 - freq) * 4
		c.dutyPosition += 1
		c.dutyPosition &= 7
	}
	if c.enabled {
		sample = uint16(c.currVol * DUTY_TABLE[c.duty][c.dutyPosition])
		sample = uint16((float64(sample) / 7.5) - 1.0)
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

func (c *Channel2) writeByte(reg uint8, val uint8) {
	switch reg {

	case NR21:
		c.duty = val >> 6
		c.length = val & 0x3F
		c.lengthTimer = float64((64 - c.length)) * (1.0 / 256.0)

	case NR22:
		c.envVol = val >> 4
		c.envDir = bits.Value(val, 3)
		c.period = val & 0x7

	case NR23:
		c.freqLowBits = val

	case NR24:
		c.freqHighBits = val & 0x7
		c.triggerBit = bits.Value(val, 7)
		c.lengthEnabled = bits.Value(val, 6)
		if c.triggerBit == 1 {
			c.trigger()
		}
	}
}

func (c *Channel2) readByte(reg uint8) uint8 {
	switch reg {

	case NR21:
		return (c.duty << 6) | uint8(c.length)

	case NR22:
		return (c.envVol << 4) | (c.envDir << 3) | c.period

	case NR23:
		return c.freqLowBits

	case NR24:
		return (c.triggerBit << 7) | (c.lengthEnabled << 6) | c.freqHighBits
	}
	return 0x00
}

func (c *Channel2) clockEnvelope() {
	if c.period != 0 {
		if c.periodTimer > 0 {
			c.periodTimer--
		}
		if c.periodTimer <= 0 {
			c.periodTimer = c.period
			if ((c.currVol < 0xF) && (c.envDir == 1)) || ((c.currVol > 0x0) && (c.envDir == 0)) {
				if c.envDir == 1 {
					c.currVol++
				} else {
					c.currVol--
				}
			}
		}
	}
}

func (c *Channel2) clockLength() {
	if c.lengthEnabled == 1 && c.lengthTimer > 0 {
		c.lengthTimer--
		if c.lengthTimer == 0 {
			c.enabled = false
		}
	}
}

func (c *Channel2) trigger() {
	c.periodTimer = c.period
	c.currVol = c.envVol
	c.enabled = true
	if c.lengthTimer == 0 {
		c.lengthTimer = 64
	}
}
