package cart

type MBC5 struct {
	ROM           []uint8
	RAM           []uint8
	romBank       uint32
	ramBank       uint32
	totalRomBanks uint32
	totalRamBanks uint32
	ramEnabled    bool
}

func NewMBC5(rom []uint8, romBanks uint32, ramBanks uint32) MBC {
	mbc := &MBC5{
		ROM:           rom,
		RAM:           make([]uint8, (ramBanks+1)*0x2000),
		romBank:       1,
		totalRomBanks: romBanks,
		totalRamBanks: ramBanks}
	return mbc
}

func (m *MBC5) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {

	case 0x0000, 0x1000, 0x2000, 0x3000:
		return m.ROM[addr]

	case 0x4000, 0x5000, 0x6000, 0x7000:
		m.romBank %= m.totalRomBanks
		return m.ROM[(uint32(m.romBank*0x4000) + uint32(addr-0x4000))]

	case 0xA000, 0xB000:
		return m.RAM[(uint32(m.ramBank*0x2000) + uint32(addr-0xA000))]
	}
	return 0xFF
}

func (m *MBC5) writeROM(addr uint16, val uint8) {
	switch addr & 0xF000 {

	case 0x0000, 0x1000:
		if (val & 0x0F) == 0x0A {
			m.ramEnabled = true
		} else {
			m.ramEnabled = false
		}

	case 0x2000:
		m.romBank = (m.romBank & 0x100) | uint32(val)

	case 0x3000:
		m.romBank = (m.romBank & 0xFF) | uint32(val&0x01)<<8

	case 0x4000, 0x5000:
		m.ramBank = uint32(val)
	}
}

func (m *MBC5) writeRAM(addr uint16, val uint8) {
	if m.ramEnabled && addr >= 0xA000 && addr <= 0xBFFF {
		m.RAM[(uint32(m.ramBank*0x2000) + uint32(addr-0xA000))] = val
	}
}

func (m *MBC5) getRomBank() uint32 {
	return m.romBank
}

func (m *MBC5) loadData(data []uint8) {
	m.RAM = data
}

func (m *MBC5) saveData() []uint8 {
	return m.RAM
}
