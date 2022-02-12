package emu

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Buttons struct {
	gb     *GameBoy
	rows   [2]uint8
	column uint8
}

func NewButtons(gb *GameBoy) *Buttons {
	b := &Buttons{gb: gb}
	b.rows[0] = 0x0F
	b.rows[1] = 0x0F
	return b
}

func (b *Buttons) readByte(addr uint16) uint8 {
	if addr == 0xFF00 {
		if b.column == 0x10 {
			return b.rows[0] | 0xC0
		} else if b.column == 0x20 {
			return b.rows[1] | 0xC0
		} else {
			return 0xCF
		}
	}
	return 0x00
}

func (b *Buttons) writeByte(addr uint16, val uint8) {
	if addr == 0xFF00 {
		b.column = val & 0x30
	}
}

func (b *Buttons) keyDown(key sdl.Keycode) {
	bHit := false
	dHit := false

	switch key {
	case sdl.K_RETURN: // Start
		b.rows[0] &= 0x7
		bHit = true
	case sdl.K_RSHIFT: // Select
		b.rows[0] &= 0xB
		bHit = true
	case sdl.K_w: // Up
		b.rows[1] &= 0xB
		dHit = true
	case sdl.K_s: // Down
		b.rows[1] &= 0x7
		dHit = true
	case sdl.K_a: // Left
		b.rows[1] &= 0xD
		dHit = true
	case sdl.K_d: // Right
		b.rows[1] &= 0xE
		dHit = true
	case sdl.K_j: // A
		b.rows[0] &= 0xE
		bHit = true
	case sdl.K_k: // B
		b.rows[0] &= 0xD
		bHit = true
	case sdl.K_ESCAPE:
		b.gb.close()
	}

	if (dHit && b.column == 0x10) || (bHit && b.column == 0x20) {
		b.gb.mmu.writeInterrupt(INT_JOYPAD)
	}
}

func (b *Buttons) keyUp(key sdl.Keycode) {
	switch key {
	case sdl.K_RETURN: // Start
		b.rows[0] |= 0x8
	case sdl.K_RSHIFT: // Select
		b.rows[0] |= 0x4
	case sdl.K_w: // Up
		b.rows[1] |= 0x4
	case sdl.K_s: // Down
		b.rows[1] |= 0x8
	case sdl.K_a: // Left
		b.rows[1] |= 0x2
	case sdl.K_d: // Right
		b.rows[1] |= 0x1
	case sdl.K_j: // A
		b.rows[0] |= 0x1
	case sdl.K_k: // B
		b.rows[0] |= 0x2
	case sdl.K_ESCAPE:
		b.gb.close()
	}
}
