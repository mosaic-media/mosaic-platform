// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package app

import (
	"context"

	"github.com/mosaic-media/platform/internal/platform/contracts"
	"github.com/mosaic-media/platform/internal/platform/telemetry"
)

// moduleSpan brackets a call into a module (ADR 0055, seam 8).
//
// This is the seam that does the most work, because the module boundary is
// also a *repository* boundary: a trace that stopped here left the hardest
// question — "was it us or the addon?" — unanswerable, which is the specific
// pain that started the telemetry thread.
//
// It attributes the span to the module, and that attribution is stamped by the
// Platform rather than claimed by the module (ADR 0059). A module that never
// mentions telemetry is still fully covered: its span exists because the
// Platform started one, the context it receives already carries the trace, and
// the HTTP client the composition root hands it propagates that trace onward.
//
// Modules are statically composed today (ADR 0007). If they ever move out of
// process, this is replaced by an interceptor at the same seam rather than
// duplicated beside it — which is why the call sites take a context back
// rather than reaching for a package-level tracer.
func moduleSpan(ctx context.Context, moduleID, operation string) (context.Context, *telemetry.Span) {
	// The logger is re-bound to the module so log records emitted beneath this
	// call are attributed to it too, not only the span.
	ctx = telemetry.Into(ctx, telemetry.From(ctx).ForModule("module", moduleID))
	return telemetry.Start(ctx, "module."+operation,
		telemetry.String("module", moduleID),
		telemetry.String("operation", operation))
}

// failSpan records err against span using the Platform error category the
// caller will return, so a failed module span and the contract error a client
// receives describe the same failure the same way.
func failSpan(span *telemetry.Span, err error) {
	if err == nil {
		return
	}
	span.Fail(string(contracts.CategoryOf(err)), err)
}
