package main

import (
	"github.com/is386/GoBoy/emu"
)

var DEBUG = false

func main() {
	cpu := emu.NewCPU()
	cpu.LoadRom("roms/10.gb")
	for {
		cpu.Execute(DEBUG)
	}
}
