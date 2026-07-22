// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package app

import (
	"context"

	"github.com/mosaic-media/platform/internal/platform/contracts"
	"github.com/mosaic-media/platform/internal/platform/policy"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// ResolvePlaybackQuery asks where an item's playable bytes are, right now. It
// names the Part rather than the node because an item may have several — the
// editions and segments of ADR 0013 — and which one plays is the caller's
// choice, not this handler's guess.
type ResolvePlaybackQuery struct {
	Caller v1.Caller
	PartID v1.PartID
}

// ResolvePlaybackResult is the upstream location a playback provider resolved,
// for the Platform's own origin to relay from (ADR 0045).
//
// It is deliberately *not* something to hand a client. URL and Headers may
// carry a debrid credential, and keeping them server-side is half the reason
// the origin exists — the transport seals them into a ticket rather than
// emitting them.
type ResolvePlaybackResult struct {
	// ModuleID is the playback module that resolved this, carried for
	// diagnostics.
	ModuleID string
	// URL is the location to fetch.
	URL string
	// Headers are request headers the URL's origin requires, nil when it can be
	// fetched bare.
	Headers map[string]string
}

// ResolvePlayback turns a Part into a playable upstream location by asking the
// installed playback provider (ADR 0045's RolePlayback). Nothing here writes,
// and nothing here opens a transaction.
//
// It runs at play time, every time. The Part's stored location is what a source
// offered when the item was materialised, and for a debrid link that is a
// short-lived address which has very likely expired — so the Part is an
// identity hint handed to the provider, not an answer read back out of the
// graph.
func (s *Service) ResolvePlayback(ctx context.Context, q ResolvePlaybackQuery) (ResolvePlaybackResult, error) {
	if q.Caller.Session == "" {
		return ResolvePlaybackResult{}, contracts.NewError(contracts.InvalidArgument, "caller is required")
	}
	if q.PartID == "" {
		return ResolvePlaybackResult{}, contracts.NewError(contracts.InvalidArgument, "part id is required")
	}

	callerID, err := s.authenticateCaller(ctx, q.Caller)
	if err != nil {
		return ResolvePlaybackResult{}, err
	}
	if err := s.authorize(ctx, policy.Subject{UserID: callerID}, ActionContentRead, policy.Resource{Type: "content"}, policy.PolicyContext{}); err != nil {
		return ResolvePlaybackResult{}, err
	}

	if s.parts == nil {
		return ResolvePlaybackResult{}, contracts.NewError(contracts.Unavailable, "no part store configured")
	}
	part, err := s.parts.FindByID(ctx, q.PartID)
	if err != nil {
		return ResolvePlaybackResult{}, err
	}

	entry, ok := s.playbackProvider()
	if !ok {
		// This is ADR 0036's inert library, reported honestly rather than as a
		// failure to play: nothing is installed that can consume what
		// materialising created.
		return ResolvePlaybackResult{}, contracts.NewError(contracts.NotFound, "no playback module is installed")
	}

	settings, err := s.readModuleSettings(ctx, entry.ModuleID)
	if err != nil {
		return ResolvePlaybackResult{}, err
	}

	res, err := entry.Provider.Resolve(ctx, v1.PlaybackRequest{
		Caller:   q.Caller,
		Settings: settings,
		Part:     part,
	})
	if err != nil {
		return ResolvePlaybackResult{}, contracts.WrapError(contracts.Unavailable, "resolve playback", err)
	}
	if res.Kind != v1.PlaybackDirect {
		// The SDK declares one variant today; a module returning anything else
		// is built against a contract this Platform does not implement, which is
		// a wiring error rather than a source failure.
		return ResolvePlaybackResult{}, contracts.NewError(contracts.Internal, "playback module returned an unsupported resolution kind")
	}
	if res.URL == "" {
		return ResolvePlaybackResult{}, contracts.NewError(contracts.NotFound, "playback module resolved no location for this part")
	}

	return ResolvePlaybackResult{ModuleID: entry.ModuleID, URL: res.URL, Headers: res.Headers}, nil
}

// playbackProvider picks the playback provider to resolve through, tolerating a
// Service built without a registry.
//
// It takes the first in stable module-id order. That is a real choice and worth
// naming: precedence *between* two installed playback modules is undecided, and
// with one installed the question does not arise. It is the consumer-side twin
// of ADR 0027's open provider-precedence seam, and it should be settled with
// that one rather than invented here.
func (s *Service) playbackProvider() (PlaybackProviderEntry, bool) {
	if s.capabilities == nil {
		return PlaybackProviderEntry{}, false
	}
	entries := s.capabilities.PlaybackProviders()
	if len(entries) == 0 {
		return PlaybackProviderEntry{}, false
	}
	return entries[0], true
}
