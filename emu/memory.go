package emu

import "fmt"

type Memory struct {
	mem []uint8
}

func NewMemory(size int32) *Memory {
	return &Memory{mem: make([]uint8, size)}
}

func (m *Memory) readByte(addr uint16) uint8 {
	return m.mem[addr]
}

func (m *Memory) writeByte(addr uint16, val uint8) {
	if (val == 0x81) && (addr == 0xFF02) {
		fmt.Println("as[fpkas[fka")
	}
	m.mem[addr] = val
}
