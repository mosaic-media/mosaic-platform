# Slice 14 — Transaction contract shape: contract (MAD-001)

> Not one of MEG-015 §12's original 14 named slices. This is the second half
> of the expand/contract migration started in slice 13: that slice added
> `Store[T]`/`StorageAdapter` alongside the existing six-method `Tx` without
> touching a single caller; this slice migrates every remaining caller onto
> `Store[T]` and then seals `Tx`. MEG-015 §12's own "Reference capability
> path" and "SDK extraction readiness" shift to local positions 15 and 16.

---

## 1. Before you start

Read `CLAUDE.md` in full, especially the "Transaction contract shape
correction (MAD-001)" section and the status checklist. Confirm both of the
following are marked `[x]` done — name them explicitly, and **stop and
report if either is missing**:

- **Slice 13 — Transaction contract shape: expand.** This slice depends
  entirely on `Store[T]` and `StorageAdapter` already existing in
  `internal/platform/contracts/`, already proven equivalent to the six
  existing `Tx` accessors against real PostgreSQL. If slice 13 isn't done,
  there is nothing here to migrate callers *onto*.
- **Supervisor handoff** (and, transitively, everything before it —
  Repository scaffold through Diagnostics and health). Same reasoning as
  slice 13's prerequisite check: this slice touches code from nearly every
  earlier slice.

## 2. What this slice implements

- [MEG-015 §03 — Platform Contracts](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md) — the sealed `Tx` shape (marker method, no accessors) becomes real in this slice, not just documented.
- [MEG-015 §12 — Build Sequence](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/12-build-sequence.md), specifically **"Contract Promotion Within Slice 13"** and **"Storage Contract Correction"** — read these again; they tell you what still belongs to local slice 15 (Reference capability) and must *not* be pulled forward into this one.
- [MAD-001 — Transactional Store Extensibility](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/index.md) §04 (Consequences) — "Every `tx.Users()`-style call site across the thirteen built slices moves to uniform resolution. The change is contained to how stores are *obtained*, not what they do." That sentence is this slice's entire scope in one line.
- Whatever this repository's slice-13 CLAUDE.md entry says about the exact
  mechanism it built for `Store[T]` — read that before assuming a shape;
  it's the authoritative record of what slice 13 actually did, not what this
  prompt predicted it would do.

## 3. Scope — what to build

This is a mechanical, compiler-driven migration against a shape slice 13
already pinned down and proved. Confirm the current state of each file
before editing it — don't assume it still matches this list if slice 13
made different choices than expected.

1. **Migrate the seven command handlers** in `internal/platform/app/` —
   `create_local_user.go`, `authenticate_local_user.go`, `revoke_session.go`,
   `set_user_status.go`, `draft_config_version.go`, `validate_config_version.go`,
   `activate_config_version.go` — from `tx.Foo()` calls to
   `contracts.Store[contracts.FooStore](tx)`. The eight-step command order
   (MEG-015 §04) does not change, only the store-acquisition line. Decide
   how a `Store[T]` resolution failure maps to a Platform `ErrorCategory`
   (likely `Internal`, since an unresolvable store is a Platform wiring
   defect, not caller input — confirm against how each handler already
   handles an unexpected contract error rather than inventing a new pattern
   per handler).

2. **Migrate the three `Tx` fakes** to use whatever `Store[T]`-compatible
   mechanism slice 13 built, returning the same fake store instances they
   already construct:
   - `internal/platform/contracts/contracts_test.go`'s `mockTx`
   - `internal/platform/app/fakes_test.go`'s `fakeTx`
   - `internal/transport/graphql/fakes_test.go`'s `fakeTx`

3. **Migrate `internal/modules/postgres/outbox_atomicity_test.go` and
   `outbox_worker_test.go`** off `tx.Users()`/`tx.Outbox()` onto `Store[T]`.

4. **Confirm zero remaining callers.** After steps 1–3, `go build ./...`
   plus a repo-wide search for `.Users()`, `.Sessions()`, `.Permissions()`,
   `.Config()`, `.Outbox()`, `.Credentials()` called on a `contracts.Tx`
   value should turn up nothing outside the six method definitions
   themselves. If anything remains, migrate it — do not seal `Tx` with a
   live caller still depending on the old accessors.

5. **Seal `Tx`.** Once step 4 confirms zero remaining callers, remove the
   six accessor methods from `internal/platform/contracts/unit_of_work.go`,
   leaving `Tx` as the marker-only interface from MEG-015 §03's example. If
   slice 13 built the internal registry mechanism directly into the
   concrete Postgres `tx` type and the three fakes independently, sealing
   should not require touching any of them again — if it does, that's a
   sign slice 13's mechanism was coupled to the accessor methods more
   tightly than intended; fix the coupling rather than leaving `Tx` half
   sealed.

