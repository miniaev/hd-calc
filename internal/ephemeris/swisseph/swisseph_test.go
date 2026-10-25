package swisseph

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/miniaev/hd-calc/internal/ephemeris"
)

const testData = "../../../ephe"

func TestSunAtJ2000(t *testing.T) {
	// Apparent geocentric Sun at 2000-01-01 12:00 UT: 280.37°.
	jd := ephemeris.JulianDayUT(time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC))
	got, err := New(testData).Longitude(ephemeris.Sun, jd)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-280.37) > 0.01 {
		t.Fatalf("sun = %.4f, want ~280.37", got)
	}
}

func TestAllPhysicalBodies(t *testing.T) {
	jd := ephemeris.JulianDayUT(time.Date(1990, 5, 14, 0, 20, 0, 0, time.UTC))
	eph := New(testData)
	for _, b := range ephemeris.Bodies {
		lon, err := eph.Longitude(b, jd)
		if b.Derived() {
			if !errors.Is(err, ephemeris.ErrUnsupportedBody) {
				t.Errorf("%s: want ErrUnsupportedBody, got %v", b, err)
			}
			continue
		}
		if err != nil || lon < 0 || lon >= 360 {
			t.Errorf("%s: lon=%v err=%v", b, lon, err)
		}
	}
}

func TestMissingFilesIsAnError(t *testing.T) {
	jd := ephemeris.JulianDayUT(time.Date(1990, 5, 14, 0, 0, 0, 0, time.UTC))
	if _, err := New("/nonexistent").Longitude(ephemeris.Mars, jd); err == nil {
		t.Fatal("expected an error instead of a silent Moshier fallback")
	}
}

func TestOutOfRange(t *testing.T) {
	jd := ephemeris.JulianDayUT(time.Date(1700, 1, 1, 0, 0, 0, 0, time.UTC))
	if _, err := New(testData).Longitude(ephemeris.Sun, jd); !errors.Is(err, ephemeris.ErrOutOfRange) {
		t.Fatalf("want ErrOutOfRange, got %v", err)
	}
}
