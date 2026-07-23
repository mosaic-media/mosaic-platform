// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package session

import (
	"testing"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

func ep(id string) v1.Node {
	return v1.Node{ID: v1.NodeID(id), Kind: v1.NodeItem, ItemType: v1.ItemEpisode}
}

func TestEpisodeAfter(t *testing.T) {
	season := []v1.Node{ep("e1"), ep("e2"), ep("e3")}

	if got, ok := episodeAfter(season, "e1"); !ok || got.ID != "e2" {
		t.Fatalf("after e1 = %q (%v), want e2", got.ID, ok)
	}
	if _, ok := episodeAfter(season, "e3"); ok {
		t.Fatal("after the last episode there is no next in the season")
	}
	if _, ok := episodeAfter(season, "missing"); ok {
		t.Fatal("an id not in the season has no next")
	}
}

func TestEpisodeAfterSkipsNonEpisodes(t *testing.T) {
	// A trailer or a special sitting between two episodes must not be offered as
	// "next" — the walk steps over anything that is not an episode item.
	special := v1.Node{ID: "extra", Kind: v1.NodeItem, ItemType: v1.ItemType("special")}
	season := []v1.Node{ep("e1"), special, ep("e2")}

	if got, ok := episodeAfter(season, "e1"); !ok || got.ID != "e2" {
		t.Fatalf("after e1 = %q (%v), want e2 (the special is skipped)", got.ID, ok)
	}
}

func TestSeasonsAfter(t *testing.T) {
	seasons := []v1.Node{
		{ID: "s1", Kind: v1.NodeContainer},
		{ID: "s2", Kind: v1.NodeContainer},
		{ID: "s3", Kind: v1.NodeContainer},
	}

	got := seasonsAfter(seasons, "s1")
	if len(got) != 2 || got[0].ID != "s2" || got[1].ID != "s3" {
		t.Fatalf("after s1 = %v, want [s2 s3]", got)
	}
	if got := seasonsAfter(seasons, "s3"); len(got) != 0 {
		t.Fatalf("after the last season = %v, want none", got)
	}
	if got := seasonsAfter(seasons, "missing"); got != nil {
		t.Fatalf("after an unknown season = %v, want nil", got)
	}
}
