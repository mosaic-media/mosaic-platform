package contracts

import (
	"context"

	"github.com/mosaic-media/mosaic-platform/internal/platform/domain"
)

// RelationStore persists the association graph (ADR 0013).
//
// Association does not nest, which is why it is a graph and not part of the
// containment tree. Computing a grouping is a background job writing rows
// here; reading it back is an indexed join on the same engine as everything
// else, with no second query path.
//
// There is no Update. Edges are written once with a confidence score, and
// ADR 0013 records that nothing yet ages or rechecks them — a changed
// assessment is a Delete and a Create, which keeps the absence of a decay
// policy visible rather than implied by a mutable score.
type RelationStore interface {
	// Create is Conflict on a duplicate (from, to, type) edge and
	// InvalidArgument when the two endpoints are the same node.
	Create(ctx context.Context, relation domain.Relation) (domain.Relation, error)
	FindByID(ctx context.Context, id domain.RelationID) (domain.Relation, error)

	// ListFrom returns outgoing edges, optionally narrowed to one type. An
	// empty relationType returns all of them.
	ListFrom(ctx context.Context, from domain.NodeID, relationType domain.RelationType) ([]domain.Relation, error)

	// ListTo returns incoming edges, optionally narrowed to one type. Both
	// directions are indexed: "what does this adapt" and "what adapts this"
	// are equally ordinary questions.
	ListTo(ctx context.Context, to domain.NodeID, relationType domain.RelationType) ([]domain.Relation, error)

	Delete(ctx context.Context, id domain.RelationID) error
}
