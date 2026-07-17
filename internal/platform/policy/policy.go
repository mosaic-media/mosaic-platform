package policy

import (
	"context"

	"github.com/mosaic-media/mosaic-platform/internal/platform/domain"
)

// Subject identifies the caller a decision is evaluated for, plus the
// session attributes MEG-009 §04's Policy Decision Point example lists
// under "subject" (roles, session strength, device trust).
type Subject struct {
	UserID       domain.UserID
	AuthStrength domain.AuthStrength
}

// Action identifies the operation being authorized, for example
// "user.create" or "user.session.revoke" (MEG-015 §07 examples).
type Action string

// Resource identifies what an Action would act upon.
type Resource struct {
	Type string
	ID   string
}

// PolicyContext carries request-scoped ABAC attributes MEG-009 §04 lists
// under "context" — network origin, admin mode, recovery mode, and
// similar. It is intentionally sparse for this slice's simple rules.
type PolicyContext struct {
	AdminMode    bool
	RecoveryMode bool
}

// Decision is a Policy Decision Point's answer to one authorization
// request. Reason exists so denials remain explainable (MEG-009 §04 —
// Auditability).
type Decision struct {
	Allowed bool
	Reason  string
}

// PolicyDecisionPoint answers authorize(subject, action, resource,
// context) with a Decision. It keeps the ABAC-ready shape required by
// MEG-009 §04 regardless of how simple the underlying rules are.
type PolicyDecisionPoint interface {
	Authorize(ctx context.Context, subject Subject, action Action, resource Resource, policyContext PolicyContext) (Decision, error)
}
