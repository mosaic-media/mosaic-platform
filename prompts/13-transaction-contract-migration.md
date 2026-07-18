# Slice 13 — Transaction contract migration (MAD-001)

> Not one of MEG-015 §12's original 14 named slices — a remediation slice inserted
> ahead of "Reference capability path" after MAD-001 corrected the transaction
> contract shape. In this repo's local build order it now sits at position 13;
> MEG-015 §12's own "Reference capability path" and "SDK extraction readiness"
> shift to local positions 14 and 15 accordingly. Do not renumber or re-open any
> already-completed slice over this — CLAUDE.md's checklist entries for slices 1–12
> stay as they are; this is new, additive work.

---

## 1. Before you start

Read `CLAUDE.md` in full, especially the "Transaction contract shape correction
(MAD-001)" section and the status checklist. Confirm every one of the following
is marked `[x]` done — this migration touches code from nearly all of them, and
if any is not actually done, the fakes/handlers you're about to migrate won't
exist yet or won't be what this prompt assumes:

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

If any of these is not `[x]` in CLAUDE.md as you read it, **stop and report** —
do not start this slice.

## 2. What this slice implements

- [MEG-015 §03 — Platform Contracts](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md), specifically the **Contract Shape** section (the sealed `Tx` / `Store[T]` example) and the **Storage Extensibility Boundary** section.
- [MEG-015 §12 — Build Sequence](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/12-build-sequence.md), specifically **"Contract Promotion Within Slice 13"** and **"Storage Contract Correction"** (these tell you what belongs in *this* slice vs. the next one — read them carefully, the boundary matters).
- [MAD-001 — Transactional Store Extensibility](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/index.md) — read the whole record, not a summary. `02-decision.md` and `05-implementation-implications.md` are the most load-bearing for what to build; `03-alternatives-considered.md` explains why the closed interface and the `any`-based extension registry were both rejected, which matters for the design point flagged in §5 below.
- Confirm [MEG-015 §04 — Application Boundaries](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/04-application-boundaries.md) is unchanged (it is — MAD-001 §04 Consequences notes §04 was already written abstractly and needed no edit). The eight-step command order does not change; only how a command handler obtains a store changes.

## 3. Scope — what to build

This is a **mechanical, behavior-preserving migration**, not a redesign. Every
command handler must do exactly what it does today, through a different
resolution mechanism. Re-derive the current shape yourself before touching
anything — don't trust a description of it from an earlier planning session;
read the actual files listed below.

1. **`internal/platform/contracts/unit_of_work.go`** — replace the closed `Tx`
   interface (currently `Users() UserStore`, `Sessions() SessionStore`,
   `Permissions() PermissionStore`, `Config() ConfigStore`, `Outbox() EventOutbox`,
   `Credentials() CredentialStore`) with a sealed, opaque `Tx` per MEG-015 §03's
   example (an unexported marker method, no accessors). Add the package-level
   generic `func Store[T any](tx Tx) (T, error)`. `UnitOfWork.WithinTx`'s
   signature does not change.

2. **Design and add a `StorageAdapter` port** in `internal/platform/contracts/`
   (new file). MAD-001 describes its *responsibility* — "provides the
   `UnitOfWork` and binds each resolved store to the live transaction" — but
   does not give a Go signature. This is the one genuinely open design point in
   this slice; see §7 before improvising a shape.

3. **`internal/modules/postgres/unit_of_work.go`** — the concrete `tx` struct
   (currently holding `q pgx.Tx` and six accessor methods) must implement the
   new sealed `Tx` and whatever binding mechanism `StorageAdapter` defines, so
   that `Store[UserStore](tx)`, `Store[EventOutbox](tx)`, etc. resolve to the
   same postgres store implementations bound to the same `pgx.Tx`. The
   atomicity guarantee — every resolved store sharing one `pgx.Tx`, single
   `Commit`/deferred `Rollback` — must be preserved exactly, not just
   approximately.

