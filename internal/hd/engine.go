// Package hd turns planetary longitudes into a Human Design chart. It knows
// nothing about where longitudes come from — only the ephemeris.Ephemeris
// interface.
package hd

import (
	"errors"
	"math"
	"slices"
	"time"

	"github.com/miniaev/hd-calc/internal/ephemeris"
)

// CalcVersion is the semver of the calculation code. Bump it on any change
// that can alter a result; together with the ephemeris id it forms
// engine_version.
const CalcVersion = "0.1.0"

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrUnknownTimezone = errors.New("unknown time zone")
	ErrIntervalTooLong = errors.New("interval too long")
)

// designArc is how far the Sun moves back from birth to the Design moment.
const designArc = 88.0

type Engine struct {
	eph ephemeris.Ephemeris
	// EngineVersion is "<CalcVersion>/<ephemeris id>", e.g. "0.1.0/swisseph-2.10.03".
	EngineVersion string
	TZDBVersion   string
}

func NewEngine(eph ephemeris.Ephemeris, ephemerisID, tzdbVersion string) *Engine {
	return &Engine{
		eph:           eph,
		EngineVersion: CalcVersion + "/" + ephemerisID,
		TZDBVersion:   tzdbVersion,
	}
}

// Chart computes the full chart for a birth moment given in local time.
func (e *Engine) Chart(req ChartRequest) (*Chart, error) {
	birth, warnings, err := resolveLocal(req.BirthLocal, req.TZ)
	if err != nil {
		return nil, err
	}
	return e.chartAt(req, birth, warnings)
}

// chartAt computes the chart for a resolved UTC birth instant.
func (e *Engine) chartAt(req ChartRequest, birth time.Time, warnings []TimeWarning) (*Chart, error) {
	jd := ephemeris.JulianDayUT(birth)
	pers, err := e.activations(jd)
	if err != nil {
		return nil, err
	}
	designJD, err := e.designJD(jd, pers[0].Longitude)
	if err != nil {
		return nil, err
	}
	des, err := e.activations(designJD)
	if err != nil {
		return nil, err
	}

	gates := activeGates(pers, des)
	active := map[int]bool{}
	for _, g := range gates {
		active[g.Gate] = true
	}
	channels := definedChannels(active)
	g := newGraph(channels)
	hdType := g.hdType()
	profile := Profile{PersonalityLine: pers[0].Line, DesignLine: des[0].Line}

	if warnings == nil {
		warnings = []TimeWarning{}
	}
	if channels == nil {
		channels = []Channel{}
	}
	return &Chart{
		EngineVersion: e.EngineVersion,
		TZDBVersion:   e.TZDBVersion,
		Input:         req,
		BirthUTC:      birth,
		Warnings:      warnings,
		Personality:   ActivationSet{UTC: birth, Activations: pers},
		Design:        ActivationSet{UTC: jdToTime(designJD), Activations: des},
		Gates:         gates,
		Channels:      channels,
		Centers:       g.centers(),
		Type:          hdType,
		Authority:     g.authority(hdType),
		Profile:       profile,
		Definition:    g.definition(),
		IncarnationCross: IncarnationCross{
			Angle:            crossAngle(profile),
			PersonalitySun:   pers[0].Gate,
			PersonalityEarth: pers[1].Gate,
			DesignSun:        des[0].Gate,
			DesignEarth:      des[1].Gate,
		},
	}, nil
}

// activations returns all 13 bodies in ephemeris.Bodies order (Sun first,
// Earth second — the code above relies on that).
func (e *Engine) activations(jd float64) ([]Activation, error) {
	lon := map[ephemeris.Body]float64{}
	out := make([]Activation, 0, len(ephemeris.Bodies))
	for _, b := range ephemeris.Bodies {
		var l float64
		switch b {
		case ephemeris.Earth:
			l = normDeg(lon[ephemeris.Sun] + 180)
		case ephemeris.SouthNode:
			l = normDeg(lon[ephemeris.NorthNode] + 180)
		default:
			var err error
			if l, err = e.eph.Longitude(b, jd); err != nil {
				return nil, err
			}
		}
		lon[b] = l
		gate, line, color, tone, base := position(l)
		out = append(out, Activation{Body: b, Longitude: l, Gate: gate, Line: line, Color: color, Tone: tone, Base: base})
	}
	return out, nil
}

// designJD finds when the Sun stood designArc degrees before its birth
// position, by secant iteration from the ~88-day first guess.
func (e *Engine) designJD(birthJD, birthSun float64) (float64, error) {
	target := normDeg(birthSun - designArc)
	f := func(jd float64) (float64, error) {
		l, err := e.eph.Longitude(ephemeris.Sun, jd)
		return math.Remainder(l-target, 360), err
	}
	x0, x1 := birthJD-designArc/0.9856, birthJD-designArc/0.9856+0.5
	f0, err := f(x0)
	if err != nil {
		return 0, err
	}
	for range 30 {
		f1, err := f(x1)
		if err != nil {
			return 0, err
		}
		if math.Abs(f1) < 1e-9 || f1 == f0 {
			return x1, nil
		}
		x0, x1, f0 = x1, x1-f1*(x1-x0)/(f1-f0), f1
	}
	return x1, nil
}

func activeGates(pers, des []Activation) []ActiveGate {
	by := map[int]*ActiveGate{}
	add := func(side Side, acts []Activation) {
		for _, a := range acts {
			g := by[a.Gate]
			if g == nil {
				g = &ActiveGate{Gate: a.Gate}
				by[a.Gate] = g
			}
			g.Sources = append(g.Sources, GateSource{Side: side, Body: a.Body, Line: a.Line})
		}
	}
	add(Personality, pers)
	add(Design, des)
	out := make([]ActiveGate, 0, len(by))
	for _, g := range by {
		out = append(out, *g)
	}
	slices.SortFunc(out, func(a, b ActiveGate) int { return a.Gate - b.Gate })
	return out
}

func jdToTime(jd float64) time.Time {
	sec := (jd - 2440587.5) * 86400
	return time.Unix(0, int64(math.Round(sec))*1e9).UTC()
}
