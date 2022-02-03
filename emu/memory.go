package emu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	DIV  uint16 = 0xFF04
	TIMA uint16 = 0xFF05
	TMA  uint16 = 0xFF06
	TAC  uint16 = 0xFF07
)

type Memory struct {
	gb   *GameBoy
	rom  [0x7FFF - 0x0000 + 1]uint8
	vram [0x9FFF - 0x8000 + 1]uint8
	eram [0xBFFF - 0xA000 + 1]uint8
	wram [0xDFFF - 0xC000 + 1]uint8
	oam  [0xFE9F - 0xFE00 + 1]uint8
	hram [0xFFFF - 0xFF00 + 1]uint8
}

func NewMemory(gb *GameBoy) *Memory {
	mem := Memory{gb: gb}
	return &mem
}

func (m *Memory) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {
	case 0x0000: // BIOS
		fallthrough
	case 0x1000: // ROM 0
		fallthrough
	case 0x2000:
		fallthrough
	case 0x3000:
		fallthrough
	case 0x4000: // ROM n
		fallthrough
	case 0x5000:
		fallthrough
	case 0x6000:
		fallthrough
	case 0x7000:
		return m.rom[addr]

	case 0x8000: // VRAM
		fallthrough
	case 0x9000:
		return m.vram[addr-0x8000]

	case 0xA000: // ERAM
		fallthrough
	case 0xB000:
		return m.eram[addr-0xA000]

	case 0xC000: // WRAM
		fallthrough
	case 0xD000:
		return m.wram[addr-0xC000]

	case 0xF000: // OAM and HRAM
		switch addr & 0x0F00 {
		case 0x0E00:
			if addr < 0xFEA0 { // OAM
				return m.oam[addr-0xFE00]
			}

		case 0x0F00: // HRAM
			return m.hram[addr-0xFF00]
		}
	}
	return 0
}

func (m *Memory) writeByte(addr uint16, val uint8) {
	switch addr & 0xF000 {
	case 0x0000: // BIOS
		fallthrough
	case 0x1000: // ROM 0
		fallthrough
	case 0x2000:
		fallthrough
	case 0x3000:
		fallthrough
	case 0x4000: // ROM n
		fallthrough
	case 0x5000:
		fallthrough
	case 0x6000:
		fallthrough
	case 0x7000:
		m.rom[addr] = val

	case 0x8000: // VRAM
		fallthrough
	case 0x9000:
		m.vram[addr-0x8000] = val

	case 0xA000: // ERAM
		fallthrough
	case 0xB000:
		m.eram[addr-0xA000] = val

	case 0xC000: // WRAM
		fallthrough
	case 0xD000:
		fallthrough
	case 0xE000: // WRAM Echo
		m.wram[addr-0xC000] = val

	case 0xF000: // OAM and HRAM
		switch addr & 0x0F00 {
		case 0x0E00:
			if addr < 0xFEA0 { // OAM
				m.oam[addr-0xFE00] = val
			}

		case 0x0F00: // HRAM
			m.hram[addr-0xFF00] = val
			if addr == 0xFF02 && val == 0x81 {
				m.gb.printSerialLink()
			}
		}
	}
}

// func (m *Memory) writeHRam(addr uint16, val uint8) {
// 	switch {
// 	case addr == 0xFF02:
// 		if val == 0x81 {
// 			m.gb.printSerialLink()
// 		}

// 	case addr == DIV:
// 		m.gb.timer.resetTimer()
// 		m.gb.timer.resetDivCyc()
// 		m.hram[DIV-0xFF00] = 0

// 	case addr == TIMA:
// 		m.hram[TIMA-0xFF00] = val

// 	case addr == TMA:
// 		m.hram[TMA-0xFF00] = val

// 	case addr == TAC:
// 		freq0 := m.gb.timer.getTimerFreq()
// 		m.hram[TAC-0xFF00] = val | 0xF8
// 		freq := m.gb.timer.getTimerFreq()
// 		if freq0 != freq {
// 			m.gb.timer.resetTimer()
// 		}

// 	case addr == 0xFF41:
// 		m.hram[0x41] = val | 0x80

// 	case addr == 0xFF44:
// 		m.hram[0x44] = 0

// 	case addr == 0xFF46:
// 		m.doDMATransfer(val)

// 	default:
// 		m.hram[addr-0xFF00] = val
// 	}
// }

func (m *Memory) doDMATransfer(val uint8) {
	addr := uint16(val) << 8
	for i := uint16(0); i < 0xA0; i++ {
		m.writeByte(0xFE00+i, m.readByte(addr+i))
	}
}

func (m *Memory) doHDMATransfer() {
}

func (m *Memory) incrDiv() {
	m.hram[DIV-0xFF00]++
}

func (m *Memory) isTimerEnabled() bool {
	return bits.Test(m.hram[0x07], 2)
}

func (m *Memory) getTimerFreq() uint8 {
	return m.hram[0x07] & 0x3
}

func (m *Memory) updateTima() {
	tima := m.hram[0x05]
	if tima == 0xFF {
		m.hram[TIMA-0xFF00] = m.hram[0x06]
		m.writeInterrupt(2)
	} else {
		m.hram[TIMA-0xFF00] = tima + 1
	}
}

func (m *Memory) writeInterrupt(i int) {
	req := m.hram[0x0F] | 0xE0
	req = bits.Set(req, uint8(i))
	m.writeByte(0xFF0F, req)
}

func (m *Memory) getLCDStatus() uint8 {
	return m.hram[0x41]
}

func (m *Memory) isLCDEnabled() bool {
	return bits.Test(m.hram[0x40], 7)
}
