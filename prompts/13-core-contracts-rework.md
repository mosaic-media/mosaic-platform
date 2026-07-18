# Slice 13 — Core contracts rework (MAD-001, expand step 1 of 6)

> Not one of MEG-015 §12's original 14 named slices. This is the first of
> six rework slices that migrate the transaction contract from MAD-001's
> corrected shape into code, one originally-affected slice at a time, using
> an expand/contract migration: `Tx`'s existing six accessor methods are
> **not removed** until the very last rework slice (18). Every rework slice
> in between leaves the whole repo green — nothing is ever half-migrated
> across a commit boundary. Order: 13 Core contracts → 14 PostgreSQL adapter
> → 15 Application service skeleton → 16 Identity, sessions and policy → 17
> Configuration versioning → 18 GraphQL (which also seals `Tx`) → 19
> Reference capability path → 20 SDK extraction readiness.

---

## 1. Before you start

Read `CLAUDE.md` in full, especially "Transaction contract shape correction
(MAD-001)" and the status checklist. Confirm all twelve original slices —
Repository scaffold, Core contracts, Application service skeleton, Identity/
sessions/policy, PostgreSQL adapter and migrations, Transactional outbox,
In-process Event Bus, Configuration versioning and reload classes, Secret
broker, GraphQL command and query surface, Diagnostics and health,
Supervisor handoff — are `[x]`. If any is not, **stop and report**.

## 2. What this slice implements

- [MEG-015 §03 — Platform Contracts](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md) — Contract Shape (the `Store[T]` example) and Storage Extensibility Boundary.
- [MAD-001 — Transactional Store Extensibility](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/index.md) in full. `02-decision.md`/`05-implementation-implications.md` are load-bearing; `03-alternatives-considered.md` is required for §7 below.

## 3. Scope — what to build

Only `internal/platform/contracts/` and its own test file change in this
slice. Re-derive the current `Tx` shape from
`internal/platform/contracts/unit_of_work.go` yourself before starting.

1. Add `func Store[T any](tx Tx) (T, error)` (new file, e.g. `store.go`) and
   a `StorageAdapter` port (new file, e.g. `storage_adapter.go`) to
   `internal/platform/contracts/`. **Do not remove or modify `Tx`'s six
   existing methods** (`Users()`, `Sessions()`, `Permissions()`, `Config()`,
   `Outbox()`, `Credentials()`) — sealing is slice 18's job, five rework
   slices from now.
2. `Store[T]` must resolve, for each of the six known store types, to the
   *exact same store instance* the matching named accessor would return for
   the same `Tx` value — this slice can only prove that with a fake, since
   the contracts package has no real transaction; full atomicity proof
   against real PostgreSQL is slice 14's job.
3. Migrate `internal/platform/contracts/contracts_test.go`: **extend**
   `mockTx` (don't replace) so it satisfies `Store[T]` resolution in
   addition to its existing six methods, backed by the same mock store
   instances (`mockUserStore`, `mockEventOutbox`, etc.) the six methods
   already return. Add a test proving, for at least two store types, that
   `Store[T](mockTx)` and the matching named accessor return the identical
   instance.

## 4. Explicitly out of scope

- Nothing in `internal/modules/postgres/` — that's slice 14.
- Nothing in `internal/platform/app/` (command handlers or `fakes_test.go`)
  — that's slice 15.
- Nothing in `internal/transport/graphql/` — slice 18.
- Do not remove `Tx`'s six accessor methods.
- Do not populate `contracts/platform/v1` — slice 19.
- Do not build the reference capability or a second storage adapter.

## 5. Exit criteria

No MEG-015 §12 table row exists for this slice. Nearest authoritative
language, MAD-001 §02: *"Every store — Core Platform or capability — is
resolved the same way... adding a store never edits it."* Scoped to this
slice: **`Store[T]`/`StorageAdapter` exist, compile, and resolve correctly
against a fake `Tx`, with zero changes outside `internal/platform/contracts/`.**

Concrete test: the equivalence test from §3.3 passes; `go build ./...` and
`go test ./...` are green with the diff confined to
`internal/platform/contracts/`.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, `go test ./... -race`. Report results.
Add a new `- [x]` entry to CLAUDE.md's checklist for this slice, between
Supervisor handoff and slice 14, and record the exact `Store[T]`/
`StorageAdapter` shape chosen (slice 14 depends on reading this, not
guessing it).

## 7. If anything is ambiguous

This is where the genuine design decision lives. Read MAD-001
`03-alternatives-considered.md` for why option (a) — an `any`-keyed
extension registry — was rejected ("a store must ask permission to join the
transaction"); whatever mechanism you choose must not reintroduce that at
`Store[T]`'s caller-facing boundary (an internal, unexported registry is
fine). Read [MEG-004 §04 — Driven Ports](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-004-hexagonal-architecture/04-driven-ports.md) for the general guidance behind that rejection. If still ambiguous, make the
smallest decision satisfying type-safe call sites and document it in
CLAUDE.md the way earlier slices documented first-cut choices.

---

**Recommended model / effort: Opus 4.8, high effort.** This is the slice
where the actual design decision gets made; slices 14–18 consume it
mechanically.
