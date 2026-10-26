package hd

import (
	"errors"
	"testing"
	"time"

	"github.com/miniaev/hd-calc/internal/ephemeris/swisseph"
)

func TestChartRange(t *testing.T) {
	eng := NewEngine(swisseph.New("../../ephe"), swisseph.Name+"-"+swisseph.Version, "")
	r, err := eng.ChartRange(ChartRangeRequest{BirthLocalFrom: "1999-04-10T00:00", BirthLocalTo: "1999-04-10T23:59", TZ: "Europe/Moscow"})
	if err != nil {
		t.Fatal(err)
	}
	moon := false
	for _, v := range r.Varying {
		if v.Element == "activation.personality.moon" {
			moon = true
		}
		if len(v.Segments) > 3 && (v.Element == "activation.design.sun" || v.Element == "activation.personality.sun") {
			t.Errorf("%s: the Sun moves about one line a day, got %d segments", v.Element, len(v.Segments))
		}
		// segments tile the interval and every change point agrees with a direct calculation
		if !v.Segments[0].FromUTC.Equal(r.FromUTC) || !v.Segments[len(v.Segments)-1].ToUTC.Equal(r.ToUTC) {
			t.Errorf("%s: segments do not cover the interval", v.Element)
		}
		for i, s := range v.Segments {
			if i > 0 && !s.FromUTC.Equal(v.Segments[i-1].ToUTC) {
				t.Errorf("%s: gap before segment %d", v.Element, i)
			}
			for _, at := range []time.Time{s.FromUTC, s.ToUTC.Add(-time.Minute)} {
				c, err := eng.chartAt(ChartRequest{TZ: "Europe/Moscow"}, at, nil)
				if err != nil {
					t.Fatal(err)
				}
				if got := elements(c)[v.Element]; got != s.Value {
					t.Errorf("%s at %s = %s, segment says %s", v.Element, at, got, s.Value)
				}
			}
		}
	}
	if !moon {
		t.Error("personality moon should change during a whole day")
	}

	// The known chart (18:15) must fall in a segment with its values.
	c, _ := eng.Chart(ChartRequest{BirthLocal: "1999-04-10T18:15", TZ: "Europe/Moscow"})
	for _, v := range r.Varying {
		for _, s := range v.Segments {
			if !c.BirthUTC.Before(s.FromUTC) && c.BirthUTC.Before(s.ToUTC) && elements(c)[v.Element] != s.Value {
				t.Errorf("%s at 18:15: chart %s, range %s", v.Element, elements(c)[v.Element], s.Value)
			}
		}
	}
}

func TestChartRangeLimits(t *testing.T) {
	eng := NewEngine(swisseph.New("../../ephe"), "x", "")
	if _, err := eng.ChartRange(ChartRangeRequest{BirthLocalFrom: "2000-01-01T00:00", BirthLocalTo: "2000-01-02T00:01", TZ: "UTC"}); !errors.Is(err, ErrIntervalTooLong) {
		t.Errorf("25h: err = %v, want ErrIntervalTooLong", err)
	}
	if _, err := eng.ChartRange(ChartRangeRequest{BirthLocalFrom: "2000-01-01T10:00", BirthLocalTo: "2000-01-01T09:00", TZ: "UTC"}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("reversed: err = %v, want ErrInvalidInput", err)
	}
	// Doubled hour at the end widens the interval: Moscow 1999-10-31 02:00–03:00 happened twice.
	r, err := eng.ChartRange(ChartRangeRequest{BirthLocalFrom: "1999-10-31T02:30", BirthLocalTo: "1999-10-31T02:30", TZ: "Europe/Moscow"})
	if err != nil {
		t.Fatal(err)
	}
	if r.ToUTC.Sub(r.FromUTC) != time.Hour || len(r.Warnings) != 2 {
		t.Errorf("ambiguous range = %s..%s, %d warnings; want 1h, 2", r.FromUTC, r.ToUTC, len(r.Warnings))
	}
}
