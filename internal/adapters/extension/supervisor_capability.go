// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package extension

import (
	"context"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// Supervised is the v1.Capability the registry holds, and it implements every
// provider role so role resolution works exactly as it does for a compiled-in
// module (the registry gates on the manifest, not on a type assertion — see
// capability_registry.go). Each method routes to the live process, or reports
// the capability degraded when there is none.
//
// Written out method by method for the same reason guard.go is: this is the
// seam where a down module becomes a degraded capability rather than a crash,
// and a reader should be able to see that every path checks liveness first
// without trusting a helper to have covered them all.

var (
	_ v1.Capability         = (*Supervised)(nil)
	_ v1.MetadataProvider   = (*Supervised)(nil)
	_ v1.SearchProvider     = (*Supervised)(nil)
	_ v1.CatalogProvider    = (*Supervised)(nil)
	_ v1.StreamProvider     = (*Supervised)(nil)
	_ v1.SubtitlesProvider  = (*Supervised)(nil)
	_ v1.ArtworkProvider    = (*Supervised)(nil)
	_ v1.PlaybackProvider   = (*Supervised)(nil)
	_ v1.SettingsUIProvider = (*Supervised)(nil)
)

// Manifest is answered from the cached copy taken at first launch, so a module's
// identity and roles are stable even while the process is momentarily down. This
// is what keeps role resolution from flickering across a restart.
func (s *Supervised) Manifest() v1.Manifest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.manifest
}

func (s *Supervised) Import(ctx context.Context, svc v1.ContentService, req v1.ImportRequest) (v1.ImportResult, error) {
	m := s.live()
	if m == nil {
		return v1.ImportResult{}, s.unavailable()
	}
	return m.Capability.Import(ctx, svc, req)
}

func (s *Supervised) Metadata(ctx context.Context, req v1.MetadataRequest) (v1.ContentMetadata, error) {
	m := s.live()
	if m == nil {
		return v1.ContentMetadata{}, s.unavailable()
	}
	return m.Capability.(v1.MetadataProvider).Metadata(ctx, req)
}

func (s *Supervised) Search(ctx context.Context, req v1.SearchRequest) (v1.SearchResponse, error) {
	m := s.live()
	if m == nil {
		return v1.SearchResponse{}, s.unavailable()
	}
	return m.Capability.(v1.SearchProvider).Search(ctx, req)
}

func (s *Supervised) Catalogs(ctx context.Context, req v1.CatalogsRequest) (v1.CatalogsResponse, error) {
	m := s.live()
	if m == nil {
		return v1.CatalogsResponse{}, s.unavailable()
	}
	return m.Capability.(v1.CatalogProvider).Catalogs(ctx, req)
}

func (s *Supervised) CatalogItems(ctx context.Context, req v1.CatalogItemsRequest) (v1.CatalogItemsResponse, error) {
	m := s.live()
	if m == nil {
		return v1.CatalogItemsResponse{}, s.unavailable()
	}
	return m.Capability.(v1.CatalogProvider).CatalogItems(ctx, req)
}

func (s *Supervised) Streams(ctx context.Context, req v1.StreamRequest) (v1.StreamResponse, error) {
	m := s.live()
	if m == nil {
		return v1.StreamResponse{}, s.unavailable()
	}
	return m.Capability.(v1.StreamProvider).Streams(ctx, req)
}

func (s *Supervised) Subtitles(ctx context.Context, req v1.SubtitlesRequest) (v1.SubtitlesResponse, error) {
	m := s.live()
	if m == nil {
		return v1.SubtitlesResponse{}, s.unavailable()
	}
	return m.Capability.(v1.SubtitlesProvider).Subtitles(ctx, req)
}

func (s *Supervised) Artwork(ctx context.Context, req v1.ArtworkRequest) (v1.ArtworkResponse, error) {
	m := s.live()
	if m == nil {
		return v1.ArtworkResponse{}, s.unavailable()
	}
	return m.Capability.(v1.ArtworkProvider).Artwork(ctx, req)
}

func (s *Supervised) Resolve(ctx context.Context, req v1.PlaybackRequest) (v1.PlaybackResolution, error) {
	m := s.live()
	if m == nil {
		return v1.PlaybackResolution{}, s.unavailable()
	}
	return m.Capability.(v1.PlaybackProvider).Resolve(ctx, req)
}

func (s *Supervised) SettingsUI(ctx context.Context, req v1.SettingsUIRequest) (v1.SettingsUIResponse, error) {
	m := s.live()
	if m == nil {
		return v1.SettingsUIResponse{}, s.unavailable()
	}
	return m.Capability.(v1.SettingsUIProvider).SettingsUI(ctx, req)
}
