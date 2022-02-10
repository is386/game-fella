package cart

type MBC1 struct {
	ROM               []uint8
	RAM               [0x2000]uint8
	romBank           uint32
	romBankUpperBits  uint32
	ramBank           uint32
	totalRomBanks     uint32
	totalRamBanks     uint32
	ramEnabled        bool
	advBankingEnabled bool
}

func NewMBC1(rom []uint8, romBanks uint32, ramBanks uint32) MBC {
	mbc := &MBC1{
		ROM:           rom,
		romBank:       1,
		totalRomBanks: romBanks,
		totalRamBanks: ramBanks}
	return mbc
}

func (m *MBC1) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {

	case 0x0000, 0x1000, 0x2000, 0x3000:
		return m.ROM[addr]

	case 0x4000, 0x5000, 0x6000, 0x7000:
		return m.ROM[(uint32(m.romBank*0x4000) + uint32(addr-0x4000))]

	case 0xA000, 0xB000:
		if m.ramEnabled {
			return m.RAM[(uint32(m.ramBank*0x2000) + uint32(addr-0xA000))]
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
		val &= 0x1F
		if val == 0 {
			val = 1
		}
		m.romBank = ((m.romBankUpperBits << 5) | uint32(val&0x1F)) % m.totalRomBanks

	case 0x4000, 0x5000:
		if m.advBankingEnabled {
			m.romBankUpperBits = uint32(val & 0x3)
		} else if m.totalRamBanks > 0 {
			m.ramBank = (uint32(val & 0x3)) % m.totalRamBanks
		}

	case 0x6000, 0x7000:
		m.advBankingEnabled = (val & 1) == 0
		if m.advBankingEnabled {
			m.ramBank = 0
		} else {
			m.romBank &= 0x1F
			m.romBankUpperBits = 0
		}
	}
}

func (m *MBC1) writeRAM(addr uint16, val uint8) {
	if m.ramEnabled {
		m.RAM[(uint32(m.ramBank*0x2000) + uint32(addr-0xA000))] = val
	}
}

func (m *MBC1) getRomBank() uint32 {
	return m.romBank
}
