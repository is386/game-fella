package emu

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/is386/GoBoy/emu/apu"
	"github.com/is386/GoBoy/emu/cart"
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
	speed          int
	isCGB          bool
	isDMGCart      bool
	cyc            int
	running, debug bool
}

func NewGameBoy(rom string, bootPath string, scale int, debug bool) *GameBoy {
	gb := &GameBoy{debug: debug, running: true, speed: 1}

	gb.mmu = NewMMU(gb)
	gb.screen = NewScreen(scale)
	gb.ppu = NewPPU(gb)
	gb.apu = apu.NewAPU()
	gb.timer = NewTimer(gb)
	gb.buttons = NewButtons(gb)

	gb.loadBootRom(bootPath)
	gb.loadCart(rom)
	if !gb.isCGB {
		gb.isDMGCart = gb.cart.IsDMGCart()
	}
	gb.isCGB = !gb.isDMGCart || gb.isCGB

	gb.cpu = NewCPU(gb, gb.isCGB, bootPath != "")

	gb.setTitle(60)

	return gb
}

func (gb *GameBoy) loadBootRom(filename string) {
	boot, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Boot ROM not valid. Skipping boot screen...")
		return
	}
	gb.isCGB = len(boot) > 0x100
	gb.mmu.loadBootRom(boot)
}

func (gb *GameBoy) loadCart(filename string) {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	gb.cart = cart.NewCartridge(filename, rom)
	gb.cart.Load()
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

func (gb *GameBoy) checkBoot() {
	if !gb.mmu.bootEnabled && gb.isCGB {
		gb.isDMGCart = gb.cart.IsDMGCart()
	}
}

func (gb *GameBoy) changeSpeed() {
	if gb.mmu.prepareSpeed == 1 {
		gb.mmu.prepareSpeed = 0
		if gb.speed == 1 {
			gb.speed = 0
		} else {
			gb.speed = 1
		}
		gb.cpu.halted = false
	}
}

func (gb *GameBoy) update() {
	for gb.cyc < CPS {
		cyc := 1
		if !gb.cpu.halted {
			if gb.debug {
				gb.cpu.print()
			}
			cyc = gb.cpu.execute(gb.speed + 1)
		}
		gb.cyc += cyc
		gb.ppu.update(cyc)
		gb.timer.update(cyc)
		gb.apu.Update(cyc)
		gb.cpu.checkIME()
	}
	gb.checkBoot()
	gb.buttons.update()
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
