package apu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	OUTPUT_LEVELS = map[uint8]uint8{
		0x0: 4,
		0x1: 0,
		0x2: 1,
		0x3: 2,
	}
)

type Channel3 struct {
	waveRAM      [16]uint8
	wavePosition int

	freqLowBits  uint8
	freqHighBits uint8
	freqTimer    int

	outputLevelByte uint8
	outputLevel     uint8
	triggerBit      uint8
	enableByte      uint8

	length        uint8
	lengthTimer   int
	lengthEnabled uint8

	leftOn  uint8
	rightOn uint8
	left    int
	right   int
	enabled bool
}

func NewChannel3() *Channel3 {
	return &Channel3{}
}

func (c *Channel3) update() {
	var sample int
	c.freqTimer--
	if c.freqTimer <= 0 {
		freq := int((uint16(c.freqHighBits) << 8) | uint16(c.freqLowBits))
		c.freqTimer = (2048 - freq) * 2
		c.wavePosition++
		c.wavePosition &= 31

		if c.enabled {
			ramPos := c.wavePosition / 2
			waveOut := c.waveRAM[ramPos]

			if (c.wavePosition % 2) == 0 { // if the wavePosition is even we want the first sample (upper bits)
				waveOut = (waveOut & 0xF0) >> 4
			} else {
				waveOut &= 0x0F // second sample (lower bits)
			}

			waveOut >>= OUTPUT_LEVELS[c.outputLevel] // shift right depending on the output level vals
			sample = int(waveOut)
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

func (c *Channel3) writeByte(reg uint8, val uint8) {
	switch reg {

	case NR30:
		c.enableByte = val
		c.enabled = bits.Test(val, 7)

	case NR31:
		c.length = val
		c.lengthTimer = 256 - int(c.length)

	case NR32:
		c.outputLevelByte = val
		c.outputLevel = (val >> 5) & 0x3

	case NR33:
		c.freqLowBits = val

	case NR34:
		c.freqHighBits = val & 0x7
		c.triggerBit = bits.Value(val, 7)
		c.lengthEnabled = bits.Value(val, 6)
		if c.triggerBit == 1 {
			c.trigger()
		}
	}

	if reg >= 0x30 && reg <= 0x3F {
		c.waveRAM[reg&0x0F] = val
	}
}

func (c *Channel3) readByte(reg uint8) uint8 {
	switch reg {
	case NR30:
		return c.enableByte

	case NR31:
		return c.length

	case NR32:
		return c.outputLevelByte

	case NR33:
		return c.freqLowBits

	case NR34:
		return (c.triggerBit << 7) | (c.lengthEnabled << 6) | c.freqHighBits
	}

	if reg >= 0x30 && reg <= 0x3F {
		return c.waveRAM[reg&0x0F]
	}

	return 0x00
}

func (c *Channel3) clockLength() {
	if c.lengthEnabled == 1 && c.lengthTimer > 0 {
		c.lengthTimer--
		if c.lengthTimer == 0 {
			c.enabled = false
		}
	}
}

func (c *Channel3) trigger() {
	c.enabled = true
	if c.lengthTimer == 0 {
		c.lengthTimer = 256
	}
	c.wavePosition = 0
}
