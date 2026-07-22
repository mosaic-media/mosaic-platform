// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package telemetry_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mosaic-media/platform/internal/platform/telemetry"
)

func TestTraceparentRoundTrips(t *testing.T) {
	tc := telemetry.NewTraceContext()
	parsed, ok := telemetry.ParseTraceparent(tc.Traceparent())
	if !ok {
		t.Fatalf("a freshly minted traceparent must parse: %q", tc.Traceparent())
	}
	if parsed.TraceID != tc.TraceID || parsed.SpanID != tc.SpanID || parsed.Sampled != tc.Sampled {
		t.Fatalf("round trip lost information: %+v vs %+v", parsed, tc)
	}
}

// TestParseTraceparentRejectsMalformedInput matters more than the happy path:
// the header is attacker-controlled, so anything not exactly right must be
// discarded outright rather than repaired or partially accepted (ADR 0054).
func TestParseTraceparentRejectsMalformedInput(t *testing.T) {
	cases := map[string]string{
		"empty":            "",
		"too few parts":    "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331",
		"too many parts":   "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01-extra",
		"short trace id":   "00-0af7651916cd43dd-b7ad6b7169203331-01",
		"short span id":    "00-0af7651916cd43dd8448eb211c80319c-b7ad6b71-01",
		"non-hex trace id": "00-zzf7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
		"forbidden ff":     "ff-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
		"zero trace id":    "00-00000000000000000000000000000000-b7ad6b7169203331-01",
		"zero span id":     "00-0af7651916cd43dd8448eb211c80319c-0000000000000000-01",
		"sql-ish":          "00-' OR 1=1 --00000000000000000000-b7ad6b7169203331-01",
	}
	for name, value := range cases {
		if _, ok := telemetry.ParseTraceparent(value); ok {
			t.Errorf("%s: %q must be rejected", name, value)
		}
	}
}

// TestParseTraceparentAcceptsAFutureVersion guards a subtle failure: rejecting
// an unknown version would silently break correlation with a newer caller,
// which is worse than accepting a header whose extra fields we ignore.
func TestParseTraceparentAcceptsAFutureVersion(t *testing.T) {
	if _, ok := telemetry.ParseTraceparent("01-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"); !ok {
		t.Fatal("a future version must still correlate")
	}
}

func TestChildKeepsTheTraceAndChangesTheSpan(t *testing.T) {
	parent := telemetry.NewTraceContext()
	child := parent.Child()
	if child.TraceID != parent.TraceID {
		t.Fatal("a child must stay in the same trace")
	}
	if child.SpanID == parent.SpanID {
		t.Fatal("a child must get its own span")
	}
}

// TestStartRequestContinuesAnInboundTrace is the property the whole
// cross-repository story rests on: the Shell's id must survive into the
// Platform, not be replaced by a fresh one.
func TestStartRequestContinuesAnInboundTrace(t *testing.T) {
	upstream := telemetry.NewTraceContext()

	ctx := telemetry.StartRequest(context.Background(), upstream.Traceparent())
	got, ok := telemetry.TraceFrom(ctx)
	if !ok {
		t.Fatal("StartRequest must seed a trace context")
	}
	if got.TraceID != upstream.TraceID {
		t.Fatalf("trace id changed across the edge: %s != %s", got.TraceIDString(), upstream.TraceIDString())
	}
	if got.SpanID == upstream.SpanID {
		t.Fatal("this process's work must be its own span, not the caller's")
	}
}

func TestStartRequestMintsWhenTheHeaderIsAbsentOrJunk(t *testing.T) {
	for _, header := range []string{"", "not-a-traceparent"} {
		ctx := telemetry.StartRequest(context.Background(), header)
		tc, ok := telemetry.TraceFrom(ctx)
		if !ok || !tc.Valid() {
			t.Fatalf("header %q: a malformed or absent header must start a fresh trace, not leave none", header)
		}
	}
}

// TestStartRequestBindsTheTraceToTheLogger closes the loop: seeding a trace is
// only useful if records emitted downstream actually carry it without anyone
// naming it.
func TestStartRequestBindsTheTraceToTheLogger(t *testing.T) {
	var buf bytes.Buffer
	base := telemetry.Into(context.Background(),
		telemetry.New(telemetry.NewJSONSink(&buf), telemetry.Resource{}, telemetry.LevelDebug))

	upstream := telemetry.NewTraceContext()
	ctx := telemetry.StartRequest(base, upstream.Traceparent())
	telemetry.From(ctx).Info("did some work")

	line := parseLogLine(t, buf.Bytes())
	if line["trace"] != upstream.TraceIDString() {
		t.Fatalf("trace = %v, want %s", line["trace"], upstream.TraceIDString())
	}
	if line["span"] == "" || line["span"] == nil {
		t.Fatal("expected a span id on the record")
	}
}

// TestUnsampledTraceStillCarriesItsID is ADR 0054's sampling rule: a sampling
// decision governs whether spans are recorded, never whether the id exists —
// otherwise an unsampled failure is unjoinable to its own logs.
func TestUnsampledTraceStillCarriesItsID(t *testing.T) {
	var buf bytes.Buffer
	base := telemetry.Into(context.Background(),
		telemetry.New(telemetry.NewJSONSink(&buf), telemetry.Resource{}, telemetry.LevelDebug))

	upstream := telemetry.NewTraceContext()
	upstream.Sampled = false

	ctx := telemetry.StartRequest(base, upstream.Traceparent())
	tc, _ := telemetry.TraceFrom(ctx)
	if tc.Sampled {
		t.Fatal("the sampling decision must be carried, not overridden")
	}
	telemetry.From(ctx).Info("unsampled but recorded")
	if line := parseLogLine(t, buf.Bytes()); line["trace"] != upstream.TraceIDString() {
		t.Fatalf("an unsampled trace must still stamp its id, got %v", line["trace"])
	}
}

func TestHTTPMiddlewareSeedsAndEchoesTheTrace(t *testing.T) {
	var seen string
	h := telemetry.HTTPMiddleware("artwork", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc, _ := telemetry.TraceFrom(r.Context())
		seen = tc.TraceIDString()
		w.WriteHeader(http.StatusOK)
	}))

	upstream := telemetry.NewTraceContext()
	req := httptest.NewRequest(http.MethodGet, "/artwork", nil)
	req.Header.Set(telemetry.TraceparentHeader, upstream.Traceparent())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen != upstream.TraceIDString() {
		t.Fatalf("handler saw trace %q, want %q", seen, upstream.TraceIDString())
	}
	// Echoed back so a bug report arrives with the one string that makes it
	// reconstructible, even when the reporter is a browser network tab.
	if got := rec.Header().Get(telemetry.TraceIDHeader); got != upstream.TraceIDString() {
		t.Fatalf("response header = %q, want %q", got, upstream.TraceIDString())
	}
}

func TestHTTPMiddlewareStartsATraceWithoutAHeader(t *testing.T) {
	h := telemetry.HTTPMiddleware("handoff", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	got := rec.Header().Get(telemetry.TraceIDHeader)
	if len(got) != 32 || strings.Trim(got, "0") == "" {
		t.Fatalf("expected a freshly minted trace id, got %q", got)
	}
}
