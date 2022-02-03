package emu

var (
	OG_WIDTH  = 160
	OG_HEIGHT = 144
	SCALE     = 2
	WIDTH     = OG_WIDTH * SCALE
	HEIGHT    = OG_HEIGHT * SCALE
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
	p.setMode()
}

func (p *PPU) setMode() {
	//mode := p.getLCDMode()

	if !p.isLCDEnabled() {
		p.clear()
	}
}

func (p *PPU) getLCDMode() uint8 {
	return p.gb.mem.getLCDMode()
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
