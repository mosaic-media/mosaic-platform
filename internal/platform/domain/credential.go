// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package domain

import "time"

// PasswordCredential is a local user's password verifier record. Hash is
// an opaque, already-hashed value (Argon2id in production); the Domain
// never sees or stores a plaintext password.
type PasswordCredential struct {
	UserID    UserID
	Hash      string
	UpdatedAt time.Time
}

// PasskeyCredential is one WebAuthn/platform passkey registered to a user.
// PublicKey is opaque to the Domain; verification belongs to a future
// crypto adapter.
type PasskeyCredential struct {
	UserID       UserID
	CredentialID string
	PublicKey    []byte
	CreatedAt    time.Time
}

// RecoveryFactor is one single-use account recovery code. Only its hash is
// held; ConsumedAt is set once the factor has been used, so a recovery
// ceremony invalidates the consumed code.
type RecoveryFactor struct {
	UserID     UserID
	CodeHash   string
	CreatedAt  time.Time
	ConsumedAt *time.Time
}

// Consumed reports whether this recovery factor has already been used.
func (f RecoveryFactor) Consumed() bool {
	return f.ConsumedAt != nil
}

// PasswordVerifier hashes and verifies password credentials. It is a
// Driven Port: the Domain needs only to hash and verify a credential, not a
// specific hashing algorithm, which belongs to a crypto adapter.
type PasswordVerifier interface {
	Hash(plaintext string) (string, error)
	Verify(plaintext string, hash string) (bool, error)
}
