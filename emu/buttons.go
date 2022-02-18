package emu

import (
	"github.com/faiface/pixel/pixelgl"
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

func (b *Buttons) update() {
	b.keyDown()
	b.keyUp()
}

func (b *Buttons) keyDown() {
	bHit := false
	dHit := false

	if b.gb.screen.win.Pressed(pixelgl.KeyEnter) { // Start
		b.rows[0] &= 0x7
		bHit = true
	}

	if b.gb.screen.win.Pressed(pixelgl.KeyRightShift) { // Select
		b.rows[0] &= 0xB
		bHit = true
	}

	if b.gb.screen.win.Pressed(pixelgl.KeyW) { // Up
		b.rows[1] &= 0xB
		dHit = true
	}

	if b.gb.screen.win.Pressed(pixelgl.KeyS) { // Down
		b.rows[1] &= 0x7
		dHit = true
	}

	if b.gb.screen.win.Pressed(pixelgl.KeyA) { // Left
		b.rows[1] &= 0xD
		dHit = true
	}

	if b.gb.screen.win.Pressed(pixelgl.KeyD) { // Right
		b.rows[1] &= 0xE
		dHit = true
	}
	if b.gb.screen.win.Pressed(pixelgl.KeyJ) { // A
		b.rows[0] &= 0xE
		bHit = true
	}

	if b.gb.screen.win.Pressed(pixelgl.KeyK) { // B
		b.rows[0] &= 0xD
		bHit = true
	}

	if b.gb.screen.win.Pressed(pixelgl.KeyEscape) {
		b.gb.close()
	}

	if (dHit && b.column == 0x10) || (bHit && b.column == 0x20) {
		b.gb.mmu.writeInterrupt(INT_JOYPAD)
	}
}

func (b *Buttons) keyUp() {
	if b.gb.screen.win.JustReleased(pixelgl.KeyEnter) { // Start
		b.rows[0] |= 0x8
	}

	if b.gb.screen.win.JustReleased(pixelgl.KeyRightShift) { // Select
		b.rows[0] |= 0x4
	}

	if b.gb.screen.win.JustReleased(pixelgl.KeyW) { // Up
		b.rows[1] |= 0x4
	}

	if b.gb.screen.win.JustReleased(pixelgl.KeyS) { // Down
		b.rows[1] |= 0x8
	}

	if b.gb.screen.win.JustReleased(pixelgl.KeyA) { // Left
		b.rows[1] |= 0x2
	}

	if b.gb.screen.win.JustReleased(pixelgl.KeyD) { // Right
		b.rows[1] |= 0x1
	}

	if b.gb.screen.win.JustReleased(pixelgl.KeyJ) { // A
		b.rows[0] |= 0x1
	}

	if b.gb.screen.win.JustReleased(pixelgl.KeyK) { // B
		b.rows[0] |= 0x2
	}

	if b.gb.screen.win.JustReleased(pixelgl.KeyEscape) {
		b.gb.close()
	}
}
