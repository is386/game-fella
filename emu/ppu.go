package emu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	WIDTH         = 160
	HEIGHT        = 144
	SCALE         = 4
	MODE1   uint8 = 144
	MODE2   int   = 376
	MODE3   int   = 204
	PALETTE       = []uint32{0xD0F8E0, 0x70C088, 0x566834, 0x201808}
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
		p.gb.mem.incrLY()
		currLine := p.gb.mem.readHRAM(LY)

		if currLine > 153 {
			p.gb.mem.writeHRAM(LY, 0)
			currLine = 0
		}

		p.scanlineCount += 456

		if currLine == uint8(HEIGHT) {
			p.gb.mem.writeInterrupt(0)
		}
	}
}

func (p *PPU) setLCDStatus() {
	stat := p.gb.mem.readHRAM(STAT)

	if !p.isLCDEnabled() {
		p.scanlineCount = 456
		p.gb.mem.writeHRAM(LY, 0)
		stat &= 252
		stat = bits.Reset(stat, 0)
		stat = bits.Reset(stat, 1)
		p.gb.mem.writeHRAM(STAT, stat)
		return
	}

	currLine := p.gb.mem.readHRAM(LY)
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
		if mode != currMode {
			p.gb.mem.doHDMATransfer()
		}
	}

	if reqInt && mode != currMode {
		p.gb.mem.writeInterrupt(1)
	}

	if currLine == p.gb.mem.readHRAM(LYC) {
		stat = bits.Set(stat, 2)
		if bits.Test(stat, 6) {
			p.gb.mem.writeInterrupt(1)
		}
	} else {
		stat = bits.Reset(stat, 2)
	}
	p.gb.mem.writeHRAM(STAT, stat)
}

func (p *PPU) isLCDEnabled() bool {
	return bits.Test(p.gb.mem.readHRAM(LCDC), 7)
}

func (p *PPU) drawScanline() {
	lcdc := p.gb.mem.readHRAM(LCDC)
	if bits.Test(lcdc, 0) {
		p.renderTiles()
	}

	if bits.Test(lcdc, 1) {
		p.renderSprites()
	}
}

func (p *PPU) renderTiles() {
	lcdc := p.gb.mem.readHRAM(LCDC)
	scanline := p.gb.mem.readHRAM(LY)

	var tileData, bgMem uint16
	unsigned := true

	scrollY := p.gb.mem.readHRAM(SCY)
	scrollX := p.gb.mem.readHRAM(SCX)
	windowY := p.gb.mem.readHRAM(WY)
	windowX := p.gb.mem.readHRAM(WX) - 7

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
			tileNum := uint8(p.gb.mem.readByte(tileAddr))
			tileLoc += uint16(tileNum * 16)
		} else {
			tileNum := int8(p.gb.mem.readByte(tileAddr))
			tileLoc += uint16(int16(tileNum) * 16)
		}

		line := (y % 8) * 2
		tile1 := p.gb.mem.readByte(tileLoc + uint16(line))
		tile2 := p.gb.mem.readByte(tileLoc + uint16(line) + 1)

		colorBit := ((int(x) % 8) - 7) * -1
		colorNum := (bits.Value(tile2, uint8(colorBit)))
		colorNum <<= 1
		colorNum |= bits.Value(tile1, uint8(colorBit))

		color := p.getColor(colorNum, BGP)
		if (scanline > 143) || (pixel > 159) {
			continue
		}
		p.gb.screen.drawPixel(int32(pixel), int32(scanline), color)
	}
}

func (p *PPU) renderSprites() {
	lcdc := p.gb.mem.readHRAM(LCDC)

	use8x16 := false
	if bits.Test(lcdc, 2) {
		use8x16 = true
	}

	for sprite := 0; sprite < 40; sprite++ {
		idx := uint16(sprite * 4)
		y := p.gb.cpu.readByte(0xFE00+idx) - 16
		x := p.gb.cpu.readByte(0xFE00+idx+1) - 8
		tileLoc := p.gb.cpu.readByte(0xFE00 + idx + 2)
		attrs := p.gb.cpu.readByte(0xFE00 + idx + 3)

		scanline := p.gb.mem.readHRAM(LY)
		yFlip := bits.Test(attrs, 6)
		xFlip := bits.Test(attrs, 5)

		ySize := 8
		if use8x16 {
			ySize = 16
		}

		if (scanline >= y) && (scanline < (y + uint8(ySize))) {
			line := int(scanline - y)

			if yFlip {
				line = ySize - line - 1
			}

			spriteAddr := uint16(0x8000+uint16(tileLoc*16)) + uint16(line*2)
			sprite1 := p.gb.mem.readByte(spriteAddr)
			sprite2 := p.gb.mem.readByte(spriteAddr + 1)

			for tilePixel := 7; tilePixel >= 0; tilePixel++ {
				colorBit := tilePixel

				if xFlip {
					colorBit -= 7
					colorBit *= -1
				}

				colorNum := (bits.Value(sprite2, uint8(colorBit)))
				colorNum <<= 1
				colorNum |= bits.Value(sprite1, uint8(colorBit))

				var colorAddr uint8
				if bits.Test(attrs, 4) {
					colorAddr = OBP0
				} else {
					colorAddr = OBP1
				}

				color := p.getColor(colorNum, colorAddr)
				xPix := 0 - tilePixel + 7
				pixel := x + uint8(xPix)
				if (scanline > 143) || (pixel > 159) {
					continue
				}
				p.gb.screen.drawPixel(int32(pixel), int32(scanline), color)
			}
		}

	}
}

func (p *PPU) getColor(colorNum uint8, addr uint8) uint32 {
	palette := p.gb.mem.readHRAM(addr)
	hi := (colorNum << 1) | 1
	lo := colorNum << 1
	c := (bits.Value(palette, hi) << 1) | bits.Value(palette, lo)
	return PALETTE[c]
}
