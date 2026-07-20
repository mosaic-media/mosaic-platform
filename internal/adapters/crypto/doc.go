// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Package crypto holds non-module-shaped cryptographic helpers: AES-256-GCM
// encryption and key derivation for the Secret Broker's encrypted local vault
// fallback (internal/platform/secrets), and an Argon2id PasswordHasher.
//
// These are adapters, not a built-in module. Unlike the PostgreSQL module,
// there is no manifest and no registration through internal/composition/
// builtin: each helper fulfils a single small port rather than a broad
// contract surface, so the composition root wires it directly. PasswordHasher
// is chosen at that seam — it satisfies the domain.PasswordVerifier port
// (Hash/Verify) and is passed to app.Service and the admin bootstrap in
// main.go — so swapping it for bcrypt, scrypt or an HSM-backed signer is a
// one-line change there, behind the same port.
//
// The package itself imports no Platform code, which keeps it a pure crypto
// utility. The compile-time proof that PasswordHasher satisfies
// domain.PasswordVerifier therefore lives in the external test package
// (password_test.go), the idiomatic place to assert satisfaction of an
// external interface without coupling the adapter to it.
package crypto
