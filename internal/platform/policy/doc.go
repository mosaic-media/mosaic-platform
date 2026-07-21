// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Package policy holds the Platform's Policy Decision Point. It answers
// authorize(subject, action, resource, context)
// and returns an allow/deny Decision. The decision point may live
// in-process, as it does here; the enforcement point — the code that
// actually refuses to mutate state on a deny — belongs entirely to
// application services (internal/platform/app), not to this package.
package policy
