package emu

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Screen struct {
	scale int
	win   *sdl.Window
	sur   *sdl.Surface
}

func NewScreen(scale int) *Screen {
	if scale < 1 {
		scale = 1
	}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	win, err := sdl.CreateWindow("", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(WIDTH*scale), int32(HEIGHT*scale), sdl.WINDOW_ALLOW_HIGHDPI)
	if err != nil {
		panic(err)
	}

	sur, err := win.GetSurface()
	if err != nil {
		panic(err)
	}

	sur.FillRect(nil, 0xF0F0F0)
	win.UpdateSurface()

	s := Screen{scale: scale, win: win, sur: sur}
	return &s
}

func (s *Screen) Destroy() {
	s.win.Destroy()
	sdl.Quit()
}

func (s *Screen) Update() {
	s.win.UpdateSurface()
}

func (s *Screen) drawPixel(x int32, y int32, color uint32) {
	s.sur.FillRect(&sdl.Rect{X: x * int32(s.scale), Y: y * int32(s.scale), W: int32(s.scale), H: int32(s.scale)}, color)
}
