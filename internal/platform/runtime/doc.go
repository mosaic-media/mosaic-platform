// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Package runtime builds the Supervisor handoff surface:
// Generation metadata (metadata.go), a process Lifecycle (lifecycle.go),
// Readiness/Liveness health (readiness.go, liveness.go), migration status
// (migration.go), configuration activation status (config_activation.go),
// and a graceful Shutdown hook (shutdown.go).
//
// This package must not import internal/modules/postgres or any other
// Module, per the inward dependency rule: the composition root
// (cmd/mosaic-platform/main.go) is what bridges concrete Postgres/events
// values into these adapter-agnostic functions, the same way it already
// does for internal/platform/diagnostics.
package runtime