6. **Add the structural boundary test.** Following this repo's existing
   `go/parser`-based boundary-test convention (see
   `internal/transport/graphql/boundary_test.go` and
   `internal/transport/health`'s equivalent), add a test proving `Tx` has no
   exported methods. Verify it actually catches a regression the way those
   earlier tests were verified: temporarily add a `Users()` method back to
   `Tx`, confirm the new test fails, then remove it again before committing.

## 4. Explicitly out of scope for this slice

- **Do not populate `contracts/platform/v1`.** MEG-015 §12's "Contract
  Promotion Within Slice 13" places that as the first step of local slice 15
  (Reference capability path) — this slice only finishes the shape in
  `internal/platform/contracts` (private); it does not promote anything.
- **Do not build the notes/tags (or any) reference capability.** Local slice
  15's own deliverable, and it depends on this slice being done first.
- **Do not build a second storage adapter (e.g. SQLite).** The port shape
  only needs to make substitution possible in principle.
- **Do not touch the query-only app handlers** — `get_user_by_id.go`,
  `list_users.go`, `get_active_config_version.go`, `get_config_version.go`,
  `get_roles_for_user.go`, `get_grants_for_user.go`,
  `get_effective_permissions.go`. They read through a directly-injected
  store and never go through `Tx`; confirm this is still true before
  skipping them, but do not add work here if it is.
- **Do not touch** `internal/platform/secrets/`, `internal/platform/diagnostics/`,
  `internal/platform/runtime/`, `internal/transport/health/`, or
  `internal/platform/events/` (`Bus`/`Worker`) — none of them reference
  `contracts.Tx`. If you find one that does, CLAUDE.md's blast-radius record
  was wrong; flag that explicitly rather than silently fixing it as a
  drive-by.
- **Do not redesign anything slice 13 already decided** (the `Store[T]`
  resolution mechanism, the `StorageAdapter` shape). This slice consumes
  those decisions; it doesn't revisit them. If you believe slice 13 got
  something wrong, stop and report rather than quietly diverging.

## 5. Exit criteria

MEG-015 §12's table has no row for this slice either, but this is the slice
where MAD-001's full decision actually lands in code. Near-verbatim from
MAD-001 §02:

> "The transaction scope exposes no fixed list of stores. Every store — Core
> Platform or capability — is resolved the same way, so nothing is
> privileged by being named on the transaction handle and adding a store
> never edits it."

Treat the exit criteria as: **`Tx` is sealed (no exported accessor methods),
every call site in the repository resolves stores through `Store[T]`, and
every test suite from every earlier slice — atomicity, policy denial,
config activation, auth, GraphQL routing — still passes with unchanged
intent.**

Concrete tests that would actually prove it, not just "write some tests":

- Full `go test ./... -race` green against real PostgreSQL, with no test's
  *assertions* changed from what they proved before this migration — only
  the store-acquisition plumbing inside each test changed.
- The structural boundary test from §3.6 passes, and was verified to
  actually fail when a `Users()` method is temporarily added back to `Tx`.
- A repo-wide grep for `tx.Users(`, `tx.Sessions(`, `tx.Permissions(`,
  `tx.Config(`, `tx.Outbox(`, `tx.Credentials(` against `contracts.Tx`
  values returns nothing.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, and `go test ./... -race` (against
real PostgreSQL). Report the actual results. Update CLAUDE.md's status
checklist: add a new `[x]` entry for this slice, and update the existing
"Reference capability path" bullet (renumber it local slice 15 in your own
notes if useful) to state plainly that its blocker is now cleared on the
code side, not just the docs side — `contracts/platform/v1` can now be
populated with a `Tx`/`Store[T]`/`StorageAdapter` surface that has no
private-internals dependency baked into its shape.

## 7. If anything is ambiguous

This slice should not encounter architectural ambiguity — that was slice
13's job, and by design this one only consumes what slice 13 already
proved. If you hit a case that feels like a genuine design decision rather
than a mechanical translation (for example, the `ErrorCategory` mapping in
§3.1, or a fake that doesn't cleanly support whatever mechanism slice 13
built), check:

- Slice 13's actual CLAUDE.md writeup for how it intended `Store[T]` to be
  consumed by non-Postgres implementations (fakes) — it should have left
  enough of a pattern that the three fakes don't each need a novel
  approach.
- [MEG-015 §03 — Error Categories](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md) for the category table, before defaulting to `Internal` by assumption.

If it's still ambiguous after that, it may mean slice 13 under-specified
something it should have pinned down — say so explicitly in your report
rather than resolving it silently, since that's feedback slice 13's
approach may need, not something to route around quietly here.

---

**Recommended model / effort: Sonnet 5, medium effort.** This is mechanical,
compiler-driven call-site migration against a shape already fully specified
by slice 13 — closer to the "Repository scaffold" or fake-updating work in
earlier slices than to genuine design work. Bump to high effort only if the
session finds real judgment calls beyond direct mechanical translation (see
§7) — but the default expectation is that it won't need to.
