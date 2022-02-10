package cart

import (
	"fmt"
	"os"
)

var (
	RAM_SIZES = map[uint8]uint32{
		0x0: 0 * 1024,
		0x1: 2 * 1024,
		0x2: 8 * 1024,
		0x3: 32 * 1024,
		0x4: 128 * 1024,
		0x5: 64 * 1024,
	}
)

type Cartridge struct {
	mbc MBC
}

func NewCartridge(rom []uint8) *Cartridge {
	cart := &Cartridge{}
	mbcType := rom[0x147]
	romSize := uint32((32 * 1024) << rom[0x148])
	ramSize := RAM_SIZES[rom[0x149]]

	switch mbcType {
	case 0:
		cart.mbc = NewMBC0()
	case 1, 2, 3:
		cart.mbc = NewMBC1(rom, romSize, ramSize, mbcType)
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
