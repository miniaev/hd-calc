package hd

import (
	"fmt"
	"maps"
	"slices"
	"time"
)

type ChartRangeRequest struct {
	BirthLocalFrom string   `json:"birth_local_from"`
	BirthLocalTo   string   `json:"birth_local_to"`
	TZ             string   `json:"tz"`
	Lat            *float64 `json:"lat,omitempty"`
	Lon            *float64 `json:"lon,omitempty"`
	StepMinutes    int      `json:"step_minutes,omitempty"`
}

type Segment struct {
	Value     string    `json:"value"`
	FromUTC   time.Time `json:"from_utc"`
	ToUTC     time.Time `json:"to_utc"`
	FromLocal string    `json:"from_local"`
	ToLocal   string    `json:"to_local"`
}

type ElementVariation struct {
	Element  string    `json:"element"`
	Segments []Segment `json:"segments"`
}

type ChartRange struct {
	EngineVersion string             `json:"engine_version"`
	TZDBVersion   string             `json:"tzdb_version"`
	FromUTC       time.Time          `json:"from_utc"`
	ToUTC         time.Time          `json:"to_utc"`
	StepMinutes   int                `json:"step_minutes"`
	Varying       []ElementVariation `json:"varying"`
	Warnings      []TimeWarning      `json:"warnings"`
}

const maxRange = 24 * time.Hour

// ChartRange reports which chart elements change inside an interval of
// possible birth times. Charts are sampled every step minutes, and every
// step whose ends differ is bisected down to one minute, so a change point
// is exact to the minute (a change that reverts within one step is missed;
// at the default 5 minutes nothing in the chart moves that fast).
func (e *Engine) ChartRange(req ChartRangeRequest) (*ChartRange, error) {
	step := req.StepMinutes
	if step == 0 {
		step = 5
	}
	if step < 1 || step > 60 {
		return nil, fmt.Errorf("%w: step_minutes must be 1..60", ErrInvalidInput)
	}
	from, w1, err := resolveLocalPick(req.BirthLocalFrom, req.TZ, false)
	if err != nil {
		return nil, err
	}
	to, w2, err := resolveLocalPick(req.BirthLocalTo, req.TZ, true)
	if err != nil {
		return nil, err
	}
	if to.Before(from) {
		return nil, fmt.Errorf("%w: birth_local_to is before birth_local_from", ErrInvalidInput)
	}
	if to.Sub(from) > maxRange {
		return nil, fmt.Errorf("%w: %s > 24h", ErrIntervalTooLong, to.Sub(from))
	}
	loc, _ := time.LoadLocation(req.TZ) // already validated by resolveLocalPick

	// Elements at minute offsets from `from`.
	at := map[int]map[string]string{}
	eval := func(m int) (map[string]string, error) {
		if el, ok := at[m]; ok {
			return el, nil
		}
		c, err := e.chartAt(ChartRequest{TZ: req.TZ, Lat: req.Lat, Lon: req.Lon}, from.Add(time.Duration(m)*time.Minute), nil)
		if err != nil {
			return nil, err
		}
		el := elements(c)
		at[m] = el
		return el, nil
	}
	var refine func(a, b int) error
	refine = func(a, b int) error {
		if b-a <= 1 {
			return nil
		}
		ea, _ := eval(a)
		eb, _ := eval(b)
		if maps.Equal(ea, eb) {
			return nil
		}
		mid := (a + b) / 2
		if _, err := eval(mid); err != nil {
			return err
		}
		if err := refine(a, mid); err != nil {
			return err
		}
		return refine(mid, b)
	}

	total := int(to.Sub(from) / time.Minute)
	marks := []int{}
	for m := 0; m < total; m += step {
		marks = append(marks, m)
	}
	marks = append(marks, total)
	for _, m := range marks {
		if _, err := eval(m); err != nil {
			return nil, err
		}
	}
	for i := 1; i < len(marks); i++ {
		if err := refine(marks[i-1], marks[i]); err != nil {
			return nil, err
		}
	}

	// Walk evaluated minutes in order; a value seen first at minute m starts
	// a segment at m (the previous evaluated minute is m-1 at a change).
	minutes := slices.Sorted(maps.Keys(at))
	keys := slices.Sorted(maps.Keys(at[0]))
	local := func(t time.Time) string { return t.In(loc).Format("2006-01-02T15:04") }
	tm := func(m int) time.Time { return from.Add(time.Duration(m) * time.Minute) }
	end := to
	varying := []ElementVariation{}
	for _, k := range keys {
		var segs []Segment
		start, val := 0, at[0][k]
		for _, m := range minutes[1:] {
			if v := at[m][k]; v != val {
				segs = append(segs, Segment{Value: val, FromUTC: tm(start), ToUTC: tm(m), FromLocal: local(tm(start)), ToLocal: local(tm(m))})
				start, val = m, v
			}
		}
		if len(segs) == 0 {
			continue
		}
		segs = append(segs, Segment{Value: val, FromUTC: tm(start), ToUTC: end, FromLocal: local(tm(start)), ToLocal: local(end)})
		varying = append(varying, ElementVariation{Element: k, Segments: segs})
	}

	warnings := append(append([]TimeWarning{}, w1...), w2...)
	return &ChartRange{
		EngineVersion: e.EngineVersion,
		TZDBVersion:   e.TZDBVersion,
		FromUTC:       from,
		ToUTC:         to,
		StepMinutes:   step,
		Varying:       varying,
		Warnings:      warnings,
	}, nil
}

// elements flattens a chart into the element keys of ElementVariation.
func elements(c *Chart) map[string]string {
	el := map[string]string{
		"type":       c.Type,
		"authority":  c.Authority,
		"profile":    fmt.Sprintf("%d/%d", c.Profile.PersonalityLine, c.Profile.DesignLine),
		"definition": c.Definition,
		"incarnation_cross": fmt.Sprintf("%s:%d/%d|%d/%d", c.IncarnationCross.Angle,
			c.IncarnationCross.PersonalitySun, c.IncarnationCross.PersonalityEarth,
			c.IncarnationCross.DesignSun, c.IncarnationCross.DesignEarth),
	}
	onOff := func(b bool, on, off string) string {
		if b {
			return on
		}
		return off
	}
	cs := c.Centers
	for name, v := range map[string]bool{Head: cs.Head, Ajna: cs.Ajna, Throat: cs.Throat, GCenter: cs.G,
		Heart: cs.Heart, Sacral: cs.Sacral, SolarPlexus: cs.SolarPlexus, Spleen: cs.Spleen, Root: cs.Root} {
		el["center."+name] = onOff(v, "defined", "open")
	}
	defined := map[string]bool{}
	for _, ch := range c.Channels {
		defined[ch.ID] = true
	}
	for _, g := range channelGates {
		id := channelID(g)
		el["channel."+id] = onOff(defined[id], "defined", "undefined")
	}
	for _, set := range []struct {
		side Side
		acts []Activation
	}{{Personality, c.Personality.Activations}, {Design, c.Design.Activations}} {
		for _, a := range set.acts {
			el[fmt.Sprintf("activation.%s.%s", set.side, a.Body)] = fmt.Sprintf("%d.%d", a.Gate, a.Line)
		}
	}
	return el
}
