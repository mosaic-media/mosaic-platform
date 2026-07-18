# Slice 13 — Transaction contract shape: expand (MAD-001)

> Not one of MEG-015 §12's original 14 named slices — a remediation slice
> inserted ahead of "Reference capability path" after MAD-001 corrected the
> transaction contract shape. It is split into two local slices using an
> expand/contract migration, because sealing `Tx` in one step would break
> every existing caller in the same commit: **this slice (13) expands** —
> adds the new resolution mechanism alongside the existing one, touching
> nothing else — and **slice 14 contracts** — migrates every caller and then
> seals `Tx`. MEG-015 §12's own "Reference capability path" and "SDK
> extraction readiness" shift to local positions 15 and 16 accordingly. Do
> not renumber or re-open any already-completed slice over this — CLAUDE.md's
> checklist entries for slices 1–12 stay as they are; this is new, additive
> work.

---

## 1. Before you start

Read `CLAUDE.md` in full, especially the "Transaction contract shape
correction (MAD-001)" section and the status checklist. Confirm every one of
the following is marked `[x]` done — this slice's proof depends on the real
Postgres adapter and the shared contract-test suite both already existing and
passing:

- Repository scaffold
- Core contracts
- Application service skeleton
- Identity, sessions and policy
- PostgreSQL adapter and migrations
- Transactional outbox
- In-process Event Bus
- Configuration versioning and reload classes
- Secret broker
- GraphQL command and query surface
- Diagnostics and health
- Supervisor handoff

If any of these is not `[x]` in CLAUDE.md as you read it, **stop and report**
— do not start this slice.

## 2. What this slice implements

- [MEG-015 §03 — Platform Contracts](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md), specifically the **Contract Shape** section (the sealed `Tx` / `Store[T]` example — this slice builds `Store[T]`, but does *not* seal `Tx` yet, see §4) and the **Storage Extensibility Boundary** section.
- [MEG-015 §12 — Build Sequence](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/12-build-sequence.md), specifically **"Storage Contract Correction"**.
- [MAD-001 — Transactional Store Extensibility](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/index.md) — read the whole record. `02-decision.md` and `05-implementation-implications.md` are most load-bearing here; `03-alternatives-considered.md` explains why the closed interface and the `any`-based extension registry were both rejected, which matters directly for the design point in §7.
- MEG-015 §04 — Application Boundaries is unchanged and untouched by this slice; nothing in `internal/platform/app/` changes yet.

## 3. Scope — what to build

This slice adds a new, independently provable resolution mechanism without
touching a single existing caller. Re-derive the current shape yourself from
the actual files before starting — don't trust a description of it from an
earlier planning session.

1. **`internal/platform/contracts/`** — add `func Store[T any](tx Tx) (T, error)`
   (new file, e.g. `store.go`) and a `StorageAdapter` port (new file, e.g.
   `storage_adapter.go`). **Do not remove or modify `Tx`'s six existing
   methods** (`Users()`, `Sessions()`, `Permissions()`, `Config()`,
   `Outbox()`, `Credentials()`) in this slice — `Tx` is not sealed yet; that
   is slice 14's job. `Store[T]` must resolve, for each of the six known
   store types, to the *exact same* store instance a given `Tx` value's
   existing named accessor would return for that same underlying
   transaction — not a look-alike constructed separately, the same bound
   instance — so behavioral identity can be proven, not assumed. See §7 for
   the open design question this raises.

