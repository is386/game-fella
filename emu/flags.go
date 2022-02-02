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

func (f *Flags) setZero8(val uint16) {
	if uint8(val) == 0 {
		f.Z = 1
	} else {
		f.Z = 0
	}
}

func (f *Flags) setZero16(val uint32) {
	if uint16(val) == 0 {
		f.Z = 1
	} else {
		f.Z = 0
	}
}

func (f *Flags) setCarryAdd8(val uint16) {
	if val > 0xFF {
		f.C = 1
	} else {
		f.C = 0
	}
}

func (f *Flags) setCarryAdd16(val uint32) {
	if val > 0xFFFF {
		f.C = 1
	} else {
		f.C = 0
	}
}

func (f *Flags) setHalfCarryAdd8(a uint8, b uint8, cy uint8) {
	if ((a & 0xF) + (b & 0xF) + cy) > 0xF {
		f.H = 1
	} else {
		f.H = 0
	}
}

func (f *Flags) setHalfCarryAdd16(a uint16, ans uint32) {
	if uint32(a&0xFFF) > (ans & 0xFFF) {
		f.H = 1
	} else {
		f.H = 0
	}
}
