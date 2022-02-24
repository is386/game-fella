package emu

import "github.com/is386/GoBoy/emu/bits"

type CRAM struct {
	CRAM     [0x40]uint8
	autoIncr bool
	index    uint8
}

func (c *CRAM) writeIndex(val uint8) {
	if bits.Test(val, 7) {
		c.autoIncr = true
	} else {
		c.autoIncr = false
	}
	c.index = val & 0x3F
}

func (c *CRAM) writeCRAM(val uint8) {
	c.CRAM[c.index] = val
	if c.autoIncr {
		c.index += 1
	}
	c.index %= 0x40
}

func (c *CRAM) readCurrentCRAM() uint8 {
	return c.CRAM[c.index]
}

func (c *CRAM) readCRAM(addr uint8) uint8 {
	return c.CRAM[addr]
}
