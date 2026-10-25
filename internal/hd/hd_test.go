package hd

import (
	"errors"
	"testing"
	"time"
)

func TestTablesConsistent(t *testing.T) {
	seen := map[int]bool{}
	for _, g := range gateOrder {
		if g < 1 || g > 64 || seen[g] {
			t.Fatalf("gateOrder: bad or duplicate gate %d", g)
		}
		seen[g] = true
	}
	if len(gateCenter) != 64 {
		t.Fatalf("gateCenter covers %d gates, want 64", len(gateCenter))
	}
	if len(channelGates) != 36 {
		t.Fatalf("%d channels, want 36", len(channelGates))
	}
	inChannel := map[int]bool{}
	for _, c := range channelGates {
		if c[0] >= c[1] {
			t.Errorf("channel %v: lower gate must come first", c)
		}
		if gateCenter[c[0]] == gateCenter[c[1]] {
			t.Errorf("channel %v joins a center to itself", c)
		}
		inChannel[c[0]], inChannel[c[1]] = true, true
	}
	if len(inChannel) != 64 {
		t.Errorf("%d gates belong to a channel, want all 64", len(inChannel))
	}
}

func TestPosition(t *testing.T) {
	cases := []struct {
		lon              float64
		gate, line, base int
	}{
		{302.0, 41, 1, 1},                      // mandala start
		{302.0 + gateWidth - 1e-9, 41, 6, 5},   // last sliver of gate 41
		{302.0 + gateWidth, 19, 1, 1},          // next gate
		{0.0, 25, 2, 1},                        // 0° Aries: gate 25 spans 358.25°..3.875°
		{358.25 - 1e-9, 36, 6, 5},              // just before gate 25
		{223.25 + 3*lineWidth + 1e-9, 1, 4, 1}, // gate 1 begins at 13°15' Scorpio
	}
	for _, c := range cases {
		g, l, _, _, b := position(c.lon)
		if g != c.gate || l != c.line || b != c.base {
			t.Errorf("position(%v) = gate %d line %d base %d, want %d %d %d", c.lon, g, l, b, c.gate, c.line, c.base)
		}
	}
}

func TestTypeAndAuthority(t *testing.T) {
	chart := func(gates ...int) (string, string, string) {
		active := map[int]bool{}
		for _, g := range gates {
			active[g] = true
		}
		g := newGraph(definedChannels(active))
		ty := g.hdType()
		return ty, g.authority(ty), g.definition()
	}
	cases := []struct {
		name          string
		gates         []int
		ty, auth, def string
	}{
		{"nothing", nil, "reflector", "lunar", "none"},
		{"sacral-root only", []int{3, 60}, "generator", "sacral", "single"},
		{"sacral to throat", []int{20, 34}, "manifesting_generator", "sacral", "single"},
		{"emotions to throat", []int{12, 22}, "manifestor", "emotional", "single"},
		{"ego to throat", []int{21, 45}, "manifestor", "ego_manifested", "single"},
		{"ego to G", []int{25, 51}, "projector", "ego_projected", "single"},
		{"G to throat", []int{1, 8}, "projector", "self_projected", "single"},
		{"head-ajna-throat", []int{4, 63, 17, 62}, "projector", "mental", "single"},
		{"spleen to G", []int{10, 57}, "projector", "splenic", "single"},
		{"two islands", []int{4, 63, 3, 60}, "generator", "sacral", "split"},
		{"root via spleen to throat", []int{18, 58, 16, 48}, "manifestor", "splenic", "single"},
	}
	for _, c := range cases {
		ty, auth, def := chart(c.gates...)
		if ty != c.ty || auth != c.auth || def != c.def {
			t.Errorf("%s: got %s/%s/%s, want %s/%s/%s", c.name, ty, auth, def, c.ty, c.auth, c.def)
		}
	}
}

func TestCrossAngle(t *testing.T) {
	for p, want := range map[[2]int]string{
		{1, 3}: "right_angle", {4, 6}: "right_angle", {3, 5}: "right_angle",
		{4, 1}: "juxtaposition",
		{5, 1}: "left_angle", {6, 3}: "left_angle",
	} {
		if got := crossAngle(Profile{p[0], p[1]}); got != want {
			t.Errorf("%d/%d: got %s, want %s", p[0], p[1], got, want)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	cases := []struct {
		local, tz, utc, warning string
	}{
		{"1990-05-14T03:20", "Europe/Moscow", "1990-05-13T23:20:00Z", ""},                       // MSD +4
		{"1990-01-14T03:20", "Europe/Moscow", "1990-01-14T00:20:00Z", ""},                       // MSK +3
		{"2012-06-01T12:00", "Europe/Moscow", "2012-06-01T08:00:00Z", ""},                       // 2011–2014: permanent +4
		{"2015-06-01T12:00", "Europe/Moscow", "2015-06-01T09:00:00Z", ""},                       // since 2014: +3
		{"1985-03-31T02:30", "Europe/Moscow", "1985-03-30T23:30:00Z", "nonexistent_local_time"}, // 02:00→03:00
		{"1985-09-29T02:30", "Europe/Moscow", "1985-09-28T22:30:00Z", "ambiguous_local_time"},   // 03:00→02:00, first pass
		{"2021-03-28T02:30", "Europe/Berlin", "2021-03-28T01:30:00Z", "nonexistent_local_time"},
		{"1990-05-14T03:20:30", "Asia/Yekaterinburg", "1990-05-13T21:20:30Z", ""},
	}
	for _, c := range cases {
		u, w, err := resolveLocal(c.local, c.tz)
		if err != nil {
			t.Errorf("%s %s: %v", c.local, c.tz, err)
			continue
		}
		got := u.UTC().Format(time.RFC3339)
		gotW := ""
		if len(w) > 0 {
			gotW = w[0].Code
		}
		if got != c.utc || gotW != c.warning {
			t.Errorf("%s %s: got %s %q, want %s %q", c.local, c.tz, got, gotW, c.utc, c.warning)
		}
	}
	if _, _, err := resolveLocal("1990-05-14T03:20", "Mars/Olympus"); !errors.Is(err, ErrUnknownTimezone) {
		t.Errorf("unknown zone: got %v", err)
	}
	if _, _, err := resolveLocal("14.05.1990 03:20", "Europe/Moscow"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("bad format: got %v", err)
	}
}
