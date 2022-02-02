package emu

type GameBoy struct {
	cpu   *CPU
	mem   *Memory
	debug bool
}

func NewGameBoy(debug bool) *GameBoy {
	gb := &GameBoy{debug: debug}
	gb.cpu = NewCPU(gb)
	gb.mem = NewMemory(gb)
	return gb
}

func (gb *GameBoy) Run(rom string) {
	gb.cpu.loadRom(rom)
	for {
		gb.update()
	}
}

func (gb *GameBoy) update() {
	if gb.debug {
		gb.cpu.print()
	}
	gb.cpu.execute()
}
