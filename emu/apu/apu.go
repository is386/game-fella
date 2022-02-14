package apu

import (
	"time"

	"github.com/hajimehoshi/oto"
	"github.com/is386/GoBoy/emu/bits"
)

var (
	NR10 uint8 = 0x10
	NR11 uint8 = 0x11
	NR12 uint8 = 0x12
	NR13 uint8 = 0x13
	NR14 uint8 = 0x14
	NR21 uint8 = 0x16
	NR22 uint8 = 0x17
	NR23 uint8 = 0x18
	NR24 uint8 = 0x19
	NR30 uint8 = 0x1A
	NR31 uint8 = 0x1B
	NR32 uint8 = 0x1C
	NR33 uint8 = 0x1D
	NR34 uint8 = 0x1E
	NR41 uint8 = 0x20
	NR42 uint8 = 0x21
	NR43 uint8 = 0x22
	NR44 uint8 = 0x23
	NR50 uint8 = 0x24
	NR51 uint8 = 0x25
	NR52 uint8 = 0x26

	SAMPLE_RATE = 48000
	BUFFER_SIZE = 2048
	CPS         = 4194304 / SAMPLE_RATE
)

type APU struct {
	c2       *Channel2
	cyc      int
	player   *oto.Player
	buffer   chan [2]uint8
	volLeft  uint8
	volRight uint8
}

func NewAPU() *APU {
	apu := &APU{}
	apu.c2 = NewChannel2()
	apu.buffer = make(chan [2]uint8, BUFFER_SIZE)

	ctx, err := oto.NewContext(SAMPLE_RATE, 2, 1, SAMPLE_RATE/30)
	if err != nil {
		panic(err)
	}

	apu.player = ctx.NewPlayer()
	apu.initSound()
	return apu
}

func (a *APU) initSound() {
	frameTime := time.Second / time.Duration(30)
	ticker := time.NewTicker(frameTime)
	targetSamples := SAMPLE_RATE / 30

	go func() {
		var reading [2]uint8
		var buffer []uint8

		for range ticker.C {
			bufLen := len(a.buffer)

			if bufLen >= targetSamples/2 {
				newBuffer := make([]uint8, bufLen*2)
				for i := 0; i < bufLen*2; i += 2 {
					reading = <-a.buffer
					newBuffer[i], newBuffer[i+1] = reading[0], reading[1]
				}
				buffer = newBuffer
			}
			a.player.Write(buffer)
		}
	}()
}

func (a *APU) Update(cyc int) {
	a.cyc += cyc
	if a.cyc < CPS {
		return
	}
	a.cyc -= CPS

	a.c2.update()
	sampleL := a.c2.left
	sampleR := a.c2.right
	a.buffer <- [2]uint8{uint8(sampleL * uint16(a.volLeft)), uint8(sampleR * uint16(a.volRight))}
}

func (a *APU) ReadByte(addr uint16) uint8 {
	switch uint8(addr & 0x00FF) {
	case NR21, NR22, NR23, NR24:
		return a.c2.readByte(uint8(addr & 0x00FF))
	}
	return 0x00
}

func (a *APU) WriteByte(addr uint16, val uint8) {
	switch uint8(addr & 0x00FF) {
	case NR21, NR22, NR23, NR24:
		a.c2.writeByte(uint8(addr&0x00FF), val)
	case NR50:
		a.volLeft = (val >> 4) & 0x7
		a.volRight = val & 0x7
	case NR51:
		a.c2.rightOn = bits.Value(val, 1)
		a.c2.leftOn = bits.Value(val, 5)
	}
}
