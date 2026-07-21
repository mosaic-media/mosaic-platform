// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Package contracts holds the private Platform contract definitions. Contracts
// pass domain value types across their boundary and report failures through
// the Platform ErrorCategory scheme; they never leak database rows or
// driver-specific types.
package contracts
