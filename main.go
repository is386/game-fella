package main

import (
	"fmt"
	"time"

	"github.com/is386/GoBoy/emu"
)

var (
	DEBUG = false
	FPS   = 60
	ROM   = "roms/2.gb"
)

func main() {
	gb := emu.NewGameBoy(ROM, DEBUG)

	frameTime := time.Second / 60
	ticker := time.NewTicker(frameTime)
	start := time.Now()
	frames := 0

	for range ticker.C {
		frames++
		gb.Update()
		elapsed := time.Since(start)
		if elapsed > time.Second {
			start = time.Now()
			gb.SetTitle(fmt.Sprintf("GoBoy - FPS: %2v\n", frames))
			frames = 0
		}
	}
}
