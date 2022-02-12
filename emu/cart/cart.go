package cart

import (
	"fmt"
	"math"
	"os"
)

var (
	RAM_BANKS = map[uint8]int{
		0x0: 0,
		0x1: 1,
		0x2: 1,
		0x3: 4,
		0x4: 16,
		0x5: 8,
	}
)

type Cartridge struct {
	mbc MBC
}

func NewCartridge(rom []uint8) *Cartridge {
	cart := &Cartridge{}
	mbcType := rom[0x147]
	romBanks := int(math.Pow(2, float64(rom[0x148])+1))
	ramBanks := RAM_BANKS[rom[0x149]]

	switch mbcType {

	case 0x00:
		cart.mbc = NewMBC0(rom)

	case 0x01, 0x02, 0x03:
		cart.mbc = NewMBC1(rom, uint32(romBanks), uint32(ramBanks))

	case 0x0F, 0x10, 0x11, 0x12, 0x13:
		cart.mbc = NewMBC3(rom, uint32(romBanks), uint32(ramBanks))

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

func (c *Cartridge) GetRomBank() uint32 {
	return c.mbc.getRomBank()
}
