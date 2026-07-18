# Slice 14 — PostgreSQL adapter rework (MAD-001, expand step 2 of 6)

> Second of six rework slices. `Tx`'s six accessor methods are still not
> removed — that's slice 18. This slice implements `Store[T]` for real,
> against a live PostgreSQL transaction, and is where the atomicity
> guarantee actually gets proven or disproven.

---

## 1. Before you start

Read `CLAUDE.md`. Confirm **Slice 13 — Core contracts rework** is `[x]` and
read its writeup for the exact `Store[T]`/`StorageAdapter` shape it built —
this slice implements that shape, it does not redesign it. If slice 13 is
not done, **stop and report**.

## 2. What this slice implements

- [MEG-015 §03 — Platform Contracts](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md) — "Storage itself is a port... the built-in PostgreSQL adapter can be replaced... without changing a call site."
- [MAD-001 §02 — Decision](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/02-decision.md) and [§05 — Implementation Implications](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/05-implementation-implications.md).
- [MEG-015 §11 — Test Gates](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/11-test-gates.md), the Outbox gate ("state change and event append commit atomically") — this slice re-proves that gate through the new path.

## 3. Scope — what to build

Re-derive the current state of `internal/modules/postgres/unit_of_work.go`
yourself; do not assume it matches an earlier description.

1. Give the concrete `tx` struct in `internal/modules/postgres/unit_of_work.go`
   whatever it needs to satisfy slice 13's `Store[T]` mechanism, resolving
   to the *same* store instances (`&userStore{q: t.q}`, `&eventOutbox{q: t.q}`,
   etc.) its six existing methods already construct — do not build a second,
   parallel construction path. Do not remove or change the six existing
   methods.
2. Per MAD-001 ("a `StorageAdapter` provides the `UnitOfWork`"), consider
   whether `NewUnitOfWork(pool)` should now be reached through a
   `StorageAdapter`-shaped type implementing slice 13's port. Do not rewire
   `main.go` or `internal/composition/builtin/` beyond what's needed to keep
   the module compiling.
3. Migrate `internal/modules/postgres/outbox_atomicity_test.go` and
   `outbox_worker_test.go` off `tx.Users()`/`tx.Outbox()` onto `Store[T](tx)`.
4. Extend `test/contract/suite.go` with a `Store[T]`-based version of its
   existing mid-transaction-failure atomicity proof (state write + outbox
   append commit together; injected failure rolls both back, verified via
   raw SQL bypassing the stores) run against real PostgreSQL. Keep the
   existing accessor-based version too — this slice adds, it doesn't replace.
5. Add a test proving, against a real transaction, that a write through
   `Store[UserStore](tx)` is visible to a read through `tx.Users()` within
   the same `WithinTx` call (and vice versa) — the real-database version of
   slice 13's fake-backed equivalence proof.

## 4. Explicitly out of scope

- Nothing in `internal/platform/app/`, its fakes, or `internal/transport/graphql/`
  — slices 15–18.
- Do not remove `Tx`'s six accessor methods.
- Do not populate `contracts/platform/v1` or build the reference capability
  — slice 19.
- Do not build a second storage adapter (e.g. SQLite) — the port shape only
  needs to make one possible.

## 5. Exit criteria

No MEG-015 §12 table row exists for this slice. Nearest authoritative
language, MAD-001 §04: *"Atomicity. One transaction, one storage adapter —
unchanged. No parallel database..."* Scoped to this slice: **`Store[T]`
resolves correctly against a real PostgreSQL transaction, is proven
transaction-equivalent to the existing accessors, and the outbox/state
atomicity guarantee holds through the new path exactly as it does through
the old one.**

Concrete test: the extended `test/contract/suite.go` atomicity proof (§3.4)
and the equivalence test (§3.5) both pass against real PostgreSQL
(embedded-postgres or `MOSAIC_TEST_POSTGRES_DSN`); `outbox_atomicity_test.go`
and `outbox_worker_test.go` pass with unchanged assertions.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, `go test ./... -race` against real
PostgreSQL. Report results. Add a new `- [x]` entry to CLAUDE.md between
slice 13 and slice 15.

## 7. If anything is ambiguous

This slice should be implementing slice 13's mechanism, not inventing a new
one — if the postgres `tx` struct can't cleanly satisfy what slice 13 built
(for example, if slice 13's registry mechanism assumed something about
`Tx` implementations that Postgres's concrete type can't provide), that is
feedback on slice 13, not a license to design a second mechanism here. Say
so explicitly in your report. Otherwise, for anything about how the module
already binds stores to `t.q`, read `internal/modules/postgres/unit_of_work.go`
directly rather than assuming.

---

**Recommended model / effort: Opus 4.8, high effort.** This slice proves —
or fails to prove — that the new resolution mechanism preserves atomicity
against a real database. That's a correctness-critical, not mechanical,
task even though the shape itself was decided in slice 13.
