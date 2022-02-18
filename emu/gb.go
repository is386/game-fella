package emu

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/is386/GoBoy/emu/apu"
	"github.com/is386/GoBoy/emu/cart"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	CLOCK_SPEED = 4194304
	FPS         = 60
	FRAMETIME   = time.Second / time.Duration(FPS)
	CPS         = CLOCK_SPEED / FPS
)

type GameBoy struct {
	cpu            *CPU
	mmu            *MMU
	screen         *Screen
	ppu            *PPU
	apu            *apu.APU
	timer          *Timer
	buttons        *Buttons
	cart           *cart.Cartridge
	cyc            int
	running, debug bool
}

func NewGameBoy(rom string, debug bool, bootEnabled bool, scale int) *GameBoy {
	gb := &GameBoy{debug: debug, running: true}
	gb.cpu = NewCPU(gb)
	gb.mmu = NewMMU(gb, bootEnabled)
	gb.screen = NewScreen(float64(scale))
	gb.ppu = NewPPU(gb)
	gb.apu = apu.NewAPU()
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
	gb.cart = cart.NewCartridge(filename, rom)
	for i := 0; i < len(rom); i++ {
		gb.mmu.writeByte(uint16(i), rom[i])
	}
	gb.mmu.startup = false
	gb.cart.Load()
	gb.setTitle(60)
}

func (gb *GameBoy) Run() {
	ticker := time.NewTicker(FRAMETIME)
	fpsTime := time.Now()
	saveTime := time.Now()
	frames := 0

	for range ticker.C {
		if !gb.running {
			break
		}

		frames++
		gb.update()
		gb.pollSDL()

		elapsed := time.Since(fpsTime)
		if elapsed > time.Second {
			fpsTime = time.Now()
			gb.setTitle(frames)
			frames = 0
		}

		elapsed = time.Since(saveTime)
		if elapsed > time.Second*30 {
			saveTime = time.Now()
			gb.cart.Save()
		}
	}
}

func (gb *GameBoy) pollSDL() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			gb.close()
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
		gb.apu.Update(cyc)
		gb.cyc += gb.cpu.checkIME()
	}
	gb.cyc -= CPS
}

func (gb *GameBoy) close() {
	gb.cart.Save()
	gb.screen.Destroy()
	gb.running = false
}

func (gb *GameBoy) setTitle(fps int) {
	gb.screen.win.SetTitle(fmt.Sprintf("GameFella | %s | %2v FPS", gb.cart.GetName(), fps))
}

func (gb *GameBoy) printSerialLink() {
	if !gb.debug {
		fmt.Printf("%c", gb.mmu.readHRAM(COMM1))
	}
}
