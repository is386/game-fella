package emu

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
			t.gb.mem.updateTima()
		}
	}
}

func (t *Timer) updateDiv(cyc int) {
	t.divCyc += cyc
	if t.divCyc >= 255 {
		t.divCyc -= 255
		t.gb.mem.incrDiv()
	}
}

func (t *Timer) isTimerEnabled() bool {
	return t.gb.mem.isTimerEnabled()
}

func (t *Timer) getTimerFreq() int {
	return FREQ_MAP[t.gb.mem.getTimerFreq()]
}

func (t *Timer) resetTimer() {
	t.count = 0
}

func (t *Timer) resetDivCyc() {
	t.divCyc = 0
}