4. **Migrate the seven command handlers in `internal/platform/app/`** that
   currently call `tx.Foo()` inside their `WithinTx` closure onto
   `contracts.Store[contracts.FooStore](tx)`:
   `create_local_user.go`, `authenticate_local_user.go`, `revoke_session.go`,
   `set_user_status.go`, `draft_config_version.go`, `validate_config_version.go`,
   `activate_config_version.go`. The eight-step command order (§04) is
   unchanged — only the store-acquisition line changes. Decide how a `Store[T]`
   resolution failure maps to a Platform `ErrorCategory` (MEG-015 §03's table
   has no obvious single answer for "requested store type isn't registered" —
   it's a Platform-side wiring defect, not caller input, so `Internal` is the
   likely fit, but confirm against how the existing handlers already treat
   unexpected contract errors before picking one).

5. **Migrate every test double that fakes `Tx`** — there are three, not one:
   - `internal/platform/contracts/contracts_test.go` (`mockTx`, six methods,
     plus the compile-time assertion `var _ contracts.Tx = mockTx{}`)
   - `internal/platform/app/fakes_test.go` (`fakeTx`, six methods)
   - `internal/transport/graphql/fakes_test.go` (`fakeTx`, six methods)

   Each needs a `Store[T]`-compatible replacement that still returns the same
   fake store instances (`fakeUserStore`, `mockEventOutbox`, etc.) so existing
   behavioral tests keep proving what they proved before.

6. **`internal/modules/postgres/outbox_atomicity_test.go` and
   `outbox_worker_test.go`** — both call `tx.Users()`/`tx.Outbox()` directly;
   migrate to `Store[T](tx)`.

7. **`test/contract/suite.go`** — the reusable, adapter-agnostic contract-test
   harness that runs against real PostgreSQL. This is the test that actually
   proves atomicity survived the refactor; migrate its `tx.Users()`/`tx.Outbox()`
   calls to `Store[T](tx)` and confirm it still passes against a real database,
   not just against fakes.

## 4. Explicitly out of scope for this slice

- **Do not populate `contracts/platform/v1`.** MEG-015 §12's "Contract
  Promotion Within Slice 13" places that as the *first step of the Reference
  capability slice* (local position 14 in this repo, not this one). This slice
  only fixes the shape in `internal/platform/contracts` (private).
- **Do not build the notes/tags (or any) reference capability.** That is local
  slice 14's own deliverable, and it depends on this slice being done first.
- **Do not build a second storage adapter (e.g. SQLite).** MAD-001 only
  requires the port *shape* make substitution possible in principle — actually
  building a second adapter isn't asked for by any named slice.
- **Do not touch the query-only app handlers** — `get_user_by_id.go`,
  `list_users.go`, `get_active_config_version.go`, `get_config_version.go`,
  `get_roles_for_user.go`, `get_grants_for_user.go`,
  `get_effective_permissions.go`. They read through a directly-injected store
  and never go through `Tx`; they are unaffected and touching them is scope
  creep.
- **Do not touch** `internal/platform/secrets/`, `internal/platform/diagnostics/`,
  `internal/platform/runtime/`, `internal/transport/health/`, or
  `internal/platform/events/` (`Bus`/`Worker`) — none of them reference
  `contracts.Tx` today (confirmed by grep this session); if you find one that
  does, CLAUDE.md's blast-radius list was wrong and you should flag that
  explicitly rather than silently fixing it as a drive-by.

## 5. Exit criteria

MEG-015 §12's table has no row for this slice — it isn't one of the 14 named
ones, so there is no verbatim criterion to quote. The nearest authoritative
language is MAD-001 §02 (Decision) and §04 (Consequences):

> "The transaction scope exposes no fixed list of stores. Every store — Core
> Platform or capability — is resolved the same way, so nothing is privileged
> by being named on the transaction handle and adding a store never edits it."
> ... "Atomicity. One transaction, one storage adapter — unchanged."

