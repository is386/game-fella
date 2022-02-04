package emu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	DIV  uint8 = 0x04
	TIMA uint8 = 0x05
	TMA  uint8 = 0x06
	TAC  uint8 = 0x07
	LCDC uint8 = 0x40
	STAT uint8 = 0x41
	SCY  uint8 = 0x42
	SCX  uint8 = 0x43
	LY   uint8 = 0x44
	LYC  uint8 = 0x45
	DMA  uint8 = 0x46
	BGP  uint8 = 0x47
	OBP0 uint8 = 0x48
	OBP1 uint8 = 0x49
	WY   uint8 = 0x4A
	WX   uint8 = 0x4B
)

type Memory struct {
	gb   *GameBoy
	ROM  [0x7FFF - 0x0000 + 1]uint8
	VRAM [0x9FFF - 0x8000 + 1]uint8
	ERAM [0xBFFF - 0xA000 + 1]uint8
	WRAM [0xDFFF - 0xC000 + 1]uint8
	OAM  [0xFE9F - 0xFE00 + 1]uint8
	HRAM [0xFFFF - 0xFF00 + 1]uint8
}

func NewMemory(gb *GameBoy) *Memory {
	mem := Memory{gb: gb}
	return &mem
}

func (m *Memory) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {
	case 0x0000, 0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000:
		return m.ROM[addr]
	case 0x8000, 0x9000:
		return m.VRAM[addr-0x8000]
	case 0xA000, 0xB000:
		return m.ERAM[addr-0xA000]
	case 0xC000, 0xD000:
		return m.WRAM[addr-0xC000]
	case 0xE000:
		return 0
	}

	switch addr & 0x0F00 {
	case 0x0E00:
		if addr < 0xFEA0 {
			return m.OAM[addr-0xFE00]
		}
	case 0x0F00:
		return m.HRAM[addr-0xFF00]
	}
	return 0
}

func (m *Memory) writeByte(addr uint16, val uint8) {
	switch addr & 0xF000 {
	case 0x0000, 0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000:
		m.ROM[addr] = val
		return
	case 0x8000, 0x9000:
		m.VRAM[addr-0x8000] = val
		return
	case 0xA000, 0xB000:
		m.ERAM[addr-0xA000] = val
		return
	case 0xC000, 0xD000:
		m.WRAM[addr-0xC000] = val
		return
	case 0xE000:
		return
	}

	switch addr & 0x0F00 {
	case 0x0E00:
		if addr < 0xFEA0 {
			m.OAM[addr-0xFE00] = val
		}
	case 0x0F00:
		m.HRAM[addr-0xFF00] = val
		if addr == 0xFF02 && val == 0x81 {
			m.gb.printSerialLink()
		}
	}
}

func (m *Memory) writeHRAM(reg uint8, val uint8) {
	switch reg {
	case DIV:
		m.gb.timer.resetTimer()
		m.gb.timer.resetDivCyc()
		m.HRAM[DIV] = 0
	case TIMA:
		m.HRAM[TIMA] = val
	case TMA:
		m.HRAM[TMA] = val
	case TAC:
		freq0 := m.gb.timer.getTimerFreq()
		m.HRAM[TAC] = val | 0xF8
		freq := m.gb.timer.getTimerFreq()
		if freq0 != freq {
			m.gb.timer.resetTimer()
		}
	case STAT:
		m.HRAM[STAT] = val | 0x80
	case LY:
		m.HRAM[LY] = 0
	case DMA:
		m.doDMATransfer(val)
	}
}

func (m *Memory) readHRAM(reg uint8) uint8 {
	return m.HRAM[reg]
}

func (m *Memory) doDMATransfer(val uint8) {
	addr := uint16(val) << 8
	for i := uint16(0); i < 0xA0; i++ {
		m.writeByte(0xFE00+i, m.readByte(addr+i))
	}
}

func (m *Memory) doHDMATransfer() {
}

func (m *Memory) writeInterrupt(i int) {
	req := m.HRAM[0x0F] | 0xE0
	req = bits.Set(req, uint8(i))
	m.writeByte(0xFF0F, req)
}

func (m *Memory) incrDiv() {
	m.HRAM[DIV]++
}

func (m *Memory) incrLY() {
	m.HRAM[LY]++
}

func (m *Memory) isTimerEnabled() bool {
	return bits.Test(m.HRAM[TAC], 2)
}

func (m *Memory) getTimerFreq() uint8 {
	return m.HRAM[TAC] & 0x3
}

func (m *Memory) updateTima() {
	tima := m.HRAM[TIMA]
	if tima == 0xFF {
		m.HRAM[TIMA] = m.HRAM[TMA]
		m.writeInterrupt(2)
	} else {
		m.HRAM[TIMA] = tima + 1
	}
}
