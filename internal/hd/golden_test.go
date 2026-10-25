package hd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/miniaev/hd-calc/internal/ephemeris/swisseph"
)

// Golden fixtures live in testdata/golden/*.json, one chart per file, each
// cross-checked against 2–3 public calculators. Every engine — Swiss
// Ephemeris today, anything else tomorrow — must pass the same set.
//
// expect is deliberately compact (gate.line per body, keys as strings): it is
// what a person can read off another calculator's screen and type in.
// Absent fields are not checked, so a fixture may pin only what it is about.

const goldenDir = "../../testdata/golden"

type goldenFixture struct {
	Name         string       `json:"name"`
	Tags         []string     `json:"tags"`
	VerifiedWith []string     `json:"verified_with"`
	Note         string       `json:"note"`
	Request      ChartRequest `json:"request"`
	Expect       goldenExpect `json:"expect"`
}

type goldenExpect struct {
	BirthUTC         string            `json:"birth_utc"`
	Type             string            `json:"type"`
	Authority        string            `json:"authority"`
	Profile          string            `json:"profile"`
	Definition       string            `json:"definition"`
	IncarnationCross string            `json:"incarnation_cross"`
	DefinedCenters   []string          `json:"defined_centers"`
	Channels         []string          `json:"channels"`
	Personality      map[string]string `json:"personality"`
	Design           map[string]string `json:"design"`
}

func loadFixtures(t *testing.T, includeTemplates bool) map[string]goldenFixture {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(goldenDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]goldenFixture{}
	for _, p := range paths {
		base := filepath.Base(p)
		if strings.HasPrefix(base, "_") != includeTemplates {
			continue
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var f goldenFixture
		dec := json.NewDecoder(strings.NewReader(string(raw)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&f); err != nil {
			t.Fatalf("%s: %v", base, err)
		}
		out[base] = f
	}
	return out
}

// TestGoldenTemplateParses keeps the fixture format and the template in sync.
func TestGoldenTemplateParses(t *testing.T) {
	if len(loadFixtures(t, true)) == 0 {
		t.Fatal("testdata/golden/_template.json is missing")
	}
}

func TestGolden(t *testing.T) {
	fixtures := loadFixtures(t, false)
	if len(fixtures) == 0 {
		t.Skip("no golden fixtures yet — add testdata/golden/*.json")
	}
	eng := NewEngine(swisseph.New("../../ephe"), swisseph.Name+"-"+swisseph.Version, "")
	for name, f := range fixtures {
		t.Run(name, func(t *testing.T) {
			c, err := eng.Chart(f.Request)
			if err != nil {
				t.Fatalf("Chart: %v", err)
			}
			for _, d := range diffChart(c, f.Expect) {
				t.Error(d)
			}
		})
	}
}

// diffChart compares only the fields the fixture sets.
func diffChart(c *Chart, e goldenExpect) []string {
	var d []string
	eq := func(field, got, want string) {
		if want != "" && got != want {
			d = append(d, fmt.Sprintf("%s: got %q, want %q", field, got, want))
		}
	}
	eq("birth_utc", c.BirthUTC.UTC().Format("2006-01-02T15:04:05Z"), e.BirthUTC)
	eq("type", c.Type, e.Type)
	eq("authority", c.Authority, e.Authority)
	eq("profile", fmt.Sprintf("%d/%d", c.Profile.PersonalityLine, c.Profile.DesignLine), e.Profile)
	eq("definition", c.Definition, e.Definition)
	x := c.IncarnationCross
	eq("incarnation_cross", fmt.Sprintf("%s:%d/%d|%d/%d", x.Angle,
		x.PersonalitySun, x.PersonalityEarth, x.DesignSun, x.DesignEarth), e.IncarnationCross)

	if e.DefinedCenters != nil {
		eq("defined_centers", strings.Join(definedCenters(c.Centers), ","), joinSorted(e.DefinedCenters))
	}
	if e.Channels != nil {
		var got []string
		for _, ch := range c.Channels {
			got = append(got, ch.ID)
		}
		eq("channels", joinSorted(got), joinSorted(e.Channels))
	}
	for side, want := range map[string]map[string]string{"personality": e.Personality, "design": e.Design} {
		set := c.Personality
		if side == "design" {
			set = c.Design
		}
		for _, a := range set.Activations {
			eq(side+"."+string(a.Body), fmt.Sprintf("%d.%d", a.Gate, a.Line), want[string(a.Body)])
		}
	}
	return d
}

func definedCenters(c Centers) []string {
	var out []string
	for name, on := range map[string]bool{
		"head": c.Head, "ajna": c.Ajna, "throat": c.Throat, "g": c.G, "heart": c.Heart,
		"sacral": c.Sacral, "solar_plexus": c.SolarPlexus, "spleen": c.Spleen, "root": c.Root,
	} {
		if on {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return out
}

func joinSorted(s []string) string {
	s = slices.Clone(s)
	slices.Sort(s)
	return strings.Join(s, ",")
}
