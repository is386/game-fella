package emu

var (
	MEM_SIZE int32 = 65536
)

type CPU struct {
	mem   *Memory
	reg   *Registers
	flags *Flags
	sp    uint16
	pc    uint16
	cyc   int
}

func NewCPU() *CPU {
	return &CPU{mem: NewMemory(MEM_SIZE)}
}

func (c *CPU) readByte(addr uint16) uint8 {
	return c.mem.readByte(addr)
}

func (c *CPU) readByteHL() uint8 {
	return c.readByte(c.reg.getHL())
}

func (c *CPU) nextByte() uint8 {
	val := c.readByte(c.pc)
	c.pc++
	return val
}

func (c *CPU) fetch() uint8 {
	return c.nextByte()
}

func (c *CPU) decode(opcode uint8) func(*CPU) {
	return INSTRUCTIONS[opcode]
}

func (c *CPU) Execute() {
	opcode := c.fetch()
	c.cyc += CYCLES[opcode]
	instr := c.decode(opcode)
	instr(c)
}

func flip(val uint8) uint8 {
	if val == 1 {
		return 0
	} else {
		return 1
	}
}

func (c *CPU) add8(a uint8, b uint8, cy uint8) uint8 {
	ans := uint16(a) + uint16(b) + uint16(cy)
	c.flags.setZero8(ans)
	c.flags.N = 0
	c.flags.setHalfCarryAdd8((a & 0xF), (b & 0xF))
	c.flags.setCarryAdd8(ans)
	return uint8(ans)
}

func (c *CPU) add16(a uint16, b uint16, cy uint8) uint16 {
	ans := uint32(a) + uint32(b) + uint32(cy)
	c.flags.setZero16(ans)
	c.flags.N = 0
	c.flags.setHalfCarryAdd16((a & 0xFF), (b & 0xFF))
	c.flags.setCarryAdd16(ans)
	return uint16(ans)
}

func (c *CPU) sub(a uint8, b uint8, cy uint8) uint8 {
	cy = flip(cy)
	ans := c.add8(a, ^b, cy)
	c.flags.N = 1
	c.flags.C = flip(c.flags.C)
	return uint8(ans)
}

func (c *CPU) and(a uint8, b uint8) uint8 {
	ans := uint16(a) & uint16(b)
	c.flags.setZero8(ans)
	c.flags.N = 0
	c.flags.H = 1
	c.flags.C = 0
	return uint8(ans)
}

func (c *CPU) or(a uint8, b uint8) uint8 {
	ans := uint16(a) | uint16(b)
	c.flags.setZero8(ans)
	c.flags.N = 0
	c.flags.H = 0
	c.flags.C = 0
	return uint8(ans)
}

func (c *CPU) xor(a uint8, b uint8) uint8 {
	ans := uint16(a) ^ uint16(b)
	c.flags.setZero8(ans)
	c.flags.N = 0
	c.flags.H = 0
	c.flags.C = 0
	return uint8(ans)
}

func addAB(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.B, 0)
}

func addAC(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.C, 0)
}

func addAD(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.D, 0)
}

func addAE(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.E, 0)
}

func addAH(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.H, 0)
}

func addAL(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.L, 0)
}

func addAHL(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.readByteHL(), 0)
}

func addAA(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.A, 0)
}

func adi(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.nextByte(), 0)
}

func adcAB(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.B, c.flags.C)
}

func adcAC(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.C, c.flags.C)
}

func adcAD(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.D, c.flags.C)
}

func adcAE(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.E, c.flags.C)
}

func adcAH(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.H, c.flags.C)
}

func adcAL(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.L, c.flags.C)
}

func adcAHL(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.readByteHL(), c.flags.C)
}

func adcAA(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.reg.A, c.flags.C)
}

func aci(c *CPU) {
	c.reg.A = c.add8(c.reg.A, c.nextByte(), c.flags.C)
}

func addHLBC(c *CPU) {
	c.reg.setHL(c.add16(c.reg.getHL(), c.reg.getBC(), 0))
}

func addHLDE(c *CPU) {
	c.reg.setHL(c.add16(c.reg.getHL(), c.reg.getDE(), 0))
}

func addHLHL(c *CPU) {
	c.reg.setHL(c.add16(c.reg.getHL(), c.reg.getHL(), 0))
}

func addHLSP(c *CPU) {
	c.reg.setHL(c.add16(c.reg.getHL(), c.sp, 0))
}

func subAB(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.B, 0)
}

func subAC(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.C, 0)
}

func subAD(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.D, 0)
}

func subAE(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.E, 0)
}

func subAH(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.H, 0)
}

func subAL(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.L, 0)
}

func subAHL(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.readByteHL(), 0)
}

func subAA(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.A, 0)
}

func sui(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.nextByte(), 0)
}

func sbcAB(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.B, c.flags.C)
}

func sbcAC(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.C, c.flags.C)
}

func sbcAD(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.D, c.flags.C)
}

func sbcAE(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.E, c.flags.C)
}

func sbcAH(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.H, c.flags.C)
}

func sbcAL(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.L, c.flags.C)
}

func sbcAHL(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.readByteHL(), c.flags.C)
}

func sbcAA(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.reg.A, c.flags.C)
}

func sbi(c *CPU) {
	c.reg.A = c.sub(c.reg.A, c.nextByte(), c.flags.C)
}

func andAB(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.reg.B)
}

func andAC(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.reg.C)
}

func andAD(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.reg.D)
}

func andAE(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.reg.E)
}

func andAH(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.reg.H)
}

func andAL(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.reg.L)
}

func andAHL(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.readByteHL())
}

func andAA(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.reg.A)
}

func ani(c *CPU) {
	c.reg.A = c.and(c.reg.A, c.nextByte())
}

func orAB(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.reg.B)
}

func orAC(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.reg.C)
}

func orAD(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.reg.D)
}

func orAE(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.reg.E)
}

func orAH(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.reg.H)
}

func orAL(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.reg.L)
}

func orAHL(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.readByteHL())
}

func orAA(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.reg.A)
}

func ori(c *CPU) {
	c.reg.A = c.or(c.reg.A, c.nextByte())
}

func xorAB(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.reg.B)
}

func xorAC(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.reg.C)
}

func xorAD(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.reg.D)
}

func xorAE(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.reg.E)
}

func xorAH(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.reg.H)
}

func xorAL(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.reg.L)
}

func xorAHL(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.readByteHL())
}

func xorAA(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.reg.A)
}

func xri(c *CPU) {
	c.reg.A = c.xor(c.reg.A, c.nextByte())
}
