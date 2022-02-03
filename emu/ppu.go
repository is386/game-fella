package emu

import "github.com/is386/GoBoy/emu/bits"

var (
	WIDTH          = 160
	HEIGHT         = 144
	SCALE          = 4
	MODE1   uint8  = 144
	MODE2   int    = 376
	MODE3   int    = 204
	LCDC    uint16 = 0xFF40
	PALETTE        = []uint32{0xD0F8E0, 0x70C088, 0x566834, 0x201808}
)

type PPU struct {
	gb            *GameBoy
	bgPriority    [][]bool
	tileScanline  []uint8
	scanlineCount int
	isClear       bool
}

func NewPPU(gb *GameBoy) *PPU {
	p := &PPU{gb: gb}
	p.bgPriority = makeBgPriority()
	p.tileScanline = make([]uint8, WIDTH)
	return p
}

func makeBgPriority() [][]bool {
	arr := make([][]bool, WIDTH)
	for i := range arr {
		arr[i] = make([]bool, HEIGHT)
	}
	return arr
}

func (p *PPU) update(cyc int) {
	p.setLCDStatus()

	if !p.isLCDEnabled() {
		return
	}

	p.scanlineCount -= cyc
	if p.scanlineCount <= 0 {
		p.gb.mem.writeHRAM(0x44, (p.gb.mem.readHRAM(0x44) + 1))
		if p.gb.mem.readHRAM(0x44) > 153 {
			p.bgPriority = makeBgPriority()
			p.gb.mem.writeHRAM(0x44, 0)
		}

		currLine := p.gb.mem.readHRAMByte(0xFF44)
		p.scanlineCount += 456 * p.gb.speed
		if currLine == uint8(HEIGHT) {
			p.gb.mem.writeInterrupt(0)
		}
	}
}

func (p *PPU) setLCDStatus() {
	stat := p.getLCDStatus()

	if !p.isLCDEnabled() {
		p.clear()
		p.scanlineCount = 456
		p.gb.mem.writeHRAM(0x44, 0)
		stat &= 252
		stat = bits.Reset(stat, 0)
		stat = bits.Reset(stat, 1)
		p.gb.mem.writeByte(0xFF41, stat)
		return
	}

	p.isClear = false
	currLine := p.gb.mem.readHRAM(0x44)
	currMode := stat & 0x3
	sendInt := false
	var mode uint8

	if currLine >= MODE1 {
		mode = 1
		stat = bits.Set(stat, 0)
		stat = bits.Reset(stat, 1)
		sendInt = ((stat >> 4) & 1) == 1
	} else if p.scanlineCount >= MODE2 {
		mode = 2
		stat = bits.Reset(stat, 0)
		stat = bits.Set(stat, 1)
		sendInt = ((stat >> 5) & 1) == 1
	} else if p.scanlineCount >= MODE3 {
		mode = 3
		stat = bits.Set(stat, 0)
		stat = bits.Set(stat, 1)
		if mode != currMode {
			p.drawScanline(currLine)
		}
	} else {
		mode = 0
		stat = bits.Reset(stat, 0)
		stat = bits.Reset(stat, 1)
		if mode != currMode {
			p.gb.mem.doHDMATransfer()
		}
	}

	if sendInt && mode != currMode {
		p.gb.mem.writeInterrupt(1)
	}

	if currLine == p.gb.mem.readHRAM(0x45) {
		stat = bits.Set(stat, 2)
		if bits.Test(stat, 6) {
			p.gb.mem.writeInterrupt(1)
		}
	} else {
		stat = bits.Reset(stat, 2)
	}
	p.gb.mem.writeHRAM(0x41, stat)
}

func (p *PPU) drawScanline(scanline uint8) {
	lcdc := p.gb.mem.readByte(LCDC)
	if bits.Test(lcdc, 0) {
		p.renderTiles(lcdc, scanline)
	}

	if bits.Test(lcdc, 1) {
		p.renderSprites(lcdc, int32(scanline))
	}
}

func (p *PPU) renderTiles(lcdc uint8, scanline uint8) {
	scrollY := p.gb.mem.readHRAMByte(0xFF42)
	scrollX := p.gb.mem.readHRAMByte(0xFF43)
	windowY := p.gb.mem.readHRAMByte(0xFF4A)
	windowX := p.gb.mem.readHRAMByte(0xFF4B) - 7
	usingWindow, unsigned, tileData, backgroundMemory := p.getTileSettings(lcdc, windowY)

	y := scanline - windowY
	if !usingWindow {
		y = scrollY + scanline
	}

	tileRow := uint16(y/8) * 32
	palette := p.gb.mem.readHRAMByte(0xFF47)
	p.tileScanline = make([]uint8, WIDTH)

	for pixel := uint8(0); pixel < uint8(WIDTH); pixel++ {
		x := pixel + scrollX

		if usingWindow && pixel >= windowX {
			x = pixel - windowX
		}

		tileCol := uint16(x / 8)
		tileAddr := backgroundMemory + tileRow + tileCol
		tileLoc := tileData

		if unsigned {
			tileNum := int16(p.gb.mem.readVRAM(tileAddr - 0x8000))
			tileLoc += uint16(tileNum * 16)
		} else {
			tileNum := int16(int8(p.gb.mem.readVRAM(tileAddr - 0x8000)))
			tileLoc = uint16(int32(tileLoc) + int32((tileNum+128)*16))
		}

		bank := uint16(0x8000)
		tileAttr := p.gb.mem.readVRAM(tileAddr - 0x6000)
		priority := bits.Test(tileAttr, 7)
		line := (y % 8) * 2

		tile1 := p.gb.mem.readVRAM(tileLoc + uint16(line) - bank)
		tile2 := p.gb.mem.readVRAM(tileLoc + uint16(line) - bank + 1)

		colorBit := uint8(int8((x%8)-7) * -1)
		colorNum := ((bits.Value(tile2, colorBit) << 1) | bits.Value(tile1, colorBit))
		p.setTilePixel(pixel, scanline, tileAttr, colorNum, palette, priority)
	}
}

