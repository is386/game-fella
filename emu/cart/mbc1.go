package cart

import (
	"github.com/is386/GoBoy/emu/bits"
)

type MBC1 struct {
	ROM            []uint8
	RAM            [0xBFFF - 0xA000 + 1]uint8
	romBank        uint32
	ramBank        uint32
	maxRomBanks    uint32
	maxRamBanks    uint32
	ramEnabled     bool
	bankingEnabled bool
}

func NewMBC1(rom []uint8, romBanks uint32, ramBanks uint32) MBC {
	mbc := &MBC1{
		ROM:         rom,
		romBank:     1,
		maxRomBanks: romBanks - 1,
		maxRamBanks: ramBanks}
	return mbc
}

func (m *MBC1) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {

	case 0x0000, 0x1000, 0x2000, 0x3000:
		return m.ROM[addr]

	case 0x4000, 0x5000, 0x6000, 0x7000:
		return m.ROM[uint32((m.romBank&0x1F)*0x4000)+uint32(addr-0x4000)]

	case 0xA000, 0xB000:
		if m.ramEnabled {
			return m.RAM[uint32(m.ramBank*0x2000)+uint32(addr-0xA000)]
		}
	}
	return 0xFF
}

func (m *MBC1) writeROM(addr uint16, val uint8) {
	switch addr & 0xF000 {

	case 0x0000, 0x1000:
		if (val & 0x0F) == 0x0A {
			m.ramEnabled = true
		} else {
			m.ramEnabled = false
		}

	case 0x2000, 0x3000:
		if m.maxRomBanks != 0 {
			m.romBank = (uint32(val) & 0x1F) % m.maxRomBanks
		} else {
			m.romBank = 0
		}

	case 0x4000, 0x5000:
		if !m.bankingEnabled {
			if m.maxRamBanks != 0 {
				m.ramBank = (uint32(val) & 0x3) % m.maxRamBanks
			} else {
				m.ramBank = 0
			}
		} else {
			b1 := uint32(val) & 1
			b2 := (uint32(val) >> 1) & 1
			m.romBank = m.romBank | (b1 << 5)
			m.romBank = m.romBank | (b2 << 6)
		}

	case 0x6000, 0x7000:
		if bits.Test(val, 0) {
			m.romBank = uint32(bits.Reset(uint8(m.romBank), 5))
			m.romBank = uint32(bits.Reset(uint8(m.romBank), 6))
			m.romBank = uint32(bits.Reset(uint8(m.romBank), 7))
			m.romBank = uint32(bits.Reset(uint8(m.romBank), 8))
			if m.maxRomBanks != 0 {
				m.romBank = (uint32(val) & 0x1F) % m.maxRomBanks
			} else {
				m.romBank = 0
			}
			m.bankingEnabled = true
		} else {
			m.bankingEnabled = false
			m.ramBank = 0
		}
	}
}

func (m *MBC1) writeRAM(addr uint16, val uint8) {
	if m.ramEnabled {
		m.RAM[uint32(m.ramBank*0x2000)+uint32(addr-0xA000)] = val
	}
}
