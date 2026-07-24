// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Command extprobe is a trivial out-of-process module, and the first thing to
// cross the extension boundary for real (ADR 0064's build order, step 1: "a
// trivial in-repo module implementing one role — nothing user-visible;
// establishes the wire, the handshake and the handle").
//
// It is deliberately not a useful module. What it proves is that the mechanism
// works end to end where the in-package tests could not reach:
//
//   - go-plugin's handshake over a real Unix socket, in a real child process
//   - a manifest read back across the boundary
//   - Import calling *back* into the Platform's ContentService, within the
//     invocation, carrying the Caller handle it was given
//   - one provider role, so role dispatch is exercised, and the seven it does
//     not fill, so the refusal path is too
//
// Its whole main is the line every module author writes. If that stops being
// true, this file is where it shows.
package main

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
	"github.com/mosaic-media/sdk/host"
)

// probeSettings is how a test asks the probe to do more work. It rides the
// ordinary settings document (ADR 0021), which the Platform stores and hands
// back uninterpreted — so using it here costs no special mechanism.
type probeSettings struct {
	// Children makes Import add this many child nodes after the work. It exists
	// for the callback-cost measurement ADR 0064 requires before the protocol is
	// fixed: a tree import's cost is (calls per import) × (cost per call), and
	// this is what lets the second be measured without the first being guessed.
	Children int `json:"children"`
}

// probe fills RoleSearch and nothing else. The registry resolves roles from the
// manifest rather than by type assertion — a proxy satisfies every provider
// interface — so declaring one role and implementing one role is what makes
// this a usable test of that.
type probe struct{}

func (probe) Manifest() v1.Manifest {
	return v1.Manifest{
		ID:       "extprobe",
		Version:  "v0.1.0",
		Name:     "Extension Probe",
		Provides: []v1.Role{v1.RoleSearch},
	}
}

// Import writes one Work through the ContentService it reaches over the
// callback stream, acting as the Caller it was handed (ADR 0017). The write is
// the point: it is the only way to prove the callback direction works, and the
// counts it returns are what the test asserts against.
func (probe) Import(ctx context.Context, svc v1.ContentService, req v1.ImportRequest) (v1.ImportResult, error) {
	// Telemetry is reached ambiently off the context, exactly as in process
	// (ADR 0059) — a module never holds one.
	v1.TelemetryFrom(ctx).Info("extprobe import",
		v1.String("native_id", req.Ref.NativeID),
		v1.Int("settings_bytes", len(req.Settings)),
	)

	out, err := svc.AddContentWork(ctx, v1.AddContentWorkCommand{
		Caller:    req.Caller,
		MediaType: req.Ref.MediaType,
		Title:     "probe: " + req.Ref.NativeID,
	})
	if err != nil {
		return v1.ImportResult{}, err
	}

	// A tree import is many calls, not one. This is what makes the probe able
	// to stand in for one — a season of episodes is exactly this shape.
	var settings probeSettings
	if len(req.Settings) > 0 {
		// A settings document the module cannot parse is the user's, not an
		// error to fail an import over: unknown or malformed fields mean the
		// defaults, which is how every other module here treats them.
		_ = json.Unmarshal(req.Settings, &settings)
	}

	for i := 0; i < settings.Children; i++ {
		if _, err := svc.AddContentChild(ctx, v1.AddContentChildCommand{
			Caller:       req.Caller,
			ParentID:     out.Work.ID,
			Kind:         v1.NodeItem,
			ItemType:     v1.ItemEpisode,
			Title:        "episode",
			NaturalOrder: float64(i),
		}); err != nil {
			return v1.ImportResult{}, err
		}
	}

	return v1.ImportResult{WorkID: out.Work.ID, Items: 1 + settings.Children}, nil
}

// Search echoes its query back as one result. It exists so role dispatch is
// exercised in both directions — the request converted on the way out, the
// result on the way back.
func (probe) Search(_ context.Context, req v1.SearchRequest) (v1.SearchResponse, error) {
	return v1.SearchResponse{Results: []v1.SearchResult{{
		Ref: v1.ContentRef{
			Provider:       "extprobe",
			NativeID:       req.Text,
			NativeType:     "movie",
			MediaType:      v1.MediaMovie,
			ExternalScheme: "probe",
			ExternalID:     req.Text,
		},
		Title: "probe result: " + req.Text,
		Year:  2026,
	}}}, nil
}

func main() {
	// Controlled death, for the lifecycle tests (ADR 0064's step 3). A real
	// module has no such switch; this exists only so the Platform's restart,
	// backoff and crash-loop policy can be exercised against a process that
	// actually dies rather than one whose crash is imagined.
	//
	// The exit is armed *before* Serve and fires from a goroutine, so the
	// handshake completes first and go-plugin reports a live process that then
	// dies — which is the case the monitor must catch. Exiting before Serve
	// would instead fail the launch, a different path the tests cover separately.
	armSelfDestruct()
	host.Serve(probe{})
}

// armSelfDestruct exits the process after a delay when EXTPROBE_EXIT_AFTER_MS is
// set, optionally only on the first launch (when EXTPROBE_CRASH_ONCE names a
// marker file that a surviving second launch finds already present).
func armSelfDestruct() {
	ms := os.Getenv("EXTPROBE_EXIT_AFTER_MS")
	if ms == "" {
		return
	}
	delay, err := strconv.Atoi(ms)
	if err != nil {
		return
	}

	if marker := os.Getenv("EXTPROBE_CRASH_ONCE"); marker != "" {
		if _, err := os.Stat(marker); err == nil {
			return // a previous launch left the marker; this one survives.
		}
		// First launch: leave the marker so the next one survives, then die.
		_ = os.WriteFile(marker, []byte("crashed"), 0o600)
	}

	go func() {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		os.Exit(1)
	}()
}
