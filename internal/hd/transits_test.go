package hd

import (
	"slices"
	"testing"
	"time"

	"github.com/miniaev/hd-calc/internal/ephemeris"
	"github.com/miniaev/hd-calc/internal/ephemeris/swisseph"
)

func TestTransitsAndOverlay(t *testing.T) {
	eng := NewEngine(swisseph.New("../../ephe"), "x", "")
	at := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	tr, err := eng.Transits(at)
	if err != nil {
		t.Fatal(err)
	}
	if g, _, _, _, _ := position(tr.Activations[0].Longitude); g != tr.Activations[0].Gate || len(tr.Activations) != 13 {
		t.Fatalf("activations: %+v", tr.Activations[0])
	}
	sun := tr.Activations[0].Gate

	// A chart holding the partner gate of the Sun's gate gets that channel completed.
	var partner int
	var id string
	for _, c := range channelGates {
		if c[0] == sun {
			partner, id = c[1], channelID(c)
		} else if c[1] == sun {
			partner, id = c[0], channelID(c)
		}
		if partner != 0 {
			break
		}
	}
	o, err := eng.Overlay(TransitOverlayRequest{At: at, ChartGates: []int{partner}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(o.CompletedChannels, func(c CompletedChannel) bool { return c.ID == id && slices.Equal(c.ChartGates, []int{partner}) }) {
		t.Errorf("sun gate %d + chart gate %d should complete %s: %+v", sun, partner, id, o.CompletedChannels)
	}
	if len(o.DefinedCenters) < 2 {
		t.Errorf("a completed channel defines two centers, got %v", o.DefinedCenters)
	}
	// A chart already holding the whole channel gets nothing new from it.
	o2, _ := eng.Overlay(TransitOverlayRequest{At: at, ChartGates: []int{sun, partner}})
	if slices.ContainsFunc(o2.CompletedChannels, func(c CompletedChannel) bool { return c.ID == id }) {
		t.Errorf("%s is already in the chart", id)
	}
	// Excluding the Sun (and Earth, which sits opposite) removes its gate.
	o3, _ := eng.Overlay(TransitOverlayRequest{At: at, ChartGates: []int{partner}, ExcludeBodies: []ephemeris.Body{ephemeris.Sun}})
	others := false
	for _, a := range tr.Activations[1:] {
		others = others || a.Gate == sun
	}
	if !others && slices.Contains(o3.TransitGates, sun) {
		t.Errorf("sun excluded but its gate %d still transits", sun)
	}
}
