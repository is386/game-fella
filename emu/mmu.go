package emu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	JOYPAD uint8 = 0x00
	COMM1  uint8 = 0x01
	COMM2  uint8 = 0x02
	DIV    uint8 = 0x04
	TIMA   uint8 = 0x05
	TMA    uint8 = 0x06
	TAC    uint8 = 0x07
	LCDC   uint8 = 0x40
	STAT   uint8 = 0x41
	SCY    uint8 = 0x42
	SCX    uint8 = 0x43
	LY     uint8 = 0x44
	LYC    uint8 = 0x45
	DMA    uint8 = 0x46
	BGP    uint8 = 0x47
	OBP0   uint8 = 0x48
	OBP1   uint8 = 0x49
	WY     uint8 = 0x4A
	WX     uint8 = 0x4B
	VBK    uint8 = 0x4F
	KEY1   uint8 = 0x4D
	HDMA1  uint8 = 0x51
	HDMA2  uint8 = 0x52
	HDMA3  uint8 = 0x53
	HDMA4  uint8 = 0x54
	HDMA5  uint8 = 0x55
	RP     uint8 = 0x56
	OPRI   uint8 = 0x6C
	SVBK   uint8 = 0x70
)

type MMU struct {
	gb          *GameBoy
	bootROM     []uint8
	VRAM        [0x4000]uint8
	WRAM        [0x8000]uint8
	OAM         [0x00A0]uint8
	HRAM        [0x0100]uint8
	bootEnabled bool
	startup     bool
}

func NewMMU(gb *GameBoy) *MMU {
	m := MMU{gb: gb, startup: true}
	return &m
}

func (m *MMU) initHRAM() {
	m.writeByte(0xFF00, 0xCF)
	m.writeByte(0xFF01, 0x00)
	m.writeByte(0xFF02, 0x7E)
	m.writeByte(0xFF04, 0xAB)
	m.writeByte(0xFF05, 0x00)
	m.writeByte(0xFF06, 0x00)
	m.writeByte(0xFF07, 0xF8)
	m.writeByte(0xFF0F, 0xE1)
	m.writeByte(0xFF10, 0x80)
	m.writeByte(0xFF11, 0xBF)
	m.writeByte(0xFF12, 0xF3)
	m.writeByte(0xFF13, 0xFF)
	m.writeByte(0xFF14, 0xBF)
	m.writeByte(0xFF16, 0x3F)
	m.writeByte(0xFF17, 0x00)
	m.writeByte(0xFF18, 0xFF)
	m.writeByte(0xFF19, 0xBF)
	m.writeByte(0xFF1A, 0x7F)
	m.writeByte(0xFF1B, 0xFF)
	m.writeByte(0xFF1C, 0x9F)
	m.writeByte(0xFF1D, 0xFF)
	m.writeByte(0xFF1E, 0xBF)
	m.writeByte(0xFF20, 0xFF)
	m.writeByte(0xFF21, 0x00)
	m.writeByte(0xFF22, 0x00)
	m.writeByte(0xFF23, 0xBF)
	m.writeByte(0xFF24, 0x77)
	m.writeByte(0xFF25, 0xF3)
	m.writeByte(0xFF26, 0xF1)
	m.writeByte(0xFF40, 0x91)
	m.writeByte(0xFF41, 0x85)
	m.writeByte(0xFF42, 0x00)
	m.writeByte(0xFF43, 0x00)
	m.writeByte(0xFF44, 0x00)
	m.writeByte(0xFF45, 0x00)
	m.writeByte(0xFF46, 0xFF)
	m.writeByte(0xFF47, 0xFC)
	m.writeByte(0xFF4A, 0x00)
	m.writeByte(0xFF4B, 0x00)
	m.writeByte(0xFF4D, 0xFF)
	m.writeByte(0xFF4F, 0xFF)
	m.writeByte(0xFF51, 0xFF)
	m.writeByte(0xFF52, 0xFF)
	m.writeByte(0xFF53, 0xFF)
	m.writeByte(0xFF54, 0xFF)
	m.writeByte(0xFF55, 0xFF)
	m.writeByte(0xFF56, 0xFF)
	m.writeByte(0xFF68, 0xFF)
	m.writeByte(0xFF69, 0xFF)
	m.writeByte(0xFF6A, 0xFF)
	m.writeByte(0xFF6B, 0xFF)
	m.writeByte(0xFF70, 0xFF)
	m.writeByte(0xFFFF, 0x00)
}

func (m *MMU) loadBootRom(rom []uint8) {
	m.bootROM = rom
	m.bootEnabled = true
}

