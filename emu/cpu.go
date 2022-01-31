package emu

type CPU struct {
	reg   *Registers
	flags *Flags
	sp    uint16
	pc    uint16
}
