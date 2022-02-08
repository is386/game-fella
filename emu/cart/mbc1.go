package cart

type MBC1 struct {
	ROM        [0x7FFF - 0x0000 + 1]uint8
	RAM        [0xBFFF - 0xA000 + 1]uint8
	romBank    int
	ramBank    int
	mode       uint8
	ramEnabled bool
}

func NewMBC1(rom []uint8, mode uint8) MBC {
	mbc := &MBC1{mode: mode, romBank: 1}
	return mbc
}

func (m *MBC1) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {
	case 0x0000, 0x1000, 0x2000, 0x3000:
		return m.ROM[addr]
	case 0x8000:
		return m.ROM[uint32(m.romBank*0x4000)+uint32(addr-0x4000)]
	}
	return m.ROM[addr]
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

	}
	m.ROM[addr] = val
}

func (m *MBC1) writeRAM(addr uint16, val uint8) {
	if m.ramEnabled {
		m.RAM[uint32(m.ramBank*0x2000)+uint32(addr)] = val
	}
}
