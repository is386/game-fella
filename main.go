package main

import "github.com/is386/GoBoy/emu"

func main() {
	cpu := emu.NewCPU()
	cpu.LoadRom("roms/10-bit ops.gb")
	for {
		cpu.Execute()
	}
}
