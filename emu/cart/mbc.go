package cart

type MBC interface {
	readByte(addr uint16) uint8
	writeROM(addr uint16, val uint8)
	writeRAM(addr uint16, val uint8)
	getRomBank() uint32
}
