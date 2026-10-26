package hd

import (
	"fmt"
	"slices"
	"time"

	"github.com/miniaev/hd-calc/internal/ephemeris"
)

type Transits struct {
	EngineVersion string       `json:"engine_version"`
	At            time.Time    `json:"at"`
	Activations   []Activation `json:"activations"`
	Gates         []int        `json:"gates"`
	Channels      []string     `json:"channels"`
}

type TransitOverlayRequest struct {
	At            time.Time        `json:"at"`
	ChartGates    []int            `json:"chart_gates"`
	ExcludeBodies []ephemeris.Body `json:"exclude_bodies,omitempty"`
}

type CompletedChannel struct {
	ID           string `json:"id"`
	ChartGates   []int  `json:"chart_gates"`
	TransitGates []int  `json:"transit_gates"`
}

type TransitOverlay struct {
	EngineVersion     string             `json:"engine_version"`
	At                time.Time          `json:"at"`
	TransitGates      []int              `json:"transit_gates"`
	CompletedChannels []CompletedChannel `json:"completed_channels"`
	DefinedCenters    []string           `json:"defined_centers"`
}

// Transits returns where the 13 bodies stand at a moment. Transits are the
// same for everyone, so callers may cache them by moment.
func (e *Engine) Transits(at time.Time) (*Transits, error) {
	acts, err := e.activations(ephemeris.JulianDayUT(at))
	if err != nil {
		return nil, err
	}
	gates := gateSet(acts, nil)
	chs := []string{}
	for _, c := range definedChannels(toSet(gates)) {
		chs = append(chs, c.ID)
	}
	return &Transits{EngineVersion: e.EngineVersion, At: at.UTC(), Activations: acts, Gates: gates, Channels: chs}, nil
}

// Overlay reports what transits add to a chart: channels completed only
// with the transit's help and centers that this makes defined for a while.
// Bodies in ExcludeBodies are ignored — e.g. the Moon, for summaries over
// days and weeks where its few hours per gate are noise.
func (e *Engine) Overlay(req TransitOverlayRequest) (*TransitOverlay, error) {
	if req.At.IsZero() {
		return nil, fmt.Errorf("%w: at is required", ErrInvalidInput)
	}
	chart := map[int]bool{}
	for _, g := range req.ChartGates {
		if g < 1 || g > 64 {
			return nil, fmt.Errorf("%w: gate %d", ErrInvalidInput, g)
		}
		chart[g] = true
	}
	for _, b := range req.ExcludeBodies {
		if !slices.Contains(ephemeris.Bodies, b) {
			return nil, fmt.Errorf("%w: unknown body %q", ErrInvalidInput, b)
		}
	}
	acts, err := e.activations(ephemeris.JulianDayUT(req.At))
	if err != nil {
		return nil, err
	}
	tGates := gateSet(acts, req.ExcludeBodies)
	transit := toSet(tGates)

	both := map[int]bool{}
	for g := range chart {
		both[g] = true
	}
	for g := range transit {
		both[g] = true
	}
	completed := []CompletedChannel{}
	for _, g := range channelGates {
		if !both[g[0]] || !both[g[1]] || (chart[g[0]] && chart[g[1]]) {
			continue
		}
		c := CompletedChannel{ID: channelID(g), ChartGates: []int{}, TransitGates: []int{}}
		for _, x := range g {
			if chart[x] {
				c.ChartGates = append(c.ChartGates, x)
			}
			if transit[x] {
				c.TransitGates = append(c.TransitGates, x)
			}
		}
		completed = append(completed, c)
	}
	before, after := newGraph(definedChannels(chart)), newGraph(definedChannels(both))
	centers := []string{}
	for _, c := range []string{Head, Ajna, Throat, GCenter, Heart, Sacral, SolarPlexus, Spleen, Root} {
		if after.defined(c) && !before.defined(c) {
			centers = append(centers, c)
		}
	}
	return &TransitOverlay{EngineVersion: e.EngineVersion, At: req.At.UTC(), TransitGates: tGates,
		CompletedChannels: completed, DefinedCenters: centers}, nil
}

func gateSet(acts []Activation, exclude []ephemeris.Body) []int {
	seen := map[int]bool{}
	for _, a := range acts {
		if !slices.Contains(exclude, a.Body) {
			seen[a.Gate] = true
		}
	}
	out := make([]int, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	slices.Sort(out)
	return out
}

func toSet(gs []int) map[int]bool {
	m := map[int]bool{}
	for _, g := range gs {
		m[g] = true
	}
	return m
}
