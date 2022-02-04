package emu

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	FRAMETIME = time.Second / 60
	CPS       = 4194304 / 60
)

type GameBoy struct {
	cpu    *CPU
	mem    *Memory
	screen *Screen
	ppu    *PPU
	timer  *Timer
	cyc    int
	debug  bool
}

func NewGameBoy(rom string, debug bool) *GameBoy {
	gb := &GameBoy{debug: debug}
	gb.cpu = NewCPU(gb)
	gb.mem = NewMemory(gb)
	gb.screen = NewScreen()
	gb.ppu = NewPPU(gb)
	gb.timer = NewTimer(gb)
	gb.LoadRom(rom)
	return gb
}

func (gb *GameBoy) LoadRom(filename string) {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(rom); i++ {
		gb.mem.writeByte(uint16(i), rom[i])
	}
}

func (gb *GameBoy) Run() {
	ticker := time.NewTicker(FRAMETIME)
	start := time.Now()
	frames := 0

	for range ticker.C {
		frames++

		gb.update()
		running := gb.pollSDL()
		if !running {
			break
		}

		elapsed := time.Since(start)
		if elapsed > time.Second {
			start = time.Now()
			gb.setTitle(fmt.Sprintf("GameFella - FPS: %2v\n", frames))
			frames = 0
		}
	}
}

func (gb *GameBoy) pollSDL() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return false
		}
		return true
	}
	return true
}

func (gb *GameBoy) update() {
	gb.cyc = 0
	for gb.cyc < CPS {
		cyc := 4
		if !gb.cpu.halted {
			if gb.debug {
				gb.cpu.print()
			}
			cyc = gb.cpu.execute()
		}
		gb.cyc += cyc
		gb.ppu.update(cyc)
		gb.timer.update(cyc)
		gb.cyc += gb.cpu.checkIME()
	}
	gb.screen.Update()
}

func (gb *GameBoy) setTitle(title string) {
	gb.screen.win.SetTitle(title)
}

func (gb *GameBoy) printSerialLink() {
	if !gb.debug {
		fmt.Printf("%c", gb.mem.readByte(0xFF01))
	}
}
