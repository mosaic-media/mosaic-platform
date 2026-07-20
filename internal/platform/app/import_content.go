package app

import (
	"context"

	"github.com/mosaic-media/mosaic-platform/internal/platform/contracts"
	"github.com/mosaic-media/mosaic-platform/internal/platform/policy"
	v1 "github.com/mosaic-media/mosaic-sdk/contracts/platform/v1"
)

// ActionContentImport is the policy action evaluated for invoking a capability
// to import content. It gates who may trigger an import at all. The capability
// then acts as the same caller, so each node it creates is authorised again
// under the content actions (ADR 0017) — this gate is the outer one.
const ActionContentImport policy.Action = "content.import"

// ImportContentCommand names a registered capability and the query to hand it.
// It is a Platform command a transport issues (the GraphQL importContent
// mutation), deliberately not part of the published ContentService: a
// capability is invoked by this command, it does not call it.
type ImportContentCommand struct {
	Caller       v1.Caller
	CapabilityID string
	Query        string
}

func validateImportContentCommand(cmd ImportContentCommand) error {
	if cmd.Caller.Session == "" {
		return contracts.NewError(contracts.InvalidArgument, "caller is required")
	}
	if cmd.CapabilityID == "" {
		return contracts.NewError(contracts.InvalidArgument, "capability id is required")
	}
	if cmd.Query == "" {
		return contracts.NewError(contracts.InvalidArgument, "query is required")
	}
	return nil
}

// ImportContent invokes a registered capability to source content into the
// graph. It follows the command boundary up to authorization, then hands the
// capability the Service itself as its ContentService and the original Caller,
// so every write the capability makes re-enters the same command order and is
// authorised as the invoking user (ADR 0017).
//
// It opens no UnitOfWork of its own: the capability's service calls each open
// theirs, one transaction per write. A capability that fails partway leaves
// the writes it already committed in place — import is not atomic across the
// whole tree, by the same reasoning that lets it search between writes.
func (s *Service) ImportContent(ctx context.Context, cmd ImportContentCommand) (v1.ImportResult, error) {
	// 1. validate command shape.
	if err := validateImportContentCommand(cmd); err != nil {
		return v1.ImportResult{}, err
	}

	// 2. authenticate caller.
	callerID, err := s.authenticateCaller(ctx, cmd.Caller)
	if err != nil {
		return v1.ImportResult{}, err
	}

	// 3. authorize the invocation itself.
	if err := s.authorize(ctx, policy.Subject{UserID: callerID}, ActionContentImport, policy.Resource{Type: "content"}, policy.PolicyContext{}); err != nil {
		return v1.ImportResult{}, err
	}

	// 4. resolve the capability by id.
	capability, ok := s.lookupCapability(cmd.CapabilityID)
	if !ok {
		return v1.ImportResult{}, contracts.NewError(contracts.NotFound, "no capability registered under id "+cmd.CapabilityID)
	}

	// 5. invoke it, forwarding the caller so it acts as the invoking user and
	// passing the Service as the ContentService it drives.
	result, err := capability.Import(ctx, s, cmd.Caller, cmd.Query)
	if err != nil {
		return v1.ImportResult{}, err
	}

	// 6. record that an import ran, for audit. The capability's own writes each
	// emit their content events; this marks the invocation itself.
	s.publishAuditEvent(ctx, "content.import.invoked", []byte(cmd.CapabilityID), string(callerID))

	return result, nil
}

// lookupCapability resolves a capability id against the registry, tolerating a
// Service constructed without one (every test that does not exercise import).
func (s *Service) lookupCapability(id string) (v1.Capability, bool) {
	if s.capabilities == nil {
		return nil, false
	}
	return s.capabilities.Lookup(id)
}
