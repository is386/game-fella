package cart

import (
	"fmt"
	"os"
)

type Cartridge struct {
	mbc MBC
}

func NewCartridge(rom []uint8) *Cartridge {
	cart := &Cartridge{}
	mbcType := rom[0x147]
	switch mbcType {
	case 0:
		cart.mbc = NewMBC0()
	case 1, 2, 3:
		cart.mbc = NewMBC1(rom, mbcType)
	default:
		fmt.Printf("Unknown MBC Type: %d\n", mbcType)
		os.Exit(0)
	}
	return cart
}

func (c *Cartridge) ReadByte(addr uint16) uint8 {
	return c.mbc.readByte(addr)
}

func (c *Cartridge) WriteROM(addr uint16, val uint8) {
	c.mbc.writeROM(addr, val)
}

func (c *Cartridge) WriteRAM(addr uint16, val uint8) {
	c.mbc.writeRAM(addr, val)
}
