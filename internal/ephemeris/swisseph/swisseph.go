// Package swisseph implements ephemeris.Ephemeris on top of the Swiss
// Ephemeris by Astrodienst AG, used under the AGPL-3.0.
//
// Skeleton: the C sources and .se1 data files are not vendored yet, so every
// call returns ErrNotImplemented. Next step: vendor the C library under
// third_party/swisseph with its licence notice, bind swe_calc_ut via cgo and
// point swe_set_ephe_path at the bundled data files.
package swisseph

import (
	"errors"

	"github.com/miniaev/hd-calc/internal/ephemeris"
)

const (
	Name    = "swisseph"
	Version = "2.10.03"
)

// ErrNotImplemented is returned until the cgo binding lands.
var ErrNotImplemented = errors.New("swisseph: not implemented yet")

type SwissEph struct {
	// DataPath is the directory with the .se1 ephemeris files.
	DataPath string
}

var _ ephemeris.Ephemeris = (*SwissEph)(nil)

func New(dataPath string) *SwissEph { return &SwissEph{DataPath: dataPath} }

func (s *SwissEph) Longitude(body ephemeris.Body, jdUT float64) (float64, error) {
	if body.Derived() {
		return 0, ephemeris.ErrUnsupportedBody
	}
	return 0, ErrNotImplemented
}
