// Package swisseph implements ephemeris.Ephemeris on top of the Swiss
// Ephemeris by Astrodienst AG, used under the AGPL-3.0.
//
// The C library is vendored unmodified in third_party/swisseph; the se_*.c
// files here just compile each of its sources via cgo. Data files (.se1) are
// read from DataPath. If they are missing the library would silently fall
// back to its lower-precision Moshier model — we treat that as an error so
// results never depend on what happens to be on disk.
//
// TLSOFF: the library keeps its state (ephemeris path, open files) in
// globals. With thread-local storage each OS thread — and goroutines hop
// between them — would see its own uninitialised copy, so TLS is off and a
// mutex serialises calls instead.
package swisseph

/*
#cgo CFLAGS: -I${SRCDIR}/../../../third_party/swisseph -DTLSOFF -O2 -w
#cgo LDFLAGS: -lm
#include <stdlib.h>
#include "swephexp.h"
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/miniaev/hd-calc/internal/ephemeris"
)

const (
	Name    = "swisseph"
	Version = "2.10.03"
)

// Dates covered by sepl_18.se1 / semo_18.se1 (1800-01-01 .. 2400-01-01).
const (
	minJD = 2378496.5
	maxJD = 2597641.5
)

var bodyID = map[ephemeris.Body]C.int32{
	ephemeris.Sun:     C.SE_SUN,
	ephemeris.Moon:    C.SE_MOON,
	ephemeris.Mercury: C.SE_MERCURY,
	ephemeris.Venus:   C.SE_VENUS,
	ephemeris.Mars:    C.SE_MARS,
	ephemeris.Jupiter: C.SE_JUPITER,
	ephemeris.Saturn:  C.SE_SATURN,
	ephemeris.Uranus:  C.SE_URANUS,
	ephemeris.Neptune: C.SE_NEPTUNE,
	ephemeris.Pluto:   C.SE_PLUTO,
}

// mu guards the library's global state (see package doc).
var mu sync.Mutex

type SwissEph struct {
	// DataPath is the directory with the .se1 ephemeris files.
	DataPath string
	// MeanNode switches the lunar node from true (default) to mean.
	MeanNode bool
}

var _ ephemeris.Ephemeris = (*SwissEph)(nil)

func New(dataPath string) *SwissEph { return &SwissEph{DataPath: dataPath} }

func (s *SwissEph) Longitude(body ephemeris.Body, jdUT float64) (float64, error) {
	ipl, ok := bodyID[body]
	if body == ephemeris.NorthNode {
		ipl, ok = C.SE_TRUE_NODE, true
		if s.MeanNode {
			ipl = C.SE_MEAN_NODE
		}
	}
	if !ok {
		return 0, ephemeris.ErrUnsupportedBody
	}
	if jdUT < minJD || jdUT >= maxJD {
		return 0, ephemeris.ErrOutOfRange
	}

	mu.Lock()
	defer mu.Unlock()

	path := C.CString(s.DataPath)
	defer C.free(unsafe.Pointer(path))
	C.swe_set_ephe_path(path) // cheap when unchanged; keeps the path right after any reset

	var xx [6]C.double
	var serr [C.AS_MAXCH]C.char
	ret := C.swe_calc_ut(C.double(jdUT), ipl, C.SEFLG_SWIEPH, &xx[0], &serr[0])
	if ret < 0 {
		return 0, fmt.Errorf("swisseph: %s", C.GoString(&serr[0]))
	}
	if ret&C.SEFLG_SWIEPH == 0 {
		return 0, fmt.Errorf("swisseph: ephemeris files not found in %q, refusing Moshier fallback: %s",
			s.DataPath, C.GoString(&serr[0]))
	}
	return float64(xx[0]), nil
}
