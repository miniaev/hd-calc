package hd

import (
	"time"

	"github.com/miniaev/hd-calc/internal/ephemeris"
)

// Types below mirror the schemas in api/openapi.yaml; JSON names must match.

type Side string

const (
	Personality Side = "personality"
	Design      Side = "design"
)

type ChartRequest struct {
	BirthLocal string   `json:"birth_local"`
	TZ         string   `json:"tz"`
	Lat        *float64 `json:"lat,omitempty"`
	Lon        *float64 `json:"lon,omitempty"`
}

type TimeWarning struct {
	Code  string `json:"code"` // ambiguous_local_time | nonexistent_local_time
	Local string `json:"local"`
}

type Activation struct {
	Body      ephemeris.Body `json:"body"`
	Longitude float64        `json:"longitude"`
	Gate      int            `json:"gate"`
	Line      int            `json:"line"`
	Color     int            `json:"color"`
	Tone      int            `json:"tone"`
	Base      int            `json:"base"`
}

type ActivationSet struct {
	UTC         time.Time    `json:"utc"`
	Activations []Activation `json:"activations"`
}

type GateSource struct {
	Side Side           `json:"side"`
	Body ephemeris.Body `json:"body"`
	Line int            `json:"line"`
}

type ActiveGate struct {
	Gate    int          `json:"gate"`
	Sources []GateSource `json:"sources"`
}

type Channel struct {
	ID      string    `json:"id"`
	Gates   [2]int    `json:"gates"`
	Centers [2]string `json:"centers"`
}

type Centers struct {
	Head        bool `json:"head"`
	Ajna        bool `json:"ajna"`
	Throat      bool `json:"throat"`
	G           bool `json:"g"`
	Heart       bool `json:"heart"`
	Sacral      bool `json:"sacral"`
	SolarPlexus bool `json:"solar_plexus"`
	Spleen      bool `json:"spleen"`
	Root        bool `json:"root"`
}

type Profile struct {
	PersonalityLine int `json:"personality_line"`
	DesignLine      int `json:"design_line"`
}

type IncarnationCross struct {
	Angle            string `json:"angle"` // right_angle | juxtaposition | left_angle
	PersonalitySun   int    `json:"personality_sun"`
	PersonalityEarth int    `json:"personality_earth"`
	DesignSun        int    `json:"design_sun"`
	DesignEarth      int    `json:"design_earth"`
}

type Chart struct {
	EngineVersion    string           `json:"engine_version"`
	TZDBVersion      string           `json:"tzdb_version"`
	Input            ChartRequest     `json:"input"`
	BirthUTC         time.Time        `json:"birth_utc"`
	Warnings         []TimeWarning    `json:"warnings"`
	Personality      ActivationSet    `json:"personality"`
	Design           ActivationSet    `json:"design"`
	Gates            []ActiveGate     `json:"gates"`
	Channels         []Channel        `json:"channels"`
	Centers          Centers          `json:"centers"`
	Type             string           `json:"type"`
	Authority        string           `json:"authority"`
	Profile          Profile          `json:"profile"`
	Definition       string           `json:"definition"`
	IncarnationCross IncarnationCross `json:"incarnation_cross"`
}
