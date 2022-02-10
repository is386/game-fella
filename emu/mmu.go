package emu

import (
	"fmt"
	"io/ioutil"

	"github.com/is386/GoBoy/emu/bits"
	"github.com/is386/GoBoy/emu/cart"
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
)

type MMU struct {
	gb          *GameBoy
	cart        *cart.Cartridge
	bootROM     [0x00FF - 0x0000 + 1]uint8
	VRAM        [0x9FFF - 0x8000 + 1]uint8
	WRAM        [0xDFFF - 0xC000 + 1]uint8
	OAM         [0xFE9F - 0xFE00 + 1]uint8
	HRAM        [0xFFFF - 0xFF00 + 1]uint8
	bootEnabled bool
	startup     bool
}

func NewMMU(gb *GameBoy, bootEnabled bool) *MMU {
	mmu := MMU{gb: gb, bootEnabled: bootEnabled, startup: true}
	if mmu.bootEnabled {
		mmu.loadBootRom()
	}
	return &mmu
}

func (m *MMU) loadBootRom() {
	boot, err := ioutil.ReadFile("roms/boot.bin")
	if err != nil {
		fmt.Println("roms/boot.bin not found. Skipping boot screen...")
		return
	}
	for i := 0; i < len(boot); i++ {
		m.bootROM[i] = boot[i]
	}
	m.gb.cpu.pc = 0x00
}

func (m *MMU) readByte(addr uint16) uint8 {
	switch addr & 0xF000 {
	case 0x0000:
		if m.bootEnabled && addr < 0x100 {
			return m.bootROM[addr]
		} else if m.bootEnabled && m.gb.cpu.pc == 0x100 {
			m.bootEnabled = false
		}
		return m.cart.ReadByte(addr)

	case 0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000:
		return m.cart.ReadByte(addr)

	case 0x8000, 0x9000:
		return m.VRAM[addr-0x8000]

	case 0xA000, 0xB000:
		return m.cart.ReadByte(addr)

	case 0xC000, 0xD000, 0xE000:
		if addr >= 0xE000 {
			addr -= 0x2000
		}
		return m.WRAM[addr-0xC000]
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
		return m.HRAM[addr-0xFF00]
	}

	return 0xFF
}

func (m *MMU) writeByte(addr uint16, val uint8) {
	switch addr & 0xF000 {

	case 0x0000, 0x1000, 0x2000, 0x3000, 0x4000, 0x5000, 0x6000, 0x7000:
		if !m.startup {
			m.cart.WriteROM(addr, val)
		}
		return

	case 0x8000, 0x9000:
		m.VRAM[addr-0x8000] = val
		return

	case 0xA000, 0xB000:
		m.cart.WriteRAM(addr, val)
		return

	case 0xC000, 0xD000, 0xE000:
		if addr >= 0xE000 {
			addr -= 0x2000
		}
		m.WRAM[addr-0xC000] = val
		return
	}

	switch addr & 0x0F00 {
	case 0x0E00:
		if addr < 0xFEA0 {
			m.OAM[addr-0xFE00] = val
		}

	case 0x0F00:
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
		m.dmaTransfer(val)

	default:
		m.HRAM[reg] = val
	}
}

func (m *MMU) readHRAM(reg uint8) uint8 {
	return m.HRAM[reg]
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
