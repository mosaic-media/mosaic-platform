// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package app

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// fanOut runs fn concurrently over items and returns the flattened results in
// input order. It is the read-side's provider scatter-gather: a query fans out
// to every registered provider, and each provider call is an independent remote
// round-trip, so running them concurrently turns a sum of latencies into a max.
//
// fn returns an item's results and an optional fatal error. Returning a nil
// error contributes the results (possibly empty — a provider that is merely
// unreachable is skipped by returning nil, nil). The first fatal error aborts
// the gather: it cancels the derived context so in-flight siblings unwind, and
// is returned with no results. This is the concurrent form of the serial code's
// two error paths — skip a downed provider, but fail the query on a settings
// read that errors.
//
// Ordering is deterministic regardless of completion order: each item writes
// only its own result slot, and the slots are concatenated in input order. The
// Service methods fn calls already serve concurrent HTTP requests, so invoking
// them across goroutines here adds no new safety obligation.
func fanOut[T, R any](ctx context.Context, items []T, fn func(context.Context, T) ([]R, error)) ([]R, error) {
	if len(items) == 0 {
		return nil, nil
	}
	perItem := make([][]R, len(items))
	g, ctx := errgroup.WithContext(ctx)
	for i, item := range items {
		g.Go(func() error {
			res, err := fn(ctx, item)
			if err != nil {
				return err
			}
			perItem[i] = res
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	var out []R
	for _, r := range perItem {
		out = append(out, r...)
	}
	return out, nil
}
