package emu

import (
	"github.com/is386/GoBoy/emu/bits"
)

var (
	WIDTH  = 160
	HEIGHT = 144
	COLORS = []uint32{0xfcefde, 0x958175, 0x894343, 0x000000}
)

type PPU struct {
	gb           *GameBoy
	cyc          int
	prevLine     uint8
	tileColorIds [160]uint8
	winLineCount int
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

	// It take 456 cycles to process one line on the screen.
	p.cyc -= cyc
	if p.cyc <= 0 {
		// Increment scanline
		p.gb.mmu.incrLY()
		currLine := p.gb.mmu.readHRAM(LY)

		// Reset scanline to 0
		if currLine > 153 {
			p.gb.mmu.writeHRAM(LY, 0)
			currLine = 0
		}

		// Reset the cycle count
		p.cyc += 456

		// Scanline goes back to the type (Vblank Interrupt)
		if currLine == uint8(HEIGHT) {
			p.gb.screen.Update()
			p.tileColorIds = [160]uint8{}
			p.gb.mmu.writeInterrupt(INT_VBLANK)
			p.winLineCount = 0
		}
	}
}

func (p *PPU) setLCDStatus() {
	// LCD Status register
	stat := p.gb.mmu.readHRAM(STAT)

	// If the LCD/PPU is not enabled, then reset/do nothing
	if !p.isLCDEnabled() {
		p.cyc = 456
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
	if currLine >= 144 {
		mode = 1
		stat = bits.Set(stat, 0)
		stat = bits.Reset(stat, 1)
		reqInt = p.isVblankInterrupt(stat)
	} else if p.cyc >= 376 {
		mode = 2
		stat = bits.Reset(stat, 0)
		stat = bits.Set(stat, 1)
		reqInt = p.isOAMInterrupt(stat)
	} else if p.cyc >= 204 {
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
		reqInt = p.isHblankInterrupt(stat)
	}

	// Request an interrupt if necessary
	if reqInt {
		p.gb.mmu.writeInterrupt(INT_LCD)
	} else if p.isLYCInterrupt(stat) && currLine != p.prevLine {
		p.prevLine = currLine
		if currLine == p.gb.mmu.readHRAM(LYC) {
			stat = bits.Set(stat, 2)
			p.gb.mmu.writeInterrupt(INT_LCD)
		}
	}

	p.gb.mmu.writeHRAM(STAT, stat)
}

func (p *PPU) drawScanline() {
	// LCDC bit 0 determines if we draw the BG
	if p.isBGEnabled() {
		p.renderBG()
	}

	if p.isWindowEnabled() {
		p.renderWindow()
	}

	// LCDC bit 1 determines if we draw the sprites
	if p.isSpritesEnabled() {
		p.renderSprites()
	}
}

func (p *PPU) renderBG() {
	scanline := p.gb.mmu.readHRAM(LY)
	scrollY := p.gb.mmu.readHRAM(SCY)
	scrollX := p.gb.mmu.readHRAM(SCX)

	var tileBaseAddr, bgMapAddr uint16

	// Determines the base address of the tiles
	if p.useFirstTileArea() {
		tileBaseAddr = 0x8000
	} else {
		tileBaseAddr = 0x9000
	}

	// Determines which of the two 32x32 background maps to
	// get the reference to the tile from
	if p.useFirstBGTileArea() {
		bgMapAddr = 0x9C00
	} else {
		bgMapAddr = 0x9800
	}

	// This array will hold the color ids of tiles for sprite
	// priority management
	p.tileColorIds = [160]uint8{}

	// Goes through each column of the screen, and draws the
	// part of the tile that is on the scanline
	for x := uint8(0); x < uint8(WIDTH); x++ {

		// Determines the x and y values after scrolling is applied
		scrolledX := uint16(x + scrollX)
		scrolledY := uint16(scanline + scrollY)

		// Determines which tile on the BG map the pixel is located
		tileX := scrolledX / 8
		tileY := scrolledY / 8

		// Determines which pixel within the tile to draw
		pixelX := scrolledX % 8
		pixelY := scrolledY % 8

		// Gets the address of the tileId in the tile map
		tileIdx := tileY*32 + tileX
		tileIdAddr := bgMapAddr + uint16(tileIdx)

		// Gets the tileId from the tile map. The tileId points
		// to the actual tile in VRAM
		tileId := p.gb.mmu.readByte(tileIdAddr)

		// Determines the address of the tile in VRAM
		tileAddr := tileBaseAddr + uint16(pixelY)*2
		if p.useFirstTileArea() {
			tileAddr += uint16(tileId) * 16
		} else {
			tileAddr = uint16(int16(tileAddr) + int16(int8(tileId))*16)
		}

		// Each row of the tile is comprised of two bytes. So
		// we use the tile address from the tileId and the row
		// of the tile we want to draw to get the two bytes
		tileByte1 := p.gb.mmu.readByte(tileAddr)
		tileByte2 := p.gb.mmu.readByte(tileAddr + 1)

		// Each bit in each byte is one pixel. The nth bit in each
		// tile byte combines to make a 2 bit color id. The current
		// pixel is the bit we want the color for
		colorId := (bits.Value(tileByte2, uint8(7-pixelX)))
		colorId <<= 1
		colorId |= bits.Value(tileByte1, uint8(7-pixelX))

		// Use the 2 bit color id and the background palette
		// to get the color for the current pixel
		color := p.getColor(colorId, BGP)
		p.gb.screen.drawPixel(int32(x), int32(scanline), color)
		p.tileColorIds[x] = colorId
	}
}

func (p *PPU) renderWindow() {
	scanline := int(p.gb.mmu.readHRAM(LY))
	windowY := int(p.gb.mmu.readHRAM(WY))
	windowX := int(p.gb.mmu.readHRAM(WX))

	if scanline < windowY || windowX > 166 {
		return
	}
	windowX -= 7

	var tileBaseAddr, bgMapAddr uint16

	// Determines the base address of the tiles
	if p.useFirstTileArea() {
		tileBaseAddr = 0x8000
	} else {
		tileBaseAddr = 0x9000
	}

	// Determines which of the two 32x32 background maps to
	// get the reference to the tile from
	if p.useFirstWindowTileArea() {
		bgMapAddr = 0x9C00
	} else {
		bgMapAddr = 0x9800
	}

	// Goes through each column of the screen, and draws the
	// part of the tile that is on the scanline
	for x := 0; (x + windowX) < WIDTH; x++ {
		// Determines which tile on the BG map the pixel is located
		tileX := x / 8
		tileY := p.winLineCount / 8

		// Determines which pixel within the tile to draw
		pixelX := x % 8
		pixelY := p.winLineCount % 8

		// Gets the address of the tileId in the tile map
		tileIdx := tileY*32 + tileX
		tileIdAddr := bgMapAddr + uint16(tileIdx)

		// Gets the tileId from the tile map. The tileId points
		// to the actual tile in VRAM
		tileId := p.gb.mmu.readByte(tileIdAddr)

		// Determines the address of the tile in VRAM
		tileAddr := tileBaseAddr + uint16(pixelY)*2
		if p.useFirstTileArea() {
			tileAddr += uint16(tileId) * 16
		} else {
			tileAddr = uint16(int16(tileAddr) + int16(int8(tileId))*16)
		}

		// Each row of the tile is comprised of two bytes. So
		// we use the tile address from the tileId and the row
		// of the tile we want to draw to get the two bytes
		tileByte1 := p.gb.mmu.readByte(tileAddr)
		tileByte2 := p.gb.mmu.readByte(tileAddr + 1)

		// Each bit in each byte is one pixel. The nth bit in each
		// ile byte combines to make a 2 bit color id. The current
		// pixel is the bit we want the color for
		colorId := (bits.Value(tileByte2, uint8(7-pixelX)))
		colorId <<= 1
		colorId |= bits.Value(tileByte1, uint8(7-pixelX))

		// Use the 2 bit color id and the background palette
		// to get the color for the current pixel
		color := p.getColor(colorId, BGP)
		p.gb.screen.drawPixel(int32(x+windowX), int32(scanline), color)
		if (x + windowX) >= 0 {
			p.tileColorIds[x+windowX] = colorId
		}
	}
	p.winLineCount += 1
}

func (p *PPU) renderSprites() {
	scanline := int(p.gb.mmu.readHRAM(LY))

	// Sprites are either 8x8 or 8x16
	spriteHeight := 8
	if p.is8x16Sprite() {
		spriteHeight = 16
	}

	// How many sprites have we drawn for this scanline
	spritesDrawn := 0

	// Keeps track of the x values for drawn pixels
	var drawnPixelXs [160]int

	// There are 40 sprites whose attributes exist in 0AM.
	// Each sprite has attributes stored in 4 bytes
	for sprite := 0; sprite < 40; sprite++ {
		// The GB could only draw 10 sprites per scan line
		if spritesDrawn >= 10 {
			break
		}

		// The sprite's index in OAM
		spriteIdx := uint16(sprite * 4)

		// The sprite's base address for its 4 attributes
		spriteBaseAddr := 0xFE00 + spriteIdx

		// Byte 0 contains the y-position of the sprite plus 16.
		// The plus 16 is for the max height of the sprite
		y := int(p.gb.cpu.readByte(spriteBaseAddr))

		// If the scanline is below the sprite's y-position or
		// if the scanline is above the sprite's height, then
		// we can't draw it
		if (scanline+16) < y || (scanline+16) >= (y+spriteHeight) {
			continue
		}
		y -= 16

		// Byte 1 contains the x-position of the sprite plus 8.
		// The plus 8 is for the max width of the sprite
		x := int(p.gb.cpu.readByte(spriteBaseAddr+1)) - 8

		// Byte 2 contains the index of the tile that contains
		// what the sprite actually looks like
		tileIdx := p.gb.cpu.readByte(spriteBaseAddr + 2)

		// Byte 3 contains 8 attributes, one for each bit. They
		// determine various things about the sprite
		attrs := p.gb.cpu.readByte(spriteBaseAddr + 3)

		// Whether or not to flip the sprite vertically/horizontally
		yFlip := p.isSpriteFlipY(attrs)
		xFlip := p.isSpriteFlipX(attrs)

		// Priority of sprite over BG
		priority := p.spriteHasPriority(attrs)

		// Bit 0 is ignored for 8x16 sprites
		if spriteHeight == 16 {
			tileIdx = bits.Reset(tileIdx, 0)
		}

		// This is the offset from the scanline to the y-coord
		// It is used to get the two bytes later
		yOffset := scanline - y
		if yFlip {
			yOffset = spriteHeight - yOffset - 1
		}

		// Gets the tile bytes for this sprite from VRAM
		tileAddr := uint16(tileIdx)*16 + uint16(yOffset)*2 + 0x8000
		tileByte1 := p.gb.mmu.readByte(tileAddr)
		tileByte2 := p.gb.mmu.readByte(tileAddr + 1)

		// Goes through the 8 pixels for current tile row
		for tilePixel := 0; tilePixel < 8; tilePixel++ {
			drawX := x + 7 - tilePixel
			if drawX < 0 || drawX >= WIDTH {
				continue
			}

			// If the pixel at drawX is not 0 (it has been drawn before)
			// and if the previous x value for this pixel has lower, it
			// has priority, so skip the current pixel
			if drawnPixelXs[drawX] != 0 && drawnPixelXs[drawX] <= x {
				continue
			}

			// Determines which pixel on the sprite we are drawing
			pixelToDraw := tilePixel
			if xFlip {
				pixelToDraw = (pixelToDraw - 7) * -1
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
			if p.useFirstPalette(attrs) {
				paletteAddr = OBP0
			} else {
				paletteAddr = OBP1
			}

			// Gets the color from the colorId
			color := p.getColor(colorId, paletteAddr)

			// If the sprite has priority or if the tile is color 0, draw the sprite
			if priority || p.tileColorIds[drawX] == 0 {
				p.gb.screen.drawPixel(int32(drawX), int32(scanline), color)
			}
			drawnPixelXs[drawX] = x
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

// LCDC Bit Checks

func (p *PPU) isLCDEnabled() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 7)
}

func (p *PPU) useFirstWindowTileArea() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 6)
}

func (p *PPU) isWindowEnabled() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 5)
}

