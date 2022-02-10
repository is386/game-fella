package main

// TODO: Memory Bank 1
// TODO: Controller Input
// TODO: Sound
// TODO: Pass rom through CLI
// TODO: Flag to change color
// TODO: Flag to output cpu debugging

import (
	"github.com/is386/GoBoy/emu"
)

var (
	DEBUG = true
	ROM   = "roms/rom_512kb.gb"
)

func main() {
	gb := emu.NewGameBoy(ROM, DEBUG)
	gb.Run()
}
