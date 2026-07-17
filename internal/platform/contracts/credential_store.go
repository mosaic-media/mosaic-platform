package contracts

import (
	"context"

	"github.com/mosaic-media/mosaic-platform/internal/platform/domain"
)

// CredentialStore provides local credential persistence and lookup for
// the identity factors listed in MEG-015 §07 — Local Identity Scope:
// password credentials, passkey credential records and local recovery
// factors.
type CredentialStore interface {
	SavePassword(ctx context.Context, credential domain.PasswordCredential) error
	FindPassword(ctx context.Context, userID domain.UserID) (domain.PasswordCredential, error)

	SavePasskey(ctx context.Context, credential domain.PasskeyCredential) error
	ListPasskeys(ctx context.Context, userID domain.UserID) ([]domain.PasskeyCredential, error)

	SaveRecoveryFactor(ctx context.Context, factor domain.RecoveryFactor) error
	ConsumeRecoveryFactor(ctx context.Context, userID domain.UserID, codeHash string) (domain.RecoveryFactor, error)
}
