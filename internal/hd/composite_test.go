package hd

import (
	"errors"
	"slices"
	"testing"

	"github.com/miniaev/hd-calc/internal/ephemeris/swisseph"
)

func gatesOf(t *testing.T, eng *Engine, local string) []int {
	c, err := eng.Chart(ChartRequest{BirthLocal: local, TZ: "Europe/Moscow"})
	if err != nil {
		t.Fatal(err)
	}
	var gs []int
	for _, g := range c.Gates {
		gs = append(gs, g.Gate)
	}
	return gs
}

func compositeOf(gates ...[]int) CompositeRequest {
	var r CompositeRequest
	for _, g := range gates {
		r.Members = append(r.Members, struct {
			Gates []int `json:"gates"`
		}{g})
	}
	return r
}

// The pair analysed by hand earlier (1999-04-10 18:15 and 1974-12-29 08:45, Moscow).
func TestCompositePair(t *testing.T) {
	eng := NewEngine(swisseph.New("../../ephe"), "x", "")
	a, b := gatesOf(t, eng, "1999-04-10T18:15"), gatesOf(t, eng, "1974-12-29T08:45")
	c, err := eng.Composite(compositeOf(a, b))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string][]string{}
	for _, ch := range c.Channels {
		got[ch.Kind] = append(got[ch.Kind], ch.ID)
	}
	want := map[string][]string{
		"electromagnetic": {"4-63", "17-62", "27-50", "32-54", "35-36"},
		"dominance":       {"3-60", "16-48", "19-49", "28-38", "42-53"},
		"compromise":      {"9-52", "18-58"},
	}
	for k, w := range want {
		g := got[k]
		slices.Sort(g)
		slices.Sort(w)
		if !slices.Equal(g, w) {
			t.Errorf("%s = %v, want %v", k, g, w)
		}
	}
	if c.Group.Centers.G || c.Group.Centers.Heart {
		t.Errorf("G and Heart are open in both, so in the composite too: %+v", c.Group.Centers)
	}
	if c.Penta != nil {
		t.Error("no Penta for a pair")
	}
}

func TestCompositeGroup(t *testing.T) {
	eng := NewEngine(swisseph.New("../../ephe"), "x", "")
	// Three people each holding one half of the Penta channels.
	c, err := eng.Composite(compositeOf([]int{7, 1, 13}, []int{31, 8}, []int{5, 2}))
	if err != nil {
		t.Fatal(err)
	}
	formed := map[string]bool{}
	for _, p := range c.Penta.Channels {
		formed[p.ID] = p.Formed
	}
	if !formed["7-31"] || !formed["1-8"] || formed["13-33"] || formed["5-15"] {
		t.Errorf("penta: %+v", c.Penta.Channels)
	}
	if !slices.Equal(c.Penta.MissingGates, []int{14, 15, 29, 33, 46}) {
		t.Errorf("missing gates = %v", c.Penta.MissingGates)
	}
	// Without member 1 (31, 8) both Throat–G channels go, and with them the Throat and G.
	im := c.Impact[1]
	if !slices.Equal(im.LostChannels, []string{"1-8", "7-31"}) || !slices.Equal(im.LostCenters, []string{Throat, GCenter}) {
		t.Errorf("impact of member 1: %+v", im)
	}
	if _, err := eng.Composite(compositeOf([]int{1})); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("one member: err = %v", err)
	}
}
