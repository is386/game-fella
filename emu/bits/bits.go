package bits

func Test(val uint8, bit uint8) bool {
	return ((val >> bit) & 1) == 1
}

func Value(val uint8, bit uint8) uint8 {
	return (val >> bit) & 1
}

func Set(val uint8, bit uint8) uint8 {
	return val | (1 << bit)
}

func Reset(val uint8, bit uint8) uint8 {
	return val & ^(1 << bit)
}
