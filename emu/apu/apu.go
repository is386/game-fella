package apu

import (
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

var (
	SAMPLE_RATE float64 = 44100.0
)

type APU struct {
	c2 Channel2
}

func NewAPU() *APU {
	sr := beep.SampleRate(SAMPLE_RATE)
	speaker.Init(sr, sr.N(time.Second/30))

	apu := &APU{}
	apu.c2 = NewChannel2(220.0)
	apu.c2.self = &apu.c2

	stream := &beep.Mixer{}
	stream.Add(apu.c2)
	speaker.Play(stream)

	return apu
}
