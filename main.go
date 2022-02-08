package main

// TODO: GB Boot Screen
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
	DEBUG = false
	ROM   = "roms/tetris.gb"
)

func main() {
	gb := emu.NewGameBoy(ROM, DEBUG)
	gb.Run()
}
