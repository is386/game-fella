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

func (c *CPU) add8(a uint8, b uint8, cy uint8) uint8 {
	ans := uint16(a) + uint16(b) + uint16(cy)
	c.flags.setZero(ans)
	c.flags.N = 0
	c.flags.setCarryAdd(ans)
	c.flags.setHalfCarryAdd((a & 0xF), (b & 0xF))
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