func (p *PPU) getTileSettings(lcdc uint8, windowY uint8) (bool, bool, uint16, uint16) {
	usingWindow := false
	unsigned := false
	tileData := uint16(0x8800)
	backgroundMemory := uint16(0x9800)

	if bits.Test(lcdc, 5) {
		if windowY <= p.gb.mem.readByte(0xFF44) {
			usingWindow = true
		}
	}

	if bits.Test(lcdc, 4) {
		tileData = 0x8000
		unsigned = true
	}

	bit := uint8(3)
	if usingWindow {
		bit = 6
	}
	if bits.Test(lcdc, bit) {
		backgroundMemory = 0x9C00
	}

	return usingWindow, unsigned, tileData, backgroundMemory
}

func (p *PPU) setTilePixel(x, y, tileAttr, colorNum, palette uint8, priority bool) {
	color := p.getColor(colorNum, palette)
	p.setPixel(x, y, color, true)
	p.tileScanline[x] = colorNum
}

func (p *PPU) renderSprites(lcdc uint8, scanline int32) {
	ySize := int32(8)
	if bits.Test(lcdc, 2) {
		ySize = 16
	}

	palette1 := p.gb.mem.readByte(0xFF48)
	palette2 := p.gb.mem.readByte(0xFF49)

	minX := make([]int32, WIDTH)
	lineSprites := 0

	for sprite := uint16(0); sprite < 40; sprite++ {
		idx := sprite * 4

		y := int32(p.gb.mem.readByte(uint16(0xFE00+idx))) - 16
		if scanline < y || scanline >= (y+ySize) {
			continue
		}

		if lineSprites >= 10 {
			break
		}
		lineSprites++

		x := int32(p.gb.mem.readByte(uint16(0xFE00+idx+1))) - 8
		tileLoc := p.gb.mem.readByte(uint16(0xFE00 + idx + 2))
		attrs := p.gb.mem.readByte(uint16(0xFE00 + idx + 3))
		bank := uint16(0)

		yFlip := bits.Test(attrs, 6)
		xFlip := bits.Test(attrs, 5)
		priority := !bits.Test(attrs, 7)

		line := scanline - y
		if yFlip {
			line = ySize - line - 1
		}

		dataAddr := (uint16(tileLoc) * 16) + uint16(line*2) + (bank * 0x2000)
		sprite1 := p.gb.mem.readVRAM(dataAddr)
		sprite2 := p.gb.mem.readVRAM(dataAddr + 1)

		for tilePixel := uint8(0); tilePixel < 8; tilePixel++ {
			pixel := int16(x) + int16(7-tilePixel)
			if pixel < 0 || pixel >= int16(WIDTH) {
				continue
			}

			if minX[pixel] != 0 && minX[pixel] <= x+100 {
				continue
			}

			colorBit := tilePixel
			if xFlip {
				colorBit = uint8(int8(colorBit-7) * -1)
			}

			colorNum := ((bits.Value(sprite2, colorBit) << 1) | bits.Value(sprite1, colorBit))
			if colorNum == 0 {
				continue
			}

			palette := palette1
			if bits.Test(attrs, 4) {
				palette = palette2
			}

			color := p.getColor(colorNum, palette)
			p.setPixel(uint8(pixel), uint8(scanline), color, priority)
			minX[pixel] = x + 100
		}
	}
}

func (p *PPU) setPixel(x uint8, y uint8, color uint32, priority bool) {
	if priority && !p.bgPriority[x][y] || p.tileScanline[x] == 0 {
		p.gb.screen.drawPixel(int32(x), int32(y), color)
	}
}

func (p *PPU) getColor(colorNum uint8, palette uint8) uint32 {
	hi := (colorNum << 1) | 1
	lo := colorNum << 1
	c := (bits.Value(palette, hi) << 1) | bits.Value(palette, lo)
	return PALETTE[c]
}

func (p *PPU) getLCDStatus() uint8 {
	return p.gb.mem.getLCDStatus()
}

func (p *PPU) isLCDEnabled() bool {
	return p.gb.mem.isLCDEnabled()
}

func (p *PPU) clear() {
	if p.isClear {
		return
	}

	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			p.gb.screen.drawPixel(int32(i), int32(j), WHITE)
		}
	}
	p.isClear = true
}
