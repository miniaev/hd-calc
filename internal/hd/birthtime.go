package hd

import (
	"fmt"
	"time"
	_ "time/tzdata" // fallback when the host has no zoneinfo
)

// resolveLocal turns a wall-clock time in an IANA zone into a UTC instant,
// with explicit rules for clock changes (Go's time.Date leaves the choice
// unspecified):
//   - ambiguous (clocks went back): the earlier of the two instants;
//   - nonexistent (clocks jumped forward): shifted forward by the jump,
//     i.e. read with the offset in force before the change.
func resolveLocal(local, tz string) (time.Time, []TimeWarning, error) {
	return resolveLocalPick(local, tz, false)
}

// resolveLocalPick is resolveLocal with a choice for the ambiguous case:
// late=true takes the later instant (used for the end of a range, so a
// doubled hour widens the interval instead of shrinking it).
func resolveLocalPick(local, tz string, late bool) (time.Time, []TimeWarning, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" || tz == "Local" {
		return time.Time{}, nil, fmt.Errorf("%w: %q", ErrUnknownTimezone, tz)
	}
	var naive time.Time
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05"} {
		if naive, err = time.Parse(layout, local); err == nil {
			break
		}
	}
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("%w: birth_local %q: want YYYY-MM-DDTHH:MM[:SS]", ErrInvalidInput, local)
	}

	// Candidate offsets: whatever was in force a day and a half either side.
	offsetAt := func(t time.Time) int { _, o := t.In(loc).Zone(); return o }
	before := offsetAt(naive.Add(-36 * time.Hour))
	after := offsetAt(naive.Add(36 * time.Hour))

	var valid []time.Time
	for _, o := range uniq(before, after, offsetAt(naive)) {
		u := naive.Add(-time.Duration(o) * time.Second)
		if offsetAt(u) == o {
			valid = append(valid, u)
		}
	}
	switch len(valid) {
	case 0:
		u := naive.Add(-time.Duration(before) * time.Second)
		return u, []TimeWarning{{Code: "nonexistent_local_time", Local: local}}, nil
	case 1:
		return valid[0], nil, nil
	default:
		pick := valid[0]
		for _, v := range valid[1:] {
			if v.Before(pick) != late {
				pick = v
			}
		}
		return pick, []TimeWarning{{Code: "ambiguous_local_time", Local: local}}, nil
	}
}

func uniq(xs ...int) []int {
	var out []int
	for _, x := range xs {
		found := false
		for _, y := range out {
			found = found || x == y
		}
		if !found {
			out = append(out, x)
		}
	}
	return out
}
