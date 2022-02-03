package emu

var (
	MEM_SIZE        = 65536
	DIV      uint16 = 0xFF04
	TIMA     uint16 = 0xFF05
	TMA      uint16 = 0xFF06
	TAC      uint16 = 0xFF07
)

type Memory struct {
	gb                 *GameBoy
	rom                [0x8000]uint8
	vram               [0x4000]uint8
	cram               [0x2000]uint8
	wram               [0x9000]uint8
	oam                [0x100]uint8
	hram               [0x100]uint8
	vramBank, wramBank uint8
}

func NewMemory(gb *GameBoy) *Memory {
	mem := Memory{gb: gb}
	mem.hram[0x44] = 0x90
	return &mem
}

func (m *Memory) readByte(addr uint16) uint8 {
	switch {
	case addr < 0x8000:
		return m.rom[addr]

	case addr < 0xA000:
		bank := uint16(m.vramBank) * 0x2000
		return m.vram[addr-0x8000+bank]

	case addr < 0xC000:
		return m.cram[addr]

	case addr < 0xD000:
		return m.wram[addr-0xC000]

	case addr < 0xE000:
		bank := (uint16(m.wramBank) * 0x1000)
		return m.wram[addr-0xC000+bank]

	case addr < 0xFE00:
		return 0

	case addr < 0xFEA0:
		return m.oam[addr-0xFE00]

	case addr < 0xFF00:
		return 0

	default:
		return m.readHram(addr)
	}
}

func (m *Memory) readHram(addr uint16) uint8 {
	switch {
	case addr == 0xFF00:
		// TODO: Joypad
		return 0

	case addr >= 0xFF10 && addr <= 0xFF26:
		// TODO: Sound
		return 0

	case addr >= 0xFF30 && addr <= 0xFF3F:
		// TODO: channel 3 waveform RAM.
		return 0

	case addr == 0xFF0F:
		return m.hram[0x0F] | 0xE0

	case addr >= 0xFF72 && addr <= 0xFF77:
		return 0

	case addr == 0xFF68:
		// TODO: BG palette index
		return 0

	case addr == 0xFF69:
		// TODO: BG Palette data
		return 0

	case addr == 0xFF6A:
		// TODO: Sprite palette index
		return 0

	case addr == 0xFF6B:
		// TODO: Sprite Palette data
		return 0

	case addr == 0xFF4D:
		// TODO: Speed switch data
		return 0

	case addr == 0xFF4F:
		return m.vramBank

	case addr == 0xFF70:
		return m.wramBank

	default:
		return m.hram[addr-0xFF00]
	}
}

func (m *Memory) writeByte(addr uint16, val uint8) {
	switch {
	case addr < 0x8000:
		m.rom[addr] = val

	case addr < 0xA000:
		bank := uint16(m.vramBank) * 0x2000
		m.vram[addr-0x8000+bank] = val

	case addr < 0xC000:
		m.cram[addr] = val

	case addr < 0xD000:
		m.wram[addr-0xC000] = val

	case addr < 0xE000:
		bank := uint16(m.wramBank) * 0x1000
		m.wram[addr-0xC000+bank] = val

	case addr < 0xFE00:
		return

	case addr < 0xFEA0:
		m.oam[addr-0xFE00] = val

	case addr < 0xFF00:
		return

	default:
		m.writeHRam(addr, val)
	}
}

func (m *Memory) writeHRam(addr uint16, val uint8) {
	switch {
	case addr >= 0xFEA0 && addr < 0xFEFF:
		return

	case addr >= 0xFF10 && addr <= 0xFF26:
		// TODO: Sound
		return

	case addr >= 0xFF30 && addr <= 0xFF3F:
		// TODO: Writing to channel 3 waveform RAM
		return

	case addr == 0xFF02:
		if val == 0x81 {
			m.gb.printSerialLink()
		}

	case addr == DIV:
		m.gb.timer.resetTimer()
		m.gb.timer.resetDivCyc()
		m.hram[DIV-0xFF00] = 0

	case addr == TIMA:
		m.hram[TIMA-0xFF00] = val

	case addr == TMA:
		m.hram[TMA-0xFF00] = val

	case addr == TAC:
		freq0 := m.gb.timer.getTimerFreq()
		m.hram[TAC-0xFF00] = val | 0xF8
		freq := m.gb.timer.getTimerFreq()
		if freq0 != freq {
			m.gb.timer.resetTimer()
		}

	case addr == 0xFF41:
		m.hram[0x41] = val | 0x80

	case addr == 0xFF44:
		m.hram[0x44] = 0

	case addr == 0xFF46:
		// TODO: DMA transfer
		return

	case addr == 0xFF4D:
		// TODO: CGB speed change
		return

	case addr == 0xFF4F:
		// TODO: VRAM bank (CGB only), blocked when HDMA is active
		return

	case addr == 0xFF55:
		// TODO: CGB DMA transfer
		return

	case addr == 0xFF68:
		// TODO: BG palette index
		return

	case addr == 0xFF69:
		// TODO: BG Palette data
		return

	case addr == 0xFF6A:
		// TODO: Sprite palette index
		return

	case addr == 0xFF6B:
		// TODO: Sprite Palette data
		return

	case addr == 0xFF70:
		// TODO: WRAM1 bank (CGB mode)
		return

	case addr >= 0xFF72 && addr <= 0xFF77:
		return

	default:
		m.hram[addr-0xFF00] = val
	}
}

func (m *Memory) incrDiv() {
	m.hram[DIV-0xFF00]++
}

func (m *Memory) isTimerEnabled() bool {
	return ((m.hram[0x07] >> 2) & 1) == 1
}

func (m *Memory) getTimerFreq() uint8 {
	return m.hram[0x07] & 0x3
}

func (m *Memory) updateTima() {
	tima := m.hram[0x05]
	if tima == 0xFF {
		m.hram[TIMA-0xFF00] = m.hram[0x06]
		req := m.hram[0x0F] | 0xE0
		req = req | (1 << 2)
		m.writeByte(0xFF0F, req)
	} else {
		m.hram[TIMA-0xFF00] = tima + 1
	}
}

func (m *Memory) getLCDMode() uint8 {
	return m.hram[0x41]
}

func (m *Memory) isLCDEnabled() bool {
	return ((m.hram[0x40] >> 7) & 1) == 1
}
