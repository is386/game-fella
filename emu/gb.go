package emu

import (
	"fmt"
)

var (
	CPS = 4194304 / 60
)

type GameBoy struct {
	cpu   *CPU
	mem   *Memory
	timer *Timer
	cyc   int
	speed int
	debug bool
}

func NewGameBoy(debug bool) *GameBoy {
	gb := &GameBoy{debug: debug}
	gb.cpu = NewCPU(gb)
	gb.mem = NewMemory(gb)
	gb.timer = NewTimer(gb)
	gb.speed = 1
	return gb
}

func (gb *GameBoy) Run(rom string) {
	gb.cpu.loadRom(rom)
	for {
		gb.update()
	}
}

func (gb *GameBoy) update() {
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
		gb.timer.update(cyc)
		gb.cyc += gb.cpu.checkIME()
	}
}

func (gb *GameBoy) printSerialLink() {
	if !gb.debug {
		fmt.Printf("%c", gb.mem.readByte(0xFF01))
	}
}