2. **`internal/modules/postgres/unit_of_work.go`** — give the concrete `tx`
   struct whatever internal mechanism makes `Store[T]` resolve correctly
   against it, without changing its six existing methods or the store
   implementations they already return. Per MAD-001 §02 ("a `StorageAdapter`
   provides the `UnitOfWork`"), consider whether the existing
   `NewUnitOfWork(pool)` constructor should now be reached through a
   `StorageAdapter`-shaped type — but do not rewire `main.go` or any
   composition-root call site beyond what's needed to keep the module
   compiling; that wiring is not required by this slice's exit criteria.

3. **`test/contract/suite.go`** — add (do not replace) a test path that
   exercises the same atomicity proof this suite already makes (state write
   + outbox append commit together; injected mid-transaction failure rolls
   both back, verified via raw SQL bypassing the stores) but going through
   `Store[UserStore](tx)` / `Store[EventOutbox](tx)` instead of
   `tx.Users()`/`tx.Outbox()`. The existing accessor-based tests in this file
   stay exactly as they are.

4. **Add a focused test proving equivalence**, not just independent
   correctness: within one `WithinTx` call, a write made through
   `Store[UserStore](tx)` must be visible to a read made through
   `tx.Users()` (and vice versa), and both must roll back together on
   failure. This is the evidence that `Store[T]` is a strict addition over
   the same transaction, not a second, divergent implementation that happens
   to also talk to Postgres.

## 4. Explicitly out of scope for this slice

- **Do not touch any of the seven command handlers** in `internal/platform/app/`
  (`create_local_user.go`, `authenticate_local_user.go`, `revoke_session.go`,
  `set_user_status.go`, `draft_config_version.go`, `validate_config_version.go`,
  `activate_config_version.go`) — they keep calling `tx.Foo()` exactly as
  they do today. Migrating them is slice 14.
- **Do not touch any of the three `Tx` fakes** —
  `internal/platform/contracts/contracts_test.go`'s `mockTx`,
  `internal/platform/app/fakes_test.go`'s `fakeTx`,
  `internal/transport/graphql/fakes_test.go`'s `fakeTx`. Slice 14.
- **Do not touch** `internal/modules/postgres/outbox_atomicity_test.go` or
  `outbox_worker_test.go` — they keep using `tx.Users()`/`tx.Outbox()`.
  Slice 14.
- **Do not remove `Tx`'s six accessor methods.** Sealing `Tx` is slice 14's
  job specifically, once every caller has moved off them — doing it here
  breaks the entire repo's build in one commit.
- **Do not populate `contracts/platform/v1`** — that's the first step of
  local slice 15 (Reference capability path), per MEG-015 §12's "Contract
  Promotion Within Slice 13."
- **Do not build the reference capability itself, or a second storage
  adapter (e.g. SQLite).** Not asked for by this slice; the port shape only
  needs to make a second adapter *possible*, not exist.

Leaving every one of those files untouched is itself part of the proof this
slice is purely additive — if you find yourself editing any of them, stop
and reconsider scope before continuing.

## 5. Exit criteria

MEG-015 §12's table has no row for this slice. Nearest authoritative
language is MAD-001 §02 (Decision) and §04 (Consequences), scoped to what
must be true at the end of *this* slice specifically — not the full
end-state MAD-001 describes, which is slice 14's job too:

> "Every store — Core Platform or capability — is resolved the same way..."
> (true for the six known stores through the new path) — "Atomicity. One
> transaction, one storage adapter — unchanged."

Treat the exit criteria as: **`Store[T]` and `StorageAdapter` exist, resolve
to the exact same transaction-bound store instances the existing accessors
already provide, and the atomicity guarantee holds through the new path
against real PostgreSQL — while every existing test in the repo still passes
completely unmodified, because no existing call site has been touched.**

Concrete tests that would actually prove it:

- The new `test/contract/suite.go` path (§3.3) passes against real
  PostgreSQL, showing the same mid-transaction-failure rollback proof
  through `Store[T]` that the existing accessor-based test already shows.
- The equivalence test (§3.4) passes, showing `Store[T]` and the matching
  named accessor observe the same transaction.
- `go build ./...` and `go test ./... -race` are green **with the diff
  confined to** `internal/platform/contracts/` (new files only) and
  `internal/modules/postgres/unit_of_work.go` plus `test/contract/suite.go`
  (additive changes only) — if the diff touches anything else, scope has
  crept into slice 14's territory.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, and `go test ./... -race` (against
real PostgreSQL — embedded-postgres by default, or
`MOSAIC_TEST_POSTGRES_DSN` for docker-compose, per the existing
`test/contract/` README). Report the actual results. Update CLAUDE.md's
status checklist: add a new `- [x]` entry for this slice between Supervisor
handoff and slice 14, and note explicitly that `Tx` is *not yet sealed* —
the old six-method surface and the new `Store[T]` path now coexist, proven
equivalent, pending slice 14's cutover.

## 7. If anything is ambiguous

The mechanism by which `Store[T]` resolves against a `Tx` value it doesn't
control the shape of is the open design point. MEG-015 §03 and MAD-001
describe the *responsibility* (uniform, typed resolution) but not the exact
Go mechanics, and here that's compounded by the expand-phase constraint that
you cannot change `Tx`'s existing method set to help. Before inventing a
mechanism:

- Read MAD-001 `03-alternatives-considered.md` again for *why* option (a),
  the `any`-keyed extension registry, was rejected ("lost at the call
  site... a store must ask permission to join the transaction"). Whatever
  internal mechanism you choose must not reintroduce that pattern at any
  boundary a caller of `Store[T]` would ever see — it is fine for the
  *concrete Postgres `tx` type* to carry an internal, unexported registry
  under the hood (keyed by `reflect.Type` or similar) as long as `Store[T]`'s
  caller-facing signature stays fully typed with no cast escaping to the
  caller.
- Read [MEG-004 — Hexagonal Architecture §04 — Driven Ports](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-004-hexagonal-architecture/04-driven-ports.md) (cited in MAD-001's own Required Reading) for the general guidance on generic infrastructure ports that shaped that rejection.
- Look at `internal/modules/postgres/unit_of_work.go` as it exists **before**
  you touch it, to see exactly how the six existing accessors construct
  their store instances from `t.q` — whatever registry you build must be
  populated from those same constructions, not parallel ones, or the
  equivalence test in §3.4 will be proving the wrong thing.

If, after that, the shape is still genuinely ambiguous, make the smallest
internal (unexported) decision that satisfies: type-safe call sites, the
same underlying transaction handle, and zero change to `Tx`'s existing
public surface — then document the choice in the CLAUDE.md slice writeup the
same way earlier slices documented first-cut decisions under similar
underspecification (e.g. the Secret broker slice's KDF choice). Do not guess
silently.

---

**Recommended model / effort: Opus 4.8, high effort.** This slice is where
the atomicity-critical and genuinely ambiguous work actually lives — proving
a new resolution mechanism is transaction-equivalent to the existing one,
against real PostgreSQL, with no fully-specified Go mechanism handed to you.
Slice 14, by contrast, is a mechanical call-site sweep against a shape this
slice will have already pinned down — that one should run on Sonnet 5 at
medium effort. Don't apply this slice's model/effort tier to that one; they
are deliberately different kinds of work.
