package emu

type Flags struct {
	Z uint8
	N uint8
	H uint8
	C uint8
}

func (f *Flags) getF() uint8 {
	val := uint8(0)
	val |= f.Z << 7
	val |= f.N << 6
	val |= f.H << 5
	val |= f.C << 4
	return val
}

func (f *Flags) setZero(val uint16) {
	if val == 0 {
		f.Z = 1
	} else {
		f.Z = 0
	}
}

func (f *Flags) setCarryAdd(val uint16) {
	if val > 0xFF {
		f.C = 1
	} else {
		f.C = 0
	}
}

func (f *Flags) setHalfCarryAdd(a uint8, b uint8) {
	if a+b > 0xF {
		f.H = 1
	} else {
		f.H = 0
	}
}
