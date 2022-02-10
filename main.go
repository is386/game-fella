package main

// TODO: Pass rom through CLI
// TODO: Flag to change color
// TODO: Flag to output cpu debugging
// TODO: Pass MBC1 Tests
// TODO: MBC2
// TODO: MBC3
// TODO: Sound

import (
	"github.com/is386/GoBoy/emu"
)

var (
	DEBUG = false
	ROM   = "roms/zelda.gb"
)

func main() {
	gb := emu.NewGameBoy(ROM, DEBUG)
	gb.Run()
}
