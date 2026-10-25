package hd

import (
	"fmt"
	"slices"
)

// Center keys match the API contract (schema Center).
const (
	Head        = "head"
	Ajna        = "ajna"
	Throat      = "throat"
	GCenter     = "g"
	Heart       = "heart"
	Sacral      = "sacral"
	SolarPlexus = "solar_plexus"
	Spleen      = "spleen"
	Root        = "root"
)

var centerGates = map[string][]int{
	Head:        {64, 61, 63},
	Ajna:        {47, 24, 4, 17, 43, 11},
	Throat:      {62, 23, 56, 35, 12, 45, 33, 8, 31, 20, 16},
	GCenter:     {7, 1, 13, 25, 46, 2, 15, 10},
	Heart:       {21, 40, 26, 51},
	Sacral:      {5, 14, 29, 59, 9, 3, 42, 27, 34},
	SolarPlexus: {6, 37, 22, 36, 30, 55, 49},
	Spleen:      {48, 57, 44, 50, 32, 28, 18},
	Root:        {53, 60, 52, 19, 39, 41, 58, 38, 54},
}

// gateCenter is the inverse of centerGates.
var gateCenter = func() map[int]string {
	m := map[int]string{}
	for c, gs := range centerGates {
		for _, g := range gs {
			m[g] = c
		}
	}
	return m
}()

// channelGates lists the 36 channels, lower gate first.
var channelGates = [][2]int{
	{1, 8}, {2, 14}, {3, 60}, {4, 63}, {5, 15}, {6, 59}, {7, 31}, {9, 52},
	{10, 20}, {10, 34}, {10, 57}, {11, 56}, {12, 22}, {13, 33}, {16, 48}, {17, 62},
	{18, 58}, {19, 49}, {20, 34}, {20, 57}, {21, 45}, {23, 43}, {24, 61}, {25, 51},
	{26, 44}, {27, 50}, {28, 38}, {29, 46}, {30, 41}, {32, 54}, {34, 57}, {35, 36},
	{37, 40}, {39, 55}, {42, 53}, {47, 64},
}

func channelID(g [2]int) string { return fmt.Sprintf("%d-%d", g[0], g[1]) }

// definedChannels returns the channels whose both gates are active, in
// channelGates order.
func definedChannels(active map[int]bool) []Channel {
	var out []Channel
	for _, g := range channelGates {
		if active[g[0]] && active[g[1]] {
			out = append(out, Channel{
				ID:      channelID(g),
				Gates:   g,
				Centers: [2]string{gateCenter[g[0]], gateCenter[g[1]]},
			})
		}
	}
	return out
}

// graph is the bodygraph as an undirected graph of defined centers.
type graph map[string][]string

func newGraph(chs []Channel) graph {
	g := graph{}
	for _, ch := range chs {
		a, b := ch.Centers[0], ch.Centers[1]
		g[a] = append(g[a], b)
		g[b] = append(g[b], a)
	}
	return g
}

func (g graph) defined(c string) bool { return len(g[c]) > 0 }

// connected reports whether a path of defined channels joins a and b.
func (g graph) connected(a, b string) bool {
	if !g.defined(a) || !g.defined(b) {
		return false
	}
	seen := map[string]bool{a: true}
	queue := []string{a}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == b {
			return true
		}
		for _, n := range g[c] {
			if !seen[n] {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	return false
}

// components counts separate groups of connected defined centers.
func (g graph) components() int {
	seen := map[string]bool{}
	n := 0
	for c := range g {
		if seen[c] {
			continue
		}
		n++
		stack := []string{c}
		seen[c] = true
		for len(stack) > 0 {
			x := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, y := range g[x] {
				if !seen[y] {
					seen[y] = true
					stack = append(stack, y)
				}
			}
		}
	}
	return n
}

func (g graph) centers() Centers {
	return Centers{
		Head: g.defined(Head), Ajna: g.defined(Ajna), Throat: g.defined(Throat),
		G: g.defined(GCenter), Heart: g.defined(Heart), Sacral: g.defined(Sacral),
		SolarPlexus: g.defined(SolarPlexus), Spleen: g.defined(Spleen), Root: g.defined(Root),
	}
}

func (g graph) hdType() string {
	if len(g) == 0 {
		return "reflector"
	}
	motorToThroat := slices.ContainsFunc([]string{Heart, SolarPlexus, Root, Sacral},
		func(m string) bool { return g.connected(m, Throat) })
	switch {
	case g.defined(Sacral) && motorToThroat:
		return "manifesting_generator"
	case g.defined(Sacral):
		return "generator"
	case motorToThroat:
		return "manifestor"
	default:
		return "projector"
	}
}

// authority follows the standard hierarchy: Solar Plexus, Sacral, Spleen,
// Ego, G (self-projected), then mental for Projectors and lunar for Reflectors.
func (g graph) authority(hdType string) string {
	switch {
	case hdType == "reflector":
		return "lunar"
	case g.defined(SolarPlexus):
		return "emotional"
	case g.defined(Sacral):
		return "sacral"
	case g.defined(Spleen):
		return "splenic"
	case g.defined(Heart) && hdType == "manifestor":
		return "ego_manifested"
	case g.defined(Heart):
		return "ego_projected"
	case g.defined(GCenter):
		return "self_projected"
	default:
		return "mental"
	}
}

func (g graph) definition() string {
	return [...]string{"none", "single", "split", "triple_split", "quadruple_split"}[min(g.components(), 4)]
}

func crossAngle(p Profile) string {
	switch {
	case p.PersonalityLine == 4 && p.DesignLine == 1:
		return "juxtaposition"
	case p.PersonalityLine >= 5:
		return "left_angle"
	default:
		return "right_angle"
	}
}
