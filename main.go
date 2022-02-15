package main

// TODO:
// - Pass blarggs timing tests
// - MBC3 RTC
// - Pass MBC3 Tests

// BUGS:
// - Some games don't center Nintendo logo
// - Static and crackling issues on all channels
// - Weird boot rom jingle when playing Zelda

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/is386/GoBoy/emu"
)

func parseArgs() (string, bool, bool, int) {
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

	scaleFlag := parser.Int("s", "--scale",
		&argparse.Options{
			Required: false,
			Help:     "Scale of the screen",
			Default:  3,
		})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(0)
	}

	return romFile.Name(), *debugFlag, *noBootFlag, *scaleFlag
}

func main() {
	rom, debug, noBoot, scale := parseArgs()
	gb := emu.NewGameBoy(rom, debug, !noBoot, scale)
	gb.Run()
}
