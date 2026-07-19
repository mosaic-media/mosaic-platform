package postgres

import (
	"testing"
	"time"
)

// TestUUIDv7Layout checks the bits PostgreSQL's uuid type and the btree
// locality argument both depend on: a malformed version or variant nibble
// would still store fine and still sort, so nothing else would catch it.
func TestUUIDv7Layout(t *testing.T) {
	id := string(NewUUIDv7Generator().NewID())

	if len(id) != 36 {
		t.Fatalf("length = %d, want 36: %q", len(id), id)
	}
	for _, i := range []int{8, 13, 18, 23} {
		if id[i] != '-' {
			t.Fatalf("expected '-' at index %d of %q", i, id)
		}
	}
	// Version 7 is the first nibble of the third group.
	if id[14] != '7' {
		t.Fatalf("version nibble = %q, want '7': %q", id[14], id)
	}
	// The RFC 4122 variant is 10xx, so the first nibble of the fourth group
	// is one of 8, 9, a or b.
	switch id[19] {
	case '8', '9', 'a', 'b':
	default:
		t.Fatalf("variant nibble = %q, want one of 8/9/a/b: %q", id[19], id)
	}
}

// TestUUIDv7IsTimeOrdered is the property the content model actually buys by
// using v7 over v4: ids minted later sort after ids minted earlier, so btree
// inserts append near the right-hand edge instead of scattering.
func TestUUIDv7IsTimeOrdered(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	at := base
	g := uuidV7Generator{now: func() time.Time { return at }}

	var previous string
	for i := range 100 {
		at = base.Add(time.Duration(i) * time.Millisecond)
		id := string(g.NewID())
		if previous != "" && id <= previous {
			t.Fatalf("id %d (%q) does not sort after its predecessor (%q)", i, id, previous)
		}
		previous = id
	}
}

// TestUUIDv7IsUnique guards the random tail: two ids minted within the same
// millisecond share a timestamp prefix and must still differ.
func TestUUIDv7IsUnique(t *testing.T) {
	fixed := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	g := uuidV7Generator{now: func() time.Time { return fixed }}

	seen := make(map[string]bool, 1000)
	for range 1000 {
		id := string(g.NewID())
		if seen[id] {
			t.Fatalf("duplicate id within one millisecond: %q", id)
		}
		seen[id] = true
	}
}
