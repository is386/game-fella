package cart

type MBC0 struct {
	ROM []uint8
}

func NewMBC0(rom []uint8) MBC {
	return &MBC0{ROM: rom}
}

func (m *MBC0) readByte(addr uint16) uint8 {
	return m.ROM[addr]
}

func (m *MBC0) writeROM(addr uint16, val uint8) {}

func (m *MBC0) writeRAM(addr uint16, val uint8) {}