func (m *MMU) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {
	case 0x0000:
		if m.bootEnabled && addr < 0x100 {
			return m.bootROM[addr]
		} else if m.bootEnabled && m.gb.cpu.pc == 0x100 {
			m.bootEnabled = false
		}
		return m.gb.cart.ReadByte(addr)

	case 0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000:
		return m.gb.cart.ReadByte(addr)

	case 0x8000, 0x9000:
		var bank uint8 = 0
		if m.gb.isCGB {
			bank = bits.Value(m.HRAM[VBK], 0)
		}
		return m.VRAM[((addr - 0x8000) + (uint16(bank) * 0x2000))]

	case 0xA000, 0xB000:
		return m.gb.cart.ReadByte(addr)

	case 0xC000:
		return m.WRAM[addr-0xC000]

	case 0xD000:
		var bank uint8 = 0
		if m.gb.isCGB {
			bank = m.HRAM[SVBK] & 0x3
		}
		return m.WRAM[((addr - 0xC000) + (uint16(bank) * 0x1000))]

	case 0xE000:
		return 0xFF
	}

	switch addr & 0x0F00 {
	case 0x0E00:
		if addr-0xFE00 < 160 {
			return m.OAM[addr-0xFE00]
		}

	case 0x0F00:
		if addr == 0xFF00 {
			return m.gb.buttons.readByte(addr)
		}
		if addr >= 0xFF10 && addr <= 0xFF26 {
			m.gb.apu.ReadByte(addr)
		} else if addr >= 0xFF30 && addr <= 0xFF3F {
			m.gb.apu.ReadByte(addr)
		}
		return m.readHRAM(uint8(addr - 0xFF00))
	}

	return 0xFF
}

func (m *MMU) writeByte(addr uint16, val uint8) {
	switch addr & 0xF000 {

	case 0x0000, 0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000:
		if !m.startup {
			m.gb.cart.WriteROM(addr, val)
		}
		return

	case 0x8000, 0x9000:
		var bank uint8 = 0
		if m.gb.isCGB {
			bank = bits.Value(m.HRAM[VBK], 0)
		}
		m.VRAM[((addr - 0x8000) + (uint16(bank) * 0x2000))] = val
		return

	case 0xA000, 0xB000:
		m.gb.cart.WriteRAM(addr, val)
		return

	case 0xC000:
		m.WRAM[addr-0xC000] = val
		return

	case 0xD000:
		var bank uint8 = 0
		if m.gb.isCGB {
			bank = m.HRAM[SVBK] & 0x3
		}
		m.WRAM[((addr - 0xC000) + (uint16(bank) * 0x1000))] = val
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
		if addr >= 0xFF10 && addr <= 0xFF26 {
			m.gb.apu.WriteByte(addr, val)
		} else if addr >= 0xFF30 && addr <= 0xFF3F {
			m.gb.apu.WriteByte(addr, val)
		}
		m.writeHRAM(uint8(addr-0xFF00), val)
	}
}

func (m *MMU) writeHRAM(reg uint8, val uint8) {
	switch reg {
	case JOYPAD:
		m.gb.buttons.writeByte(0xFF00, val)

	case COMM2:
		if val == 0x81 {
			m.gb.printSerialLink()
		}

	case DIV:
		m.gb.timer.resetTimer()
		m.gb.timer.resetDivCyc()
		m.HRAM[DIV] = 0

	case TAC:
		freq0 := m.gb.timer.getTimerFreq()
		m.HRAM[TAC] = val | 0xF8
		freq := m.gb.timer.getTimerFreq()
		if freq0 != freq {
			m.gb.timer.resetTimer()
		}

	case DMA:
		m.dmaTransfer(val)

	case HDMA5:
		if m.gb.isCGB {
			// HDMA Transfer
		}

	case KEY1:
		if m.gb.isCGB {
			// Prepare Speed Switch
		}

	case RP:
		if m.gb.isCGB {
			// Infrared Port
		}

	default:
		m.HRAM[reg] = val
	}
}

func (m *MMU) readHRAM(reg uint8) uint8 {
	switch {

	case reg == VBK && m.gb.isCGB:
		return m.HRAM[VBK] & 0xFE

	default:
		return m.HRAM[reg]
	}
}

func (m *MMU) dmaTransfer(val uint8) {
	addr := uint16(val) << 8
	for i := uint16(0); i < 0xA0; i++ {
		m.writeByte(0xFE00+i, m.readByte(addr+i))
	}
}

func (m *MMU) writeInterrupt(i int) {
	req := m.HRAM[0x0F] | 0xE0
	req = bits.Set(req, uint8(i))
	m.writeByte(0xFF0F, req)
}

func (m *MMU) incrDiv() {
	m.HRAM[DIV]++
}

func (m *MMU) incrLY() {
	m.HRAM[LY]++
}

func (m *MMU) incrTima() {
	m.HRAM[TIMA]++
}
