package cart

import "time"

type RTC struct {
	time     int64
	lastTime int64
	sec      uint8
	min      uint8
	hr       uint8
	day      uint8
	latchVal uint8
}

func NewRTC() *RTC {
	return &RTC{lastTime: time.Now().Unix()}
}

func (r *RTC) latch(val uint8) {
	if r.latchVal == 0 && val == 1 {
		r.update()

		r.sec = uint8(r.time % 60)
		r.min = uint8((r.time / 60) % 60)
		r.hr = uint8((r.time / 3600) % 24)
		r.day = uint8((r.time / 3600 / 24) & 511)
	}
	r.latchVal = val
}

func (r *RTC) read(reg uint8) uint8 {
	switch reg {
	case 0x08:
		return r.sec
	case 0x09:
		return r.min
	case 0x0A:
		return r.hr
	case 0x0B:
		return r.day & 0xFF
	case 0x0C:
		return uint8(uint16(r.day) >> 8)
	default:
		return 0
	}
}

func (r *RTC) write(reg int8, val uint8) {
	switch reg {
	case 0x08:
		r.time -= int64(r.time%60) - int64(val)

	case 0x09:
		r.time -= int64((r.time/60)%60) - (int64(val) * 60)

	case 0x0A:
		r.time -= int64((r.time/3600)%24) - (int64(val) * 3600)

	case 0x0B:
		r.time -= int64((r.time/3600)/24) - (int64(val) * 3600 * 24)

	case 0x0C:
		r.time -= int64((r.time/3600)/24) - (int64(uint16(val&1)<<8) * 3600 * 24)
	}
}

func (r *RTC) update() {
	t := time.Now().Unix()
	r.time += t - r.lastTime
	r.lastTime = t
}
