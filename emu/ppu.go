package emu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	WIDTH        = 160
	HEIGHT       = 144
	SCALE        = 3
	MODE1  uint8 = 144
	MODE2  int   = 376
	MODE3  int   = 204
	DMG          = []uint32{0x0FBC9B, 0x0FAC8B, 0x306230, 0x0F380F}
	MGB          = []uint32{0xCDDBE0, 0x949FA8, 0x666B70, 0x262B2B}
	COLORS       = DMG
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

	// Operations related to moving/resetting the scanline
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
	// LCD Status register
	stat := p.gb.mmu.readHRAM(STAT)

	// If the LCD/PPU is not enabled, then reset/do nothing
	if !p.isLCDEnabled() {
		p.scanlineCount = 456
		p.gb.mmu.writeHRAM(LY, 0)
		stat &= 252
		stat = bits.Reset(stat, 0)
		stat = bits.Reset(stat, 1)
		p.gb.mmu.writeHRAM(STAT, stat)
		return
	}

	// LY contains the current scanline
	currLine := p.gb.mmu.readHRAM(LY)

	// STAT contains the LCD's current mode
	currMode := stat & 0x3
	mode := uint8(0)
	reqInt := false

	// Mode 0: pad time when we don't draw to the whole line
	// Mode 1: pad time for the 10 additional invisible rows
	// Mode 2: fetch asset
	// Mode 3: render
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

	// Write an interrupt between modes
	if reqInt && mode != currMode {
		p.gb.mmu.writeInterrupt(1)
	}

	// Used to change palettes or perform special effects
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

	// LCDC bit 0 determines if we draw the BG/Window
	if bits.Test(lcdc, 0) {
		p.renderTiles()
	}

	// LCDC bit 1 determines if we draw sprites
	if bits.Test(lcdc, 1) {
		p.renderSprites()
	}
}

func (p *PPU) renderTiles() {
	lcdc := p.gb.mmu.readHRAM(LCDC)
	scanline := p.gb.mmu.readHRAM(LY)

	var tileBaseAddr, bgMapAddr uint16
	unsigned := true

	scrollY := p.gb.mmu.readHRAM(SCY)
	scrollX := p.gb.mmu.readHRAM(SCX)
	windowY := p.gb.mmu.readHRAM(WY)
	windowX := p.gb.mmu.readHRAM(WX) - 7

	// Determines if we draw to the window or not. The background
	// is scrollable while the window is fixed
	usingWindow := bits.Test(lcdc, 5) && windowY <= scanline

	// Determines the base address of the tiles
	if bits.Test(lcdc, 4) {
		tileBaseAddr = 0x8000
	} else {
		tileBaseAddr = 0x9000
		unsigned = false
	}

	// Determines which of the two 32x32 background maps to
	// get the reference to the tile from
	if !usingWindow {
		if bits.Test(lcdc, 3) {
			bgMapAddr = 0x9C00
		} else {
			bgMapAddr = 0x9800
		}
	} else {
		if bits.Test(lcdc, 6) {
			bgMapAddr = 0x9C00
		} else {
			bgMapAddr = 0x9800
		}
	}

	// The y-position of the tile
	y := scanline - windowY
	if !usingWindow {
		y = scrollY + scanline
	}

	// Determines the row of the tile in the tile map
	tileMapRow := uint16(y/8) * 32

	// Goes through each column of the screen, and draws the
	// part of the tile that is on the scanline
	for column := uint8(0); column < uint8(WIDTH); column++ {
		// The x-position of the tile
		x := column + scrollX
		if usingWindow && column >= windowX {
			x = column - windowX
		}

		// Determines the column of the tile in the tile map
		tileMapCol := uint16(x / 8)

		// Gets the address of the tileId in the tile map
		tileIdAddr := bgMapAddr + tileMapRow + tileMapCol

		// Gets the tileId from the tile map. The tileId points
		// to the actual tile in VRAM
		tileAddr := tileBaseAddr
		if unsigned {
			tileId := int16(p.gb.mmu.readByte(tileAddr))
			tileAddr += uint16(tileId * 16)
		} else {
			tileNum := int16(int8(p.gb.mmu.readByte(tileIdAddr)))
			tileAddr = uint16(int32(tileAddr) + int32(tileNum*16))
		}

		// Gets a row of the current tile
		tileRow := (y % 8) * 2

		// Each row of the tile is comprised of two bytes. So
		// we use the tile address from the tileId and the row
		// of the tile we want to draw to get the two bytes
		tileByte1 := p.gb.mmu.readByte(tileAddr + uint16(tileRow))
		tileByte2 := p.gb.mmu.readByte(tileAddr + uint16(tileRow) + 1)

		// The pixel we are drawing, based on the current column
		pixelToDraw := uint8(int8((x%8)-7) * -1)

		// Each bit in each byte is one pixel. The nth bit in each
		// ile byte combines to make a 2 bit color id. The current
		// pixel is the bit we want the color for
		colorId := (bits.Value(tileByte2, uint8(pixelToDraw)))
		colorId <<= 1
		colorId |= bits.Value(tileByte1, uint8(pixelToDraw))

		// Use the 2 bit color id and the background palette
		// to get the color for the current pixel
		color := p.getColor(colorId, BGP)
		p.gb.screen.drawPixel(int32(column), int32(scanline), color)
	}
}

