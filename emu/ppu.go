package emu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	WIDTH         = 160
	HEIGHT        = 144
	SCALE         = 3
	MODE1   uint8 = 144
	MODE2   int   = 376
	MODE3   int   = 204
	DMG           = []uint32{0x0FBC9B, 0x0FAC8B, 0x306230, 0x0F380F}
	MGB           = []uint32{0xCDDBE0, 0x949FA8, 0x666B70, 0x262B2B}
	PALETTE       = DMG
)

type PPU struct {
	gb            *GameBoy
	scanlineCount int
}

func NewPPU(gb *GameBoy) *PPU {
	p := &PPU{gb: gb}
	return p
}

func (p *PPU) update(cyc int) {
	p.setLCDStatus()

	if !p.isLCDEnabled() {
		return
	}

	p.scanlineCount -= cyc
	if p.scanlineCount <= 0 {
		p.gb.mmu.incrLY()
		currLine := p.gb.mmu.readHRAM(LY)

		if currLine > 153 {
			p.gb.mmu.writeHRAM(LY, 0)
			currLine = 0
		}

		p.scanlineCount += 456

		if currLine == uint8(HEIGHT) {
			p.gb.mmu.writeInterrupt(0)
		}
	}
}

func (p *PPU) setLCDStatus() {
	stat := p.gb.mmu.readHRAM(STAT)

	if !p.isLCDEnabled() {
		p.scanlineCount = 456
		p.gb.mmu.writeHRAM(LY, 0)
		stat &= 252
		stat = bits.Reset(stat, 0)
		stat = bits.Reset(stat, 1)
		p.gb.mmu.writeHRAM(STAT, stat)
		return
	}

	currLine := p.gb.mmu.readHRAM(LY)
	currMode := stat & 0x3
	mode := uint8(0)
	reqInt := false

	if currLine >= MODE1 {
		mode = 1
		stat = bits.Set(stat, 0)
		stat = bits.Reset(stat, 1)
		reqInt = bits.Test(stat, 4)
	} else if p.scanlineCount >= MODE2 {
		mode = 2
		stat = bits.Reset(stat, 0)
		stat = bits.Set(stat, 1)
		reqInt = bits.Test(stat, 5)
	} else if p.scanlineCount >= MODE3 {
		mode = 3
		stat = bits.Set(stat, 0)
		stat = bits.Set(stat, 1)
		if mode != currMode {
			p.drawScanline()
		}
	} else {
		mode = 0
		stat = bits.Reset(stat, 0)
		stat = bits.Reset(stat, 1)
		reqInt = bits.Test(stat, 3)
	}

	if reqInt && mode != currMode {
		p.gb.mmu.writeInterrupt(1)
	}

	if currLine == p.gb.mmu.readHRAM(LYC) {
		stat = bits.Set(stat, 2)
		if bits.Test(stat, 6) {
			p.gb.mmu.writeInterrupt(1)
		}
	} else {
		stat = bits.Reset(stat, 2)
	}
	p.gb.mmu.writeHRAM(STAT, stat)
}

func (p *PPU) isLCDEnabled() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 7)
}

func (p *PPU) drawScanline() {
	lcdc := p.gb.mmu.readHRAM(LCDC)
	if bits.Test(lcdc, 0) {
		p.renderTiles()
	}

	if bits.Test(lcdc, 1) {
		p.renderSprites()
	}
}

