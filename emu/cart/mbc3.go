package cart

type MBC3 struct {
	ROM           []uint8
	RAM           []uint8
	rtc           []uint8
	latchedRtc    []uint8
	isLatched     bool
	romBank       uint32
	ramBank       uint32
	totalRomBanks uint32
	totalRamBanks uint32
	ramEnabled    bool
}

func NewMBC3(rom []uint8, romBanks uint32, ramBanks uint32) MBC {
	mbc := &MBC3{
		ROM:           rom,
		RAM:           make([]uint8, (ramBanks+1)*0x2000),
		romBank:       1,
		totalRomBanks: romBanks,
		totalRamBanks: ramBanks,
		rtc:           make([]uint8, 0x10),
		latchedRtc:    make([]uint8, 0x10)}
	return mbc
}

func (m *MBC3) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {

	case 0x0000, 0x1000, 0x2000, 0x3000:
		return m.ROM[addr]

	case 0x4000, 0x5000, 0x6000, 0x7000:
		return m.ROM[(uint32(m.romBank*0x4000) + uint32(addr-0x4000))]

	case 0xA000, 0xB000:
		if m.ramEnabled && m.ramBank < 0x04 {
			return m.RAM[(uint32(m.ramBank*0x2000) + uint32(addr-0xA000))]
		} else {
			if m.isLatched {
				return m.latchedRtc[m.ramBank]
			}
			return m.rtc[m.ramBank]
		}
	}
	return 0xFF
}

func (m *MBC3) writeROM(addr uint16, val uint8) {
	switch addr & 0xF000 {

	case 0x0000, 0x1000:
		if (val & 0x0F) == 0x0A {
			m.ramEnabled = true
		} else {
			m.ramEnabled = false
		}

	case 0x2000, 0x3000:
		m.romBank = uint32(val) % m.totalRomBanks

	case 0x4000, 0x5000:
		m.ramBank = uint32(val)

	case 0x6000, 0x7000:
		if val == 0x01 {
			m.isLatched = false
		} else if val == 0x00 {
			m.isLatched = true
			copy(m.rtc, m.latchedRtc)
		}
	}
}

func (m *MBC3) writeRAM(addr uint16, val uint8) {
	if m.ramEnabled && addr >= 0xA000 && addr <= 0xBFFF {
		if m.ramBank < 0x04 {
			m.RAM[(uint32(m.ramBank*0x2000) + uint32(addr-0xA000))] = val
		} else {
			m.rtc[m.ramBank] = val
		}
	}
}

func (m *MBC3) getRomBank() uint32 {
	return m.romBank
}

func (m *MBC3) loadData(data []uint8) {
	m.RAM = data
}

func (m *MBC3) saveData() []uint8 {
	return m.RAM
}
