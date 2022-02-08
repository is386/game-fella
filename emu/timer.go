package emu

import "github.com/is386/GoBoy/emu/bits"

var (
	FREQ_MAP = map[uint8]int{
		0: 1024,
		1: 16,
		2: 64,
		3: 256,
	}
)

type Timer struct {
	gb     *GameBoy
	divCyc int
	count  int
}

func NewTimer(gb *GameBoy) *Timer {
	return &Timer{gb: gb}
}

func (t *Timer) update(cyc int) {
	t.updateDiv(cyc)
	if t.isTimerEnabled() {
		t.count += cyc
		freq := t.getTimerFreq()
		for t.count >= freq {
			t.count -= freq
			t.updateTima()
		}
	}
}

func (t *Timer) updateDiv(cyc int) {
	t.divCyc += cyc
	if t.divCyc >= 255 {
		t.divCyc -= 255
		t.gb.mmu.incrDiv()
	}
}

func (t *Timer) isTimerEnabled() bool {
	return bits.Test(t.gb.mmu.readHRAM(TAC), 2)
}

func (t *Timer) getTimerFreq() int {
	return FREQ_MAP[t.gb.mmu.readHRAM(TAC)&0x3]
}

func (t *Timer) updateTima() {
	tima := t.gb.mmu.readHRAM(TIMA)
	if tima == 0xFF {
		t.gb.mmu.writeHRAM(TIMA, t.gb.mmu.readHRAM(TMA))
		t.gb.mmu.writeInterrupt(INT_TIMER)
	} else {
		t.gb.mmu.incrTima()
	}
}

func (t *Timer) resetTimer() {
	t.count = 0
}

func (t *Timer) resetDivCyc() {
	t.divCyc = 0
}