func (p *PPU) useFirstTileArea() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 4)
}

func (p *PPU) useFirstBGTileArea() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 3)
}

func (p *PPU) is8x16Sprite() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 2)
}

func (p *PPU) isSpritesEnabled() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 1)
}

func (p *PPU) isBGEnabled() bool {
	return bits.Test(p.gb.mmu.readHRAM(LCDC), 0)
}

// LCD Status Bit Checks

func (p *PPU) isLYCInterrupt(stat uint8) bool {
	return bits.Test(stat, 6)
}

func (p *PPU) isOAMInterrupt(stat uint8) bool {
	return bits.Test(stat, 5)
}

func (p *PPU) isVblankInterrupt(stat uint8) bool {
	return bits.Test(stat, 4)
}

func (p *PPU) isHblankInterrupt(stat uint8) bool {
	return bits.Test(stat, 3)
}

// Sprite Attribute Bit Checks

func (p *PPU) spriteHasPriority(spriteAttrs uint8) bool {
	return !bits.Test(spriteAttrs, 7)
}

func (p *PPU) isSpriteFlipY(spriteAttrs uint8) bool {
	return bits.Test(spriteAttrs, 6)
}

func (p *PPU) isSpriteFlipX(spriteAttrs uint8) bool {
	return bits.Test(spriteAttrs, 5)
}

func (p *PPU) useFirstPalette(spriteAttrs uint8) bool {
	return !bits.Test(spriteAttrs, 4)
}
