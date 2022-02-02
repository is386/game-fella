package main

import (
	"github.com/is386/GoBoy/emu"
)

var DEBUG = false

// PASS: 1, 3, 4, 5, 6, 7, 8, 9
// FAIL: 2, 10, 11

func main() {
	cpu := emu.NewCPU(DEBUG)
	cpu.LoadRom("roms/10.gb")
	for {
		cpu.Execute()
	}
}
