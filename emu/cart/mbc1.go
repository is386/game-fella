package cart

import (
	"github.com/is386/GoBoy/emu/bits"
)

type MBC1 struct {
	ROM            []uint8
	RAM            [0xBFFF - 0xA000 + 1]uint8
	romBank        int
	ramBank        int
	maxRomBanks    int
	maxRamBanks    int
	mode           uint8
	ramEnabled     bool
	bankingEnabled bool
}

func NewMBC1(rom []uint8, mode uint8, romBanks int, ramBanks int) MBC {
	mbc := &MBC1{
		ROM:         rom,
		mode:        mode,
		romBank:     1,
		maxRomBanks: romBanks,
		maxRamBanks: ramBanks}
	return mbc
}

func (m *MBC1) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {
	case 0x0000, 0x1000, 0x2000, 0x3000:
		return m.ROM[addr]
	case 0x4000, 0x5000, 0x6000, 0x7000:
		return m.ROM[uint32(m.romBank*0x4000)+uint32(addr-0x4000)]
	default:
		return m.RAM[uint32(m.ramBank*0x2000)+uint32(addr-0xA000)]
	}
}

func (m *MBC1) writeROM(addr uint16, val uint8) {
	switch addr & 0xF000 {

	case 0x0000, 0x1000:
		if val == 0x0 {
			m.ramEnabled = false
		} else if val == 0xA {
			m.ramEnabled = true
		}

	case 0x2000, 0x3000:
		if m.maxRomBanks != 0 {
			m.romBank = int((val & 0x1F) % uint8(m.maxRomBanks))
		} else {
			m.romBank = 1
		}

	case 0x4000, 0x5000:
		if !m.bankingEnabled {
			if m.maxRamBanks != 0 {
				m.ramBank = int((val & 0x3) % uint8(m.maxRamBanks))
			} else {
				m.ramBank = 0
			}
		} else {
			b1 := int(val & 1)
			b2 := int((val >> 1) & 1)
			m.romBank = int(m.romBank | (b1 << 5))
			m.romBank = int(m.romBank | (b2 << 6))
		}

	case 0x6000, 0x7000:
		if bits.Test(val, 0) {
			m.romBank = int(bits.Reset(uint8(m.romBank), 5))
			m.romBank = int(bits.Reset(uint8(m.romBank), 6))
			m.romBank = int(bits.Reset(uint8(m.romBank), 7))
			m.romBank = int(bits.Reset(uint8(m.romBank), 8))
			m.romBank = int((val & 0x1F))
			m.bankingEnabled = true
		} else {
			m.bankingEnabled = false
			m.ramBank = 0
		}
	}
}

func (m *MBC1) writeRAM(addr uint16, val uint8) {
	if m.ramEnabled {
		m.RAM[uint32(m.ramBank*0x2000)+uint32(addr)] = val
	}
}
