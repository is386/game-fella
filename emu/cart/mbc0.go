package cart

type MBC0 struct {
	ROM [0x7FFF - 0x0000 + 1]uint8
}

func NewMBC0() MBC {
	return &MBC0{}
}

func (m *MBC0) readByte(addr uint16) uint8 {
	return m.ROM[addr]
}

func (m *MBC0) writeROM(addr uint16, val uint8) {
	m.ROM[addr] = val
}

func (m *MBC0) writeRAM(addr uint16, val uint8) {}
