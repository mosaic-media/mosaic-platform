package contracts

import (
	"context"

	"github.com/mosaic-media/mosaic-platform/internal/platform/domain"
)

// SourceBindingStore persists identity resolution (ADR 0013).
//
// The point of the table is that resolution is explicit and inspectable. A
// weak match is a row a user can see and act on, not a silent merge of two
// works that share a title.
type SourceBindingStore interface {
	// Create is Conflict when the (provider, ref) pair is already bound —
	// one source binds to at most one node.
	Create(ctx context.Context, binding domain.SourceBinding) (domain.SourceBinding, error)
	FindByID(ctx context.Context, id domain.SourceBindingID) (domain.SourceBinding, error)

	// FindBySource looks a binding up by where it came from, which is how a
	// rescan discovers that a source is already resolved.
	FindBySource(ctx context.Context, provider, ref string) (domain.SourceBinding, error)

	// Update persists a status change or a move to a different node. A
	// confirmation and a split are both this call.
	Update(ctx context.Context, binding domain.SourceBinding) (domain.SourceBinding, error)

	// ListByNode returns every binding behind a node. A node whose list is
	// empty is orphaned, not deleted.
	ListByNode(ctx context.Context, nodeID domain.NodeID) ([]domain.SourceBinding, error)

	// ListPendingReview returns the bindings waiting on a person, oldest
	// first. Identity resolution becoming visible means the Platform needs
	// a surface for a user to act on, and this is the read behind it.
	ListPendingReview(ctx context.Context, limit int) ([]domain.SourceBinding, error)

	Delete(ctx context.Context, id domain.SourceBindingID) error
}
