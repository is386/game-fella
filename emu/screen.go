package emu

import (
	"github.com/veandco/go-sdl2/sdl"
)

var (
	WHITE uint32 = 0xFFFFFF
)

type Screen struct {
	win *sdl.Window
	sur *sdl.Surface
	ren *sdl.Renderer
	tex *sdl.Texture
}

func NewScreen() *Screen {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	win := newWindow()
	ren := newRenderer(win)
	tex := newTexture(ren)
	sur := newSurface()
	screen := Screen{win: win, ren: ren, tex: tex, sur: sur}
	return &screen
}

func newWindow() *sdl.Window {
	win, err := sdl.CreateWindow("GameFella", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(WIDTH*SCALE), int32(HEIGHT*SCALE), sdl.WINDOW_ALLOW_HIGHDPI)
	if err != nil {
		panic(err)
	}
	return win
}

func newRenderer(win *sdl.Window) *sdl.Renderer {
	ren, err := sdl.CreateRenderer(win, -1,
		sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	ren.SetLogicalSize(int32(WIDTH*SCALE), int32(HEIGHT*SCALE))
	return ren
}

func newTexture(ren *sdl.Renderer) *sdl.Texture {
	tex, err := ren.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32),
		sdl.TEXTUREACCESS_STREAMING, int32(WIDTH*SCALE), int32(HEIGHT*SCALE))
	if err != nil {
		panic(err)
	}
	return tex
}

func newSurface() *sdl.Surface {
	sur, err := sdl.CreateRGBSurface(0, int32(WIDTH*SCALE), int32(HEIGHT*SCALE), 32, 0, 0, 0, 0)
	if err != nil {
		panic(err)
	}
	sur.SetRLE(true)
	return sur
}

func (s *Screen) Destroy() {
	s.tex.Destroy()
	s.ren.Destroy()
	s.win.Destroy()
	sdl.Quit()
}

func (s *Screen) Update() {
	s.updateTexture()
	s.ren.Copy(s.tex, nil, nil)
	s.ren.Present()
}

func (s *Screen) drawPixel(x int32, y int32, color uint32) {
	s.sur.FillRect(&sdl.Rect{X: x * int32(SCALE), Y: y * int32(SCALE), W: int32(SCALE), H: int32(SCALE)}, color)
}

func (s *Screen) updateTexture() {
	pixels, _, err := s.tex.Lock(nil)
	if err != nil {
		panic(err)
	}
	copy(pixels, s.sur.Pixels())
	s.tex.Unlock()
}
