package main

// TODO: Flag to change color
// TODO: Flag to output cpu debugging
// TODO: Pass MBC1 Tests
// TODO: MBC2
// TODO: MBC3
// TODO: Sound

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/is386/GoBoy/emu"
)

var (
	DEBUG = false
	ROM   = "roms/zelda.gb"
)

func parseArgs() string {
	parser := argparse.NewParser("GameFella", "A simple GameBoy emulator written in Go.")
	romFile := parser.File("r", "rom", os.O_RDWR, 0600,
		&argparse.Options{
			Required: true,
			Help:     "Path to ROM file"})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(0)
	}
	return romFile.Name()
}

func main() {
	rom := parseArgs()
	gb := emu.NewGameBoy(rom, DEBUG)
	gb.Run()
}
