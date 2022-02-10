package main

// TODO: Flag for scale
// TODO: Add print funcs to mbc
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

func parseArgs() (string, bool, bool) {
	parser := argparse.NewParser("GameFella", "A simple GameBoy emulator written in Go.")

	romFile := parser.File("r", "rom", os.O_RDWR, 0600,
		&argparse.Options{
			Required: true,
			Help:     "Path to ROM file"})

	debugFlag := parser.Flag("d", "debug",
		&argparse.Options{
			Required: false,
			Help:     "Turns on debugging mode",
			Default:  false,
		})

	noBootFlag := parser.Flag("n", "noboot",
		&argparse.Options{
			Required: false,
			Help:     "Turns off boot screen",
			Default:  false,
		})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(0)
	}

	return romFile.Name(), *debugFlag, *noBootFlag
}

func main() {
	rom, debug, noBoot := parseArgs()
	gb := emu.NewGameBoy(rom, debug, !noBoot)
	gb.Run()
}
