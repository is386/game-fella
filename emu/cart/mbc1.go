package cart

type MBC1 struct {
	ROM         []uint8
	RAM         [0x2000]uint8
	romType     uint8
	romBank     uint32
	upperBank   uint32
	romOffset   uint32
	romBankMask uint32
	ramBank     uint32
	ramOffset   uint32
	ramBankMask uint32
	mode        bool
	ramEnabled  bool
}

func NewMBC1(rom []uint8, romSize uint32, ramSize uint32, romType uint8) MBC {
	mbc := &MBC1{
		ROM:         rom,
		romType:     romType,
		romBank:     1,
		romBankMask: (romSize / 0x4000) - 1,
		ramBankMask: (ramSize / 0x2000) - 1}
	return mbc
}

func (m *MBC1) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {

	case 0x0000, 0x1000, 0x2000, 0x3000:
		if m.mode {
			offset := ((m.upperBank << 5) & m.romBankMask) * 0x4000
			return m.ROM[offset+uint32(addr)]
		} else {
			return m.ROM[addr]
		}

	case 0x4000, 0x5000, 0x6000, 0x7000:
		return m.ROM[m.romOffset+uint32(addr-0x4000)]

	case 0xA000, 0xB000:
		if m.ramEnabled {
			return m.RAM[(m.ramOffset + uint32(addr-0xA000))]
		} else {
			return 0xFF
		}
	}
	return 0x00
}

func (m *MBC1) writeROM(addr uint16, val uint8) {
	switch addr & 0xF000 {

	case 0x0000, 0x1000:
		if m.romType == 2 || m.romType == 3 {
			m.ramEnabled = (val & 0x0F) == 0x0A
		}

	case 0x2000, 0x3000:
		val &= 0x1F
		if val == 0 {
			val = 1
		}
		m.romBank = uint32(val) & m.romBankMask
		m.updateBanks()

	case 0x4000, 0x5000:
		m.upperBank = uint32(val) & 0x3
		m.updateBanks()

	case 0x6000, 0x7000:
		m.mode = (val & 1) != 0
		m.updateBanks()
	}
}

func (m *MBC1) writeRAM(addr uint16, val uint8) {
	if m.ramEnabled {
		m.RAM[m.ramOffset+uint32(addr-0xA000)] = val
	}
}

func (m *MBC1) updateBanks() {
	if m.mode {
		m.ramBank = m.upperBank & m.ramBankMask
		m.ramOffset = m.ramBank * 0x2000
	} else {
		m.ramOffset = 0
	}
	m.romOffset = ((m.romBank | (m.upperBank << 5)) & m.romBankMask) * 0x4000
}
