// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Package secrets implements the Platform Secret Broker. Broker prefers the
// OS keychain and falls back to an encrypted local vault protected by a
// separate recovery key when the keychain is unavailable. It parses and
// formats the secret:// reference URIs that configuration values store
// instead of raw secret values.
package secrets
