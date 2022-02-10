package cart

type MBC0 struct {
	ROM []uint8
	RAM [0x2000]uint8
}

func NewMBC0(rom []uint8) MBC {
	return &MBC0{ROM: rom}
}

func (m *MBC0) readByte(addr uint16) uint8 {
	if addr <= 0x7FFF {
		return m.ROM[addr]
	} else if (addr >= 0xA000) && (addr <= 0xBFFF) {
		return m.RAM[addr-0xA000]
	}
	return 0x00
}

func (m *MBC0) writeROM(addr uint16, val uint8) {
}

func (m *MBC0) writeRAM(addr uint16, val uint8) {
	m.RAM[addr-0xA000] = val
}

func (m *MBC0) getRomBank() uint32 {
	return 0
}