func (p *PPU) renderSprites() {
	lcdc := p.gb.mmu.readHRAM(LCDC)
	scanline := p.gb.mmu.readHRAM(LY)

	// Sprites are either 8x8 or 8x16
	spriteHeight := uint8(8)
	if bits.Test(lcdc, 2) {
		spriteHeight = 16
	}

	// How many sprites have we drawn for this scanline
	spritesDrawn := 0

	// There are 40 sprites whose attributes exist in 0AM.
	// Each sprite has attributes stored in 4 bytes
	for sprite := 0; sprite < 40; sprite++ {
		// The GB could only draw 10 sprites per scan line
		if spritesDrawn == 10 {
			break
		}

		// The sprite's index in OAM
		spriteIdx := uint16(sprite * 4)

		// The sprite's base address for its 4 attributes
		spriteBaseAddr := 0xFE00 + spriteIdx

		// Byte 0 contains the y-position of the sprite plus 16.
		// The plus 16 is for the max height of the sprite
		y := p.gb.cpu.readByte(spriteBaseAddr) - 16

		// If the scanline is below the sprite's y-position or
		// if the scanline is above the sprite's height, then
		// we can't draw it
		if scanline < y || scanline >= (y+spriteHeight) {
			continue
		}

		// Byte 1 contains the x-position of the sprite plus 8.
		// The plus 8 is for the max width of the sprite
		x := p.gb.cpu.readByte(spriteBaseAddr+1) - 8

		// Byte 2 contains the index of the tile that contains
		// what the sprite actually looks like
		tileIdx := p.gb.cpu.readByte(spriteBaseAddr + 2)

		// Byte 3 contains 8 attributes, one for each bit. They
		// determine various things about the sprite
		attrs := p.gb.cpu.readByte(spriteBaseAddr + 3)

		// Whether or not to flip the sprite vertically/horizontally
		yFlip := bits.Test(attrs, 6)
		xFlip := bits.Test(attrs, 5)

		line := scanline - y
		if yFlip {
			line = spriteHeight - line - 1
		}

		// Gets the tile bytes for this sprite from VRAM
		tileAddr := uint16(tileIdx)*16 + uint16(line)*2 + 0x8000
		tileByte1 := p.gb.mmu.readByte(tileAddr)
		tileByte2 := p.gb.mmu.readByte(tileAddr + 1)

		// Goes through the 8 pixels for current tile row
		for tilePixel := uint8(0); tilePixel < 8; tilePixel++ {
			pixel := int16(x) + int16(7-tilePixel)
			if pixel < 0 || pixel >= int16(WIDTH) {
				continue
			}

			// Determines which pixel we are drawing
			pixelToDraw := tilePixel
			if xFlip {
				pixelToDraw = uint8(int8(pixelToDraw-7) * -1)
			}

			// Gets the color on the tile for the pixel we are
			// drawing
			colorId := (bits.Value(tileByte2, uint8(pixelToDraw)))
			colorId <<= 1
			colorId |= bits.Value(tileByte1, uint8(pixelToDraw))

			// Color 0 is just transparent for sprites
			if colorId == 0 {
				continue
			}

			// Determines if we are using the palette at 0xFF48
			// or 0xFF49. Each of these palettes will utilize
			// the 4 colors in different ways
			var paletteAddr uint8
			if bits.Test(attrs, 4) {
				paletteAddr = OBP1
			} else {
				paletteAddr = OBP0
			}

			// Gets the color from the colorId
			color := p.getColor(colorId, paletteAddr)
			p.gb.screen.drawPixel(int32(pixel), int32(scanline), color)
		}
		spritesDrawn++
	}
}

func (p *PPU) getColor(colorId uint8, paletteAddr uint8) uint32 {
	// Gets the palette at the address
	palette := p.gb.mmu.readHRAM(paletteAddr)
	// Uses the colorId to get the color number 1-4 from the palette
	hi := (colorId << 1) | 1
	lo := colorId << 1
	colorNum := (bits.Value(palette, hi) << 1) | bits.Value(palette, lo)
	return COLORS[colorNum]
}
