package emu

import (
	"fmt"
	"io/ioutil"
)

var (
	CPS = 4194304 / 60
)

type GameBoy struct {
	cpu    *CPU
	mem    *Memory
	screen *Screen
	ppu    *PPU
	timer  *Timer
	cyc    int
	speed  int
	debug  bool
}

func NewGameBoy(rom string, debug bool) *GameBoy {
	gb := &GameBoy{debug: debug}
	gb.cpu = NewCPU(gb)
	gb.mem = NewMemory(gb)
	gb.screen = NewScreen()
	gb.ppu = NewPPU(gb)
	gb.timer = NewTimer(gb)
	gb.speed = 1
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

func (gb *GameBoy) Update() {
	gb.cyc = 0
	for gb.cyc < (CPS * gb.speed) {
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

func (gb *GameBoy) SetTitle(title string) {
	gb.screen.win.SetTitle(title)
}

func (gb *GameBoy) printSerialLink() {
	if !gb.debug {
		fmt.Printf("%c", gb.mem.readByte(0xFF01))
	}
}