func (p *PPU) renderTiles() {
	lcdc := p.gb.mmu.readHRAM(LCDC)
	scanline := p.gb.mmu.readHRAM(LY)

	var tileData, bgMem uint16
	unsigned := true

	scrollY := p.gb.mmu.readHRAM(SCY)
	scrollX := p.gb.mmu.readHRAM(SCX)
	windowY := p.gb.mmu.readHRAM(WY)
	windowX := p.gb.mmu.readHRAM(WX) - 7

	usingWindow := bits.Test(lcdc, 5) && windowY <= scanline

	if bits.Test(lcdc, 4) {
		tileData = 0x8000
	} else {
		tileData = 0x9000
		unsigned = false
	}

	if !usingWindow {
		if bits.Test(lcdc, 3) {
			bgMem = 0x9C00
		} else {
			bgMem = 0x9800
		}
	} else {
		if bits.Test(lcdc, 6) {
			bgMem = 0x9C00
		} else {
			bgMem = 0x9800
		}
	}

	y := scanline - windowY
	if !usingWindow {
		y = scrollY + scanline
	}

	tileRow := uint16(y/8) * 32
	for pixel := uint8(0); pixel < uint8(WIDTH); pixel++ {
		x := pixel + scrollX

		if usingWindow && pixel >= windowX {
			x = pixel - windowX
		}

		tileCol := uint16(x / 8)
		tileAddr := bgMem + tileRow + tileCol
		tileLoc := tileData
		if unsigned {
			tileNum := int16(p.gb.mmu.readByte(tileAddr))
			tileLoc += uint16(tileNum * 16)
		} else {
			tileNum := int16(int8(p.gb.mmu.readByte(tileAddr)))
			tileLoc = uint16(int32(tileLoc) + int32(tileNum*16))
		}

		line := (y % 8) * 2
		tile1 := p.gb.mmu.readByte(tileLoc + uint16(line))
		tile2 := p.gb.mmu.readByte(tileLoc + uint16(line) + 1)

		colorBit := uint8(int8((x%8)-7) * -1)
		colorNum := (bits.Value(tile2, uint8(colorBit)))
		colorNum <<= 1
		colorNum |= bits.Value(tile1, uint8(colorBit))

		color := p.getColor(colorNum, BGP)
		p.gb.screen.drawPixel(int32(pixel), int32(scanline), color)
	}
}

func (p *PPU) renderSprites() {
	lcdc := p.gb.mmu.readHRAM(LCDC)
	scanline := p.gb.mmu.readHRAM(LY)

	ySize := uint8(8)
	if bits.Test(lcdc, 2) {
		ySize = 16
	}

	spritesDrawn := 0
	for sprite := 0; sprite < 40; sprite++ {
		if spritesDrawn == 10 {
			break
		}

		idx := uint16(sprite * 4)
		y := p.gb.cpu.readByte(0xFE00+idx) - 16
		if scanline < y || scanline >= (y+ySize) {
			continue
		}

		x := p.gb.cpu.readByte(0xFE00+idx+1) - 8
		tileLoc := p.gb.cpu.readByte(0xFE00 + idx + 2)
		attrs := p.gb.cpu.readByte(0xFE00 + idx + 3)

		yFlip := bits.Test(attrs, 6)
		xFlip := bits.Test(attrs, 5)

		line := scanline - y
		if yFlip {
			line = ySize - line - 1
		}

		spriteAddr := uint16(tileLoc)*16 + uint16(line)*2 + 0x8000
		sprite1 := p.gb.mmu.readByte(spriteAddr)
		sprite2 := p.gb.mmu.readByte(spriteAddr + 1)

		for tilePixel := uint8(0); tilePixel < 8; tilePixel++ {
			pixel := int16(x) + int16(7-tilePixel)
			if pixel < 0 || pixel >= int16(WIDTH) {
				continue
			}

			colorBit := tilePixel
			if xFlip {
				colorBit = uint8(int8(colorBit-7) * -1)
			}

			colorNum := (bits.Value(sprite2, uint8(colorBit)))
			colorNum <<= 1
			colorNum |= bits.Value(sprite1, uint8(colorBit))
			if colorNum == 0 {
				continue
			}

			var colorAddr uint8
			if bits.Test(attrs, 4) {
				colorAddr = OBP1
			} else {
				colorAddr = OBP0
			}

			color := p.getColor(colorNum, colorAddr)
			p.gb.screen.drawPixel(int32(pixel), int32(scanline), color)
		}
		spritesDrawn++
	}
}

func (p *PPU) getColor(colorNum uint8, addr uint8) uint32 {
	palette := p.gb.mmu.readHRAM(addr)
	hi := (colorNum << 1) | 1
	lo := colorNum << 1
	c := (bits.Value(palette, hi) << 1) | bits.Value(palette, lo)
	return PALETTE[c]
}
