package hd

import "math"

// The Rave Mandala: 64 gates of 5.625° each around the tropical zodiac,
// starting with gate 41 at 302° (2° Aquarius). Each gate has 6 lines, each
// line 6 colors, each color 6 tones, each tone 5 bases.
const (
	mandalaStart = 302.0
	gateWidth    = 360.0 / 64
	lineWidth    = gateWidth / 6
	colorWidth   = lineWidth / 6
	toneWidth    = colorWidth / 6
	baseWidth    = toneWidth / 5
)

var gateOrder = [64]int{
	41, 19, 13, 49, 30, 55, 37, 63, 22, 36, 25, 17, 21, 51, 42, 3,
	27, 24, 2, 23, 8, 20, 16, 35, 45, 12, 15, 52, 39, 53, 62, 56,
	31, 33, 7, 4, 29, 59, 40, 64, 47, 6, 46, 18, 48, 57, 32, 50,
	28, 44, 1, 43, 14, 34, 9, 5, 26, 11, 10, 58, 38, 54, 61, 60,
}

func normDeg(d float64) float64 {
	d = math.Mod(d, 360)
	if d < 0 {
		d += 360
	}
	return d
}

// position splits a longitude into gate, line, color, tone and base.
func position(lon float64) (gate, line, color, tone, base int) {
	off := normDeg(lon - mandalaStart)
	idx := min(int(off/gateWidth), 63)
	in := off - float64(idx)*gateWidth
	sub := func(x, w float64, n int) (int, float64) {
		k := min(int(x/w), n-1)
		return k + 1, x - float64(k)*w
	}
	line, in = sub(in, lineWidth, 6)
	color, in = sub(in, colorWidth, 6)
	tone, in = sub(in, toneWidth, 6)
	base, _ = sub(in, baseWidth, 5)
	return gateOrder[idx], line, color, tone, base
}
