// Package ephemeris hides where planetary positions come from. Everything
// above it (design moment, gates, channels, type…) talks only to the
// Ephemeris interface, so swapping Swiss Ephemeris for a commercial licence
// or our own implementation means writing one new type, nothing else.
package ephemeris

import (
	"errors"
	"time"
)

// Body is a point the chart is built from. String values match the API
// contract (api/openapi.yaml, schema Body).
type Body string

const (
	Sun       Body = "sun"
	Earth     Body = "earth"
	NorthNode Body = "north_node"
	SouthNode Body = "south_node"
	Moon      Body = "moon"
	Mercury   Body = "mercury"
	Venus     Body = "venus"
	Mars      Body = "mars"
	Jupiter   Body = "jupiter"
	Saturn    Body = "saturn"
	Uranus    Body = "uranus"
	Neptune   Body = "neptune"
	Pluto     Body = "pluto"
)

// Bodies lists all 13 bodies in the order the API returns activations.
var Bodies = []Body{
	Sun, Earth, NorthNode, SouthNode, Moon, Mercury, Venus,
	Mars, Jupiter, Saturn, Uranus, Neptune, Pluto,
}

// Derived reports whether the engine computes the body from another one
// (Earth = Sun + 180°, South Node = North Node + 180°). An Ephemeris is never
// asked for derived bodies.
func (b Body) Derived() bool { return b == Earth || b == SouthNode }

// Ephemeris returns the geocentric tropical ecliptic longitude of body, in
// degrees [0, 360), at Julian Day jdUT (Universal Time).
type Ephemeris interface {
	Longitude(body Body, jdUT float64) (float64, error)
}

var (
	// ErrOutOfRange: the moment is outside the dates the ephemeris covers.
	ErrOutOfRange = errors.New("ephemeris: date out of range")
	// ErrUnsupportedBody: the implementation cannot compute this body.
	ErrUnsupportedBody = errors.New("ephemeris: unsupported body")
)

// unixEpochJD is the Julian Day of 1970-01-01T00:00:00Z.
const unixEpochJD = 2440587.5

// JulianDayUT converts t to a Julian Day in Universal Time.
func JulianDayUT(t time.Time) float64 {
	return unixEpochJD + float64(t.UnixNano())/(86400*1e9)
}
