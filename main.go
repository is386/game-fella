package main

// TODO:
// - Pass blarggs timing tests
// - MBC3 RTC

// BUGS:
// - Some games don't center Nintendo logo
// - Static and crackling issues on all channels
// - Weird boot rom jingle when playing some games
// - Pokemon Yellow intro is super crackly

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/faiface/pixel/pixelgl"
	"github.com/is386/GoBoy/emu"
	"github.com/sqweek/dialog"
)

func parseArgs() (bool, bool, int) {
	parser := argparse.NewParser("GameFella", "A simple GameBoy emulator written in Go.")

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

	return *debugFlag, *noBootFlag, *scaleFlag
}

func run() {
	rom, err := dialog.File().Filter("GameBoy Rom File", "gb").Load()
	if err != nil {
		panic(err)
	}
	debug, noBoot, scale := parseArgs()
	gb := emu.NewGameBoy(rom, debug, !noBoot, scale)
	gb.Run()
}

func main() {
	pixelgl.Run(run)
}
