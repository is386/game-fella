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
