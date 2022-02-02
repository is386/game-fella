package emu

type Flags struct {
	Z uint8
	N uint8
	H uint8
	C uint8
}

func NewFlags() *Flags {
	return &Flags{Z: 1, N: 0, H: 1, C: 1}
}

func (f *Flags) getF() uint8 {
	val := uint8(0)
	val |= f.Z << 7
	val |= f.N << 6
	val |= f.H << 5
	val |= f.C << 4
	return val
}

func (f *Flags) setZero(cond bool) {
	if cond {
		f.Z = 1
	} else {
		f.Z = 0
	}
}

func (f *Flags) setHalfCarry(cond bool) {
	if cond {
		f.H = 1
	} else {
		f.H = 0
	}
}

func (f *Flags) setCarry(cond bool) {
	if cond {
		f.C = 1
	} else {
		f.C = 0
	}
}
