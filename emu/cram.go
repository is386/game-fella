package emu

import "github.com/is386/GoBoy/emu/bits"

type CRAM struct {
	CRAM     [0x0040]uint8
	autoIncr uint8
	cramAddr uint8
}

func (c *CRAM) writeIndex(val uint8) {
	c.autoIncr = bits.Value(val, 7)
	c.cramAddr = val & 0x31
}

func (c *CRAM) writeCRAM(val uint8) {
	c.CRAM[c.cramAddr] = val
	c.cramAddr += c.autoIncr
	c.cramAddr &= 0x3F
}

func (c *CRAM) readCRAM(addr uint8) uint8 {
	return c.CRAM[addr]
}
