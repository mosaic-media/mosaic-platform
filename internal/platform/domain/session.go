// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package domain

import "time"

// AuthStrength records which authentication factor produced a Session, so
// policy decisions can weigh session strength (MEG-009 §04 — Attribute-
// Based Access Control) without depending on how the factor was verified.
type AuthStrength string

const (
	AuthStrengthPassword AuthStrength = "password"
	AuthStrengthPasskey  AuthStrength = "passkey"
	AuthStrengthRecovery AuthStrength = "recovery"
)

// Session is a server-issued, revocable Platform session (MEG-015 §07 —
// Session Model). Fields match §07's session table exactly, plus RevokedAt
// to record the revocation §07 requires ("sessions should be ... revocable
// ... remote sign-out should revoke server-side session records, not rely
// on clients deleting tokens").
type Session struct {
	ID           SessionID
	UserID       UserID
	DeviceID     DeviceID
	IssuedAt     time.Time
	LastSeenAt   time.Time
	ExpiresAt    time.Time
	AuthStrength AuthStrength
	Capabilities []Permission
	RevokedAt    *time.Time
}

// Revoked reports whether the session has been explicitly revoked.
func (s Session) Revoked() bool {
	return s.RevokedAt != nil
}

// ExpiredAt reports whether the session is expired as of at.
func (s Session) ExpiredAt(at time.Time) bool {
	return !at.Before(s.ExpiresAt)
}
