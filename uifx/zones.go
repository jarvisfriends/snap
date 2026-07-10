package uifx

import (
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/geom"
)

// Zones is a set of named, content-relative hit zones built from lipgloss
// layers — the same primitives the View can render. A component builds its
// frame from positioned blocks, registers the interactive ones here by ID
// during View, and its OnMouse handlers ask which zone a pointer event
// landed in. Because each zone's position and size come from the very block
// the View rendered, the zones stay correct under responsive layouts,
// wrapped content, and overlapping layers (top-most z wins) — no
// hand-maintained rectangles to drift out of sync.
type Zones struct {
	comp *lipgloss.Compositor
}

// NewZones builds a zone set from lipgloss layers
// (lipgloss.NewLayer(content).ID(name).X(x).Y(y) in the component's content
// coordinates). Layers without an ID never match a hit; nested layers
// resolve relative to their parent.
func NewZones(layers ...*lipgloss.Layer) *Zones {
	return &Zones{comp: lipgloss.NewCompositor(layers...)}
}

// Hit returns the ID of the top-most zone containing (x, y), or "" when the
// point is outside every zone. A nil Zones (no View rendered yet) misses.
func (z *Zones) Hit(x, y int) string {
	if z == nil {
		return ""
	}
	return z.comp.Hit(x, y).ID()
}

// Bounds returns the content-relative bounds of the named zone. It reports
// flat (unnested) zones only — a nested layer's coordinates are relative to
// its parent. Mainly for tests aiming events at a zone.
func (z *Zones) Bounds(id string) (geom.Rect, bool) {
	if z == nil {
		return geom.Rect{}, false
	}
	l := z.comp.GetLayer(id)
	if l == nil {
		return geom.Rect{}, false
	}
	return geom.Rect{X: l.GetX(), Y: l.GetY(), W: l.Width(), H: l.Height()}, true
}
