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
	BGPI   uint8 = 0x68
	BGPD   uint8 = 0x69
	OBPI   uint8 = 0x6A
	OBPD   uint8 = 0x6B
	OPRI   uint8 = 0x6C
	SVBK   uint8 = 0x70
)

type MMU struct {
	gb          *GameBoy
	bootROM     []uint8
	VRAM        [0x4000]uint8
	WRAM        [0x9000]uint8
	OAM         [0x00A0]uint8
	HRAM        [0x0100]uint8
	vramBank    uint8
	wramBank    uint8
	bgCRAM      CRAM
	spriteCRAM  CRAM
	bootEnabled bool
}

func NewMMU(gb *GameBoy) *MMU {
	m := MMU{gb: gb, wramBank: 1}
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
		if m.bootEnabled && (addr < 0x100 || (addr >= 0x200 && addr < 0x900)) {
			return m.bootROM[addr]
		} else if m.bootEnabled && m.gb.cpu.pc == 0x100 {
			m.bootEnabled = false
		}
		return m.gb.cart.ReadByte(addr)

	case 0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000:
		return m.gb.cart.ReadByte(addr)

	case 0x8000, 0x9000:
		return m.VRAM[((addr - 0x8000) + (uint16(m.vramBank) * 0x2000))]

	case 0xA000, 0xB000:
		return m.gb.cart.ReadByte(addr)

	case 0xC000, 0xD000, 0xE000:
		if addr >= 0xE000 {
			addr -= 0x2000
		}
		if addr <= 0xCFFF {
			return m.WRAM[addr-0xC000]
		} else {
			return m.WRAM[((addr - 0xC000) + (uint16(m.wramBank) * 0x1000))]
		}
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
		m.gb.cart.WriteROM(addr, val)
		return

	case 0x8000, 0x9000:
		m.VRAM[((addr - 0x8000) + (uint16(m.vramBank) * 0x2000))] = val
		return

	case 0xA000, 0xB000:
		m.gb.cart.WriteRAM(addr, val)
		return

	case 0xC000, 0xD000, 0xE000:
		if addr >= 0xE000 {
			addr -= 0x2000
		}
		if addr <= 0xCFFF {
			m.WRAM[addr-0xC000] = val
		} else {
			m.WRAM[((addr - 0xC000) + (uint16(m.wramBank) * 0x1000))] = val
		}
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

	case VBK:
		if m.gb.isCGB {
			m.vramBank = val & 1
		}

	case SVBK:
		if m.gb.isCGB {
			m.wramBank = val & 0x7
			if m.wramBank == 0 {
				m.wramBank = 1
			}
		}

	case HDMA5:
		if m.gb.isCGB {
			m.newDMATransfer(val)
		}

	case KEY1:
		if m.gb.isCGB {
			// Prepare Speed Switch
		}

	case BGPI:
		if m.gb.isCGB {
			m.bgCRAM.writeIndex(val)
		}

	case BGPD:
		if m.gb.isCGB {
			m.bgCRAM.writeCRAM(val)
		}

	case OBPI:
		if m.gb.isCGB {
			m.spriteCRAM.writeIndex(val)
		}

	case OBPD:
		if m.gb.isCGB {
			m.spriteCRAM.writeCRAM(val)
		}
	}
	m.HRAM[reg] = val
}

func (m *MMU) readBgCRAM(addr uint8) uint8 {
	return m.bgCRAM.readCRAM(addr)
}

func (m *MMU) readSpriteCRAM(addr uint8) uint8 {
	return m.spriteCRAM.readCRAM(addr)
}

func (m *MMU) readVRAM(addr uint16) uint8 {
	return m.VRAM[addr]
}

func (m *MMU) readHRAM(reg uint8) uint8 {
	switch {

	case reg == KEY1 && m.gb.isCGB:
		return 0

	case reg == BGPI:
		return m.bgCRAM.readAddr()

	case reg == BGPD:
		return m.bgCRAM.readCurrentCRAM()

	case reg == OBPI:
		return m.spriteCRAM.readAddr()

	case reg == OBPD:
		return m.spriteCRAM.readCurrentCRAM()

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

func (m *MMU) newDMATransfer(val uint8) {
	mode := bits.Value(val, 7)

	if mode == 0 {
		m.gdmaTransfer(val)
	} else {
		// HDMA
	}
}

func (m *MMU) gdmaTransfer(val uint8) {
	hdma1 := m.HRAM[HDMA1]
	hdma2 := m.HRAM[HDMA2] & 0xF0
	src := (uint16(hdma1) << 8) | uint16(hdma2)

	hdma3 := m.HRAM[HDMA3] & 0x1F
	hdma4 := m.HRAM[HDMA4] & 0xF0
	dst := (uint16(hdma3) << 8) | uint16(hdma4)
	dst += 0x8000

	len := uint16(((val & 0x7F) + 1) * 0x10)

	for i := uint16(0); i < len; i++ {
		m.writeByte(dst+i, m.readByte(src+i))
	}

	m.HRAM[HDMA1] = uint8((src & 0xFF00) >> 8)
	m.HRAM[HDMA2] = uint8((src & 0x00FF) & 0xF0)
	m.HRAM[HDMA3] = uint8(((dst & 0xFF00) >> 8) & 0x1F)
	m.HRAM[HDMA4] = uint8((dst & 0x00FF) & 0xF0)
	m.HRAM[HDMA5] = 0xFF
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
