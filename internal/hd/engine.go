// Package hd turns planetary longitudes into a Human Design chart. It knows
// nothing about where longitudes come from — only the ephemeris.Ephemeris
// interface.
package hd

import (
	"errors"

	"github.com/miniaev/hd-calc/internal/ephemeris"
)

// CalcVersion is the semver of the calculation code. Bump it on any change
// that can alter a result; together with the ephemeris id it forms
// engine_version.
const CalcVersion = "0.1.0"

var ErrNotImplemented = errors.New("hd: not implemented yet")

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
	return nil, ErrNotImplemented
}
