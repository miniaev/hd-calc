package hd

import (
	"fmt"
	"slices"
)

// Composite: how 2–5 charts work together. Only gates matter here — a
// channel or center is defined for the group when the union of everyone's
// gates defines it.
//
// For a pair, each channel the two form gets the classic category:
//   - electromagnetic: each has one gate, neither has both — attraction;
//   - companionship: both have the whole channel — shared ground;
//   - dominance: one has it whole, the other has neither gate — one leads;
//   - compromise: one has it whole, the other one gate of it — friction.
//
// For 3–5 people the Penta applies: six channels between the Throat, the G
// and the Sacral carry the group's function, and missing ones are its gaps.

type CompositeRequest struct {
	Members []struct {
		Gates []int `json:"gates"`
	} `json:"members"`
}

type CompositeChannel struct {
	ID string `json:"id"`
	// Kind is set for pairs only.
	Kind string `json:"kind,omitempty"` // electromagnetic | companionship | dominance | compromise
	// Holders[i] lists the members (0-based) holding the i-th gate of the channel.
	Holders [2][]int `json:"holders"`
	// Whole lists the members who have the channel on their own.
	Whole []int `json:"whole"`
}

type GroupChart struct {
	Gates      []int    `json:"gates"`
	Centers    Centers  `json:"centers"`
	Type       string   `json:"type"`
	Authority  string   `json:"authority"`
	Definition string   `json:"definition"`
	Channels   []string `json:"channels"`
}

type PentaChannel struct {
	ID     string `json:"id"`
	Formed bool   `json:"formed"`
}

type Penta struct {
	Channels     []PentaChannel `json:"channels"`
	MissingGates []int          `json:"missing_gates"`
}

type MemberImpact struct {
	Member int `json:"member"`
	// What the group loses without this member: channels nobody else
	// completes, and centers that then turn open.
	LostChannels []string `json:"lost_channels"`
	LostCenters  []string `json:"lost_centers"`
}

type Composite struct {
	EngineVersion string             `json:"engine_version"`
	Members       int                `json:"members"`
	Group         GroupChart         `json:"group"`
	Channels      []CompositeChannel `json:"channels"`
	Penta         *Penta             `json:"penta,omitempty"`
	Impact        []MemberImpact     `json:"impact"`
}

// pentaChannels: Throat–G (upper) and G–Sacral (lower).
var pentaChannels = [][2]int{{7, 31}, {1, 8}, {13, 33}, {5, 15}, {2, 14}, {29, 46}}

var centerOrder = []string{Head, Ajna, Throat, GCenter, Heart, Sacral, SolarPlexus, Spleen, Root}

func (e *Engine) Composite(req CompositeRequest) (*Composite, error) {
	n := len(req.Members)
	if n < 2 || n > 5 {
		return nil, fmt.Errorf("%w: 2 to 5 members, got %d", ErrInvalidInput, n)
	}
	sets := make([]map[int]bool, n)
	for i, m := range req.Members {
		sets[i] = map[int]bool{}
		for _, g := range m.Gates {
			if g < 1 || g > 64 {
				return nil, fmt.Errorf("%w: member %d gate %d", ErrInvalidInput, i, g)
			}
			sets[i][g] = true
		}
	}
	all := union(sets, -1)
	chs := definedChannels(all)
	g := newGraph(chs)
	ty := g.hdType()
	out := &Composite{
		EngineVersion: e.EngineVersion,
		Members:       n,
		Group: GroupChart{Gates: sortedKeys(all), Centers: g.centers(), Type: ty, Authority: g.authority(ty),
			Definition: g.definition(), Channels: ids(chs)},
		Channels: []CompositeChannel{},
		Impact:   []MemberImpact{},
	}

	for _, ch := range chs {
		c := CompositeChannel{ID: ch.ID, Holders: [2][]int{{}, {}}, Whole: []int{}}
		for i, s := range sets {
			for k, gate := range ch.Gates {
				if s[gate] {
					c.Holders[k] = append(c.Holders[k], i)
				}
			}
			if s[ch.Gates[0]] && s[ch.Gates[1]] {
				c.Whole = append(c.Whole, i)
			}
		}
		if n == 2 {
			c.Kind = pairKind(sets[0], sets[1], ch.Gates)
		}
		out.Channels = append(out.Channels, c)
	}

	if n >= 3 {
		p := &Penta{Channels: []PentaChannel{}, MissingGates: []int{}}
		for _, pc := range pentaChannels {
			p.Channels = append(p.Channels, PentaChannel{ID: channelID(pc), Formed: all[pc[0]] && all[pc[1]]})
			for _, gate := range pc {
				if !all[gate] {
					p.MissingGates = append(p.MissingGates, gate)
				}
			}
		}
		slices.Sort(p.MissingGates)
		out.Penta = p
	}

	for i := range n {
		rest := union(sets, i)
		restChs := definedChannels(rest)
		have := map[string]bool{}
		for _, c := range restChs {
			have[c.ID] = true
		}
		im := MemberImpact{Member: i, LostChannels: []string{}, LostCenters: []string{}}
		for _, c := range chs {
			if !have[c.ID] {
				im.LostChannels = append(im.LostChannels, c.ID)
			}
		}
		rg := newGraph(restChs)
		for _, c := range centerOrder {
			if g.defined(c) && !rg.defined(c) {
				im.LostCenters = append(im.LostCenters, c)
			}
		}
		out.Impact = append(out.Impact, im)
	}
	return out, nil
}

func pairKind(a, b map[int]bool, gates [2]int) string {
	count := func(s map[int]bool) int {
		n := 0
		for _, g := range gates {
			if s[g] {
				n++
			}
		}
		return n
	}
	ca, cb := count(a), count(b)
	switch {
	case ca == 2 && cb == 2:
		return "companionship"
	case ca == 2 && cb == 0, cb == 2 && ca == 0:
		return "dominance"
	case ca == 2 || cb == 2:
		return "compromise"
	default: // one gate each, different gates — the union defines the channel
		return "electromagnetic"
	}
}

// union merges all sets except the one at skip (-1 to keep all).
func union(sets []map[int]bool, skip int) map[int]bool {
	u := map[int]bool{}
	for i, s := range sets {
		if i == skip {
			continue
		}
		for g := range s {
			u[g] = true
		}
	}
	return u
}

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func ids(chs []Channel) []string {
	out := make([]string, 0, len(chs))
	for _, c := range chs {
		out = append(out, c.ID)
	}
	return out
}
