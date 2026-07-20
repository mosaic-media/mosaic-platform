// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package graphql

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql"
)

// The SDUI action ABI (ADR 0029): an action carries its session in the
// Authorization header (context), and an Invoke wraps its argument as
// input:{...}. These check the two helpers that absorb that shape.

func TestCallerPrefersArgThenHeaderSession(t *testing.T) {
	// An explicit callerSessionId argument wins.
	p := graphql.ResolveParams{
		Context: withSession(context.Background(), "header-sess"),
		Args:    map[string]interface{}{"callerSessionId": "arg-sess"},
	}
	if got := caller(p).Session; got != "arg-sess" {
		t.Fatalf("caller = %q, want the explicit argument", got)
	}

	// Absent the argument, the header session (context) is used.
	p = graphql.ResolveParams{
		Context: withSession(context.Background(), "header-sess"),
		Args:    map[string]interface{}{},
	}
	if got := caller(p).Session; got != "header-sess" {
		t.Fatalf("caller = %q, want the header session", got)
	}

	// Neither present: an empty caller (which authenticates as nothing).
	p = graphql.ResolveParams{Context: context.Background(), Args: map[string]interface{}{}}
	if got := caller(p).Session; got != "" {
		t.Fatalf("caller = %q, want empty", got)
	}
}

func TestImportRefFromInputEnvelopeAndDirectArg(t *testing.T) {
	// The runtime's Invoke shape: input:{ref:{...}}.
	p := graphql.ResolveParams{Args: map[string]interface{}{
		"input": map[string]interface{}{"ref": map[string]interface{}{
			"provider": "stremio", "nativeId": "tt1254207", "nativeType": "movie",
			"externalScheme": "imdb", "externalId": "tt1254207",
		}},
	}}
	ref := importRef(p)
	if ref.Provider != "stremio" || ref.NativeID != "tt1254207" || ref.ExternalID != "tt1254207" {
		t.Fatalf("ref from input envelope = %+v", ref)
	}

	// A direct ref argument, for a non-action caller.
	p = graphql.ResolveParams{Args: map[string]interface{}{
		"ref": map[string]interface{}{"provider": "x", "nativeId": "y", "nativeType": "movie"},
	}}
	if got := importRef(p); got.Provider != "x" || got.NativeID != "y" {
		t.Fatalf("ref from direct arg = %+v", got)
	}
}
