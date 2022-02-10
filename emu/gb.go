package emu

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/is386/GoBoy/emu/cart"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	FRAMETIME = time.Second / 60
	CPS       = 4194304 / 60
)

type GameBoy struct {
	cpu     *CPU
	mmu     *MMU
	screen  *Screen
	ppu     *PPU
	timer   *Timer
	buttons *Buttons
	cyc     int
	debug   bool
}

func NewGameBoy(rom string, debug bool) *GameBoy {
	gb := &GameBoy{debug: debug}
	gb.cpu = NewCPU(gb)
	gb.mmu = NewMMU(gb)
	gb.screen = NewScreen()
	gb.ppu = NewPPU(gb)
	gb.timer = NewTimer(gb)
	gb.buttons = NewButtons(gb)
	gb.loadRom(rom)
	return gb
}

func (gb *GameBoy) loadRom(filename string) {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	gb.mmu.cart = cart.NewCartridge(rom)
	for i := 0; i < len(rom); i++ {
		gb.mmu.writeByte(uint16(i), rom[i])
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
		switch e := event.(type) {
		case *sdl.QuitEvent:
			return false
		case *sdl.KeyboardEvent:
			switch e.Type {
			case sdl.KEYDOWN:
				gb.buttons.keyDown(e.Keysym.Sym)
			case sdl.KEYUP:
				gb.buttons.keyUp(e.Keysym.Sym)
			}
		}
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
		fmt.Printf("%c", gb.mmu.readHRAM(COMM1))
	}
}
