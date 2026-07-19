package contracts

import (
	"context"

	"github.com/mosaic-media/mosaic-platform/internal/platform/domain"
)

// NodeStore persists the containment tree (ADR 0013).
//
// Every traversal here is by parent, never by an assumed level: variable
// depth is the property that lets a film be Work → Item and a series be
// Work → Container → Item without either being a special case, and it costs
// the discipline of never assuming a node has a parent or that a work's
// children are containers.
//
// Implementations must store the open type vocabularies canonically —
// domain.Node.Canonical() — so that "Anime Series", "anime-series" and
// "anime_series" are one media type rather than three (ADR 0015). Writes
// return the canonical value, which may therefore differ from what was
// passed in.
type NodeStore interface {
	Create(ctx context.Context, node domain.Node) (domain.Node, error)
	FindByID(ctx context.Context, id domain.NodeID) (domain.Node, error)
	Update(ctx context.Context, node domain.Node) (domain.Node, error)

	// ListChildren returns the direct children of a node ordered by
	// NaturalOrder. This is the single most common query a media browser
	// makes and it is served by a plain indexed scan — no recursion at read
	// time.
	ListChildren(ctx context.Context, parentID domain.NodeID) ([]domain.Node, error)

	// ListByWork returns every node in one work's tree, the work itself
	// included, ordered by NaturalOrder. It reads the denormalised work id
	// rather than walking parents.
	ListByWork(ctx context.Context, workID domain.NodeID) ([]domain.Node, error)

	// ListWorks returns the root of every tree — the nodes with no parent —
	// optionally narrowed to one media type. An empty mediaType returns all
	// of them.
	ListWorks(ctx context.Context, mediaType domain.MediaType) ([]domain.Node, error)

	// Delete removes one node. It is Conflict when the node still has
	// children or parts: ADR 0013 rules that deletion is a decision a user
	// confirms, never a silent cascade, so the store refuses rather than
	// taking a subtree with it. Callers delete depth-first.
	Delete(ctx context.Context, id domain.NodeID) error
}
