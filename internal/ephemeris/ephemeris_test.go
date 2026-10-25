package ephemeris

import (
	"math"
	"testing"
	"time"
)

func TestJulianDayUT(t *testing.T) {
	cases := []struct {
		t    time.Time
		want float64
	}{
		{time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), 2451545.0}, // J2000.0
		{time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), 2440587.5},
		{time.Date(1957, 10, 4, 19, 26, 24, 0, time.UTC), 2436116.31}, // Meeus, example 7.a
	}
	for _, c := range cases {
		if got := JulianDayUT(c.t); math.Abs(got-c.want) > 1e-6 {
			t.Errorf("JulianDayUT(%v) = %.6f, want %.6f", c.t, got, c.want)
		}
	}
}

func TestBodies(t *testing.T) {
	if len(Bodies) != 13 {
		t.Fatalf("got %d bodies, want 13", len(Bodies))
	}
}
