package app

import v1 "github.com/mosaic-media/mosaic-sdk/contracts/platform/v1"

// CapabilityRegistry holds the optional-module capabilities the composition
// root registered, keyed by manifest id. The Platform routes an ImportContent
// command to one of them by id. It is populated once, at composition, and read
// at invocation — there is no runtime registration path (ADR 0007: modules are
// selected before the build, not discovered at runtime).
//
// It lives in the app package rather than under composition/ so the Service
// can hold it without an import cycle: it depends only on the published SDK,
// exactly as a module does.
type CapabilityRegistry struct {
	byID map[string]v1.Capability
}

// NewCapabilityRegistry returns an empty registry.
func NewCapabilityRegistry() *CapabilityRegistry {
	return &CapabilityRegistry{byID: make(map[string]v1.Capability)}
}

// Register adds a capability under its manifest id. Registration order is the
// composition root's, and a repeated id replaces the earlier registration —
// the composition root controls both, so this stays a plain assignment rather
// than an error path.
func (r *CapabilityRegistry) Register(c v1.Capability) {
	r.byID[c.Manifest().ID] = c
}

// Lookup returns the capability registered under id, and whether one was.
func (r *CapabilityRegistry) Lookup(id string) (v1.Capability, bool) {
	c, ok := r.byID[id]
	return c, ok
}

// Manifests returns the manifest of every registered capability, so the
// composition root can report what it wired.
func (r *CapabilityRegistry) Manifests() []v1.Manifest {
	manifests := make([]v1.Manifest, 0, len(r.byID))
	for _, c := range r.byID {
		manifests = append(manifests, c.Manifest())
	}
	return manifests
}