Treat the exit criteria as: **every store is resolved through `Store[T]`, `Tx`
exposes no accessor method, and the outbox/state atomicity guarantee is
unchanged and re-proven, not just re-asserted.**

Concrete tests that would actually prove it (don't just "write some tests"):

- Re-run `test/contract/suite.go`'s existing mid-transaction-failure proof
  (`TestOutboxStateAtomicOnMidTransactionFailure` or equivalent) against real
  PostgreSQL, now going through `Store[T](tx)` instead of `tx.Users()`/
  `tx.Outbox()` — it must still show neither row persists on injected failure,
  via raw SQL bypassing the stores, exactly as it did before this migration.
- Add a mechanical check that `Tx` genuinely has no exported accessor
  surface — following the existing repo convention of `go/parser`-based
  import-boundary tests (see `internal/transport/graphql/boundary_test.go` and
  `internal/transport/health`'s equivalent) rather than trusting the interface
  definition by eyeball. A test that fails if someone adds `Users()` back to
  `Tx` is the actual proof that "adding a store never edits this interface"
  holds structurally, not just today.
- Every one of the seven migrated command handlers' existing tests (the ones
  proving policy denial doesn't mutate state, atomic persistence, etc. from
  their original slices) must still pass unmodified in intent — same
  assertions, different plumbing underneath.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, and `go test ./... -race` (against real
PostgreSQL — embedded-postgres by default, or `MOSAIC_TEST_POSTGRES_DSN` for
docker-compose, per the existing `test/contract/` README). Report the actual
results. Update CLAUDE.md's status checklist: add a new `- [x]` entry for this
slice (between Supervisor handoff and Reference capability path), and update
the existing "Reference capability path" bullet to note that its blocker is
now cleared on the code side, not just the docs side.

## 7. If anything is ambiguous

The `StorageAdapter` port shape (§3.2) is the one place this prompt cannot
hand you a finished answer — MEG-015 §03 and MAD-001 describe its
responsibility, not its Go signature. Before inventing one:

- Read MAD-001 `03-alternatives-considered.md` again for *why* option (a), the
  `any`-keyed extension registry, was rejected ("lost at the call site... a
  store must ask permission to join the transaction"). Whatever internal
  mechanism you choose must not reintroduce that pattern at a boundary a
  capability author would ever touch — it's fine for the *concrete Postgres
  `tx` type* to use an internal registry keyed by `reflect.Type` or similar
  under the hood, as long as `Store[T]`'s caller-facing signature stays fully
  typed with no cast.
- Read [MEG-004 — Hexagonal Architecture §04 — Driven Ports](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-004-hexagonal-architecture/04-driven-ports.md) (cited in MAD-001's own Required Reading) for the general guidance on generic infrastructure ports that shaped that rejection.
- Look at `internal/modules/postgres/unit_of_work.go` as it exists **before**
  you touch it, to see exactly how store binding to a live `pgx.Tx` works
  today — the new mechanism must preserve that binding, not redesign it.

If, after that, the shape is still genuinely ambiguous, make the smallest
internal (unexported) decision that satisfies: type-safe call sites, one
shared transaction handle, and no accessor exposed on the public `Tx`
interface — then document the choice in the CLAUDE.md slice writeup the same
way earlier slices documented first-cut decisions under similar
underspecification (e.g. the Secret broker slice's KDF choice, the
Transactional outbox slice's redaction taxonomy). Do not guess silently.

---

**Recommended model / effort: Opus 4.8, high effort.** This slice touches
transaction atomicity directly (the entire point is preserving the
single-`pgx.Tx`-per-`WithinTx` guarantee while changing how stores are
obtained) and requires resolving a genuine, not-fully-specified architectural
design point (§7) rather than following a fully pinned-down shape. Both
conditions independently call for Opus 4.8 at high effort per the standing
model-selection rule of thumb.
