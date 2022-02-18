package emu

import (
	"image/color"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type Screen struct {
	win   *pixelgl.Window
	pic   *pixel.PictureData
	scale float64
}

func NewScreen(scale float64) *Screen {
	s := &Screen{scale: scale}

	cfg := pixelgl.WindowConfig{
		Bounds: pixel.R(0, 0, float64(WIDTH)*scale, float64(HEIGHT)*scale),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	pic := &pixel.PictureData{
		Pix:    make([]color.RGBA, WIDTH*HEIGHT),
		Stride: WIDTH,
		Rect:   pixel.R(0, 0, WIDTH, HEIGHT),
	}

	s.win = win
	s.pic = pic
	s.win.Clear(COLORS[0])
	return s
}

func (s *Screen) Destroy() {
	s.win.Destroy()
}

func (s *Screen) Update(screenData [WIDTH][HEIGHT]color.RGBA) {
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			s.pic.Pix[(HEIGHT-1-y)*WIDTH+x] = screenData[x][y]
		}
	}

	spr := pixel.NewSprite(pixel.Picture(s.pic), pixel.R(0, 0, WIDTH, HEIGHT))
	spr.Draw(s.win, pixel.IM)
	s.updateCamera()
	s.win.Update()
}

func (s *Screen) updateCamera() {
	xScale := s.win.Bounds().W() / 160
	yScale := s.win.Bounds().H() / 144
	scale := math.Min(yScale, xScale)

	shift := s.win.Bounds().Size().Scaled(0.5).Sub(pixel.ZV)
	cam := pixel.IM.Scaled(pixel.ZV, scale).Moved(shift)
	s.win.SetMatrix(cam)
}

func (s *Screen) closed() bool {
	return s.win.Closed()
}
