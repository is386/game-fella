package main

import (
	"fmt"
	"time"

	"github.com/is386/GoBoy/emu"
)

var (
	DEBUG     = false
	FRAMETIME = time.Second / 60
	ROM       = "roms/drmario.gb"
)

func main() {
	gb := emu.NewGameBoy(ROM, DEBUG)

	ticker := time.NewTicker(FRAMETIME)
	start := time.Now()
	frames := 0

	for range ticker.C {
		frames++
		gb.Update()
		elapsed := time.Since(start)
		if elapsed > time.Second {
			start = time.Now()
			gb.SetTitle(fmt.Sprintf("GameDude - FPS: %2v\n", frames))
			frames = 0
		}
	}
}
