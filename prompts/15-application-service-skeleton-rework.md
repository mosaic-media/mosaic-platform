# Slice 15 — Application service skeleton rework (MAD-001, expand step 3 of 6)

> Third of six rework slices. Mechanical: migrate the one command handler
> this original slice introduced, and — since this is the first app-layer
> slice in the rework sequence — extend the shared app-package fake.

---

## 1. Before you start

Read `CLAUDE.md`. Confirm **Slice 13 — Core contracts rework** and **Slice
14 — PostgreSQL adapter rework** are both `[x]`; read slice 13's writeup for
the exact `Store[T]` shape. If either is missing, **stop and report**.

## 2. What this slice implements

- [MEG-015 §04 — Application Boundaries](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/04-application-boundaries.md) — the eight-step command order is unchanged; only step 5's store-acquisition mechanism changes.
- [MAD-001 §04 — Consequences](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/04-consequences.md): *"Every `tx.Users()`-style call site... moves to uniform resolution. The change is contained to how stores are obtained, not what they do."*

## 3. Scope — what to build

1. Migrate `internal/platform/app/create_local_user.go` from `tx.Users()`/
   `tx.Outbox()`/`tx.Credentials()` (confirm the exact set it currently
   calls — it gained credential persistence in the Identity slice) to
   `contracts.Store[contracts.FooStore](tx)`. Command order and behavior
   are unchanged; only the acquisition line changes. Decide the
   `ErrorCategory` for a `Store[T]` resolution failure — check MEG-015 §03's
   Error Categories table and how this handler already treats an unexpected
   contract error before picking one (`Internal` is the likely fit for a
   Platform-side wiring defect, but confirm rather than assume).
2. `internal/platform/app/fakes_test.go`'s `fakeTx` is shared by every test
   in the `app` package (`service_test.go`, `config_test.go`,
   `config_queries_test.go`, `users_and_permissions_test.go`) — confirm this
   is still true before editing. **Extend** it (don't replace) to satisfy
   `Store[T]` resolution the same way slice 13/14 did, backed by the same
   fake store instances its six methods already return. This is a one-time
   extension: slices 16 and 17 will depend on it already being done and
   should not need to touch this file again.

## 4. Explicitly out of scope

- Nothing in `internal/platform/app/authenticate_local_user.go`,
  `revoke_session.go` (slice 16), or the config version handlers (slice 17).
- Nothing in `internal/modules/postgres/` or `internal/transport/graphql/`.
- Do not remove `Tx`'s six accessor methods.
- Do not populate `contracts/platform/v1` or build the reference capability.

## 5. Exit criteria

No MEG-015 §12 table row exists for this slice. Scoped exit criteria:
**`create_local_user.go` resolves its stores through `Store[T]`; the shared
app-package fake supports `Store[T]` resolution; every existing app-package
test still passes with unchanged assertions.**

Concrete test: `TestCreateLocalUserFollowsCommandBoundaryOrder`,
`TestCreateLocalUserRejectsDuplicateUsernameAndRollsBack`,
`TestCreateLocalUserThenAuthenticateSucceeds`, and every other test in
`service_test.go` that exercises `CreateLocalUser` pass unmodified in
intent — same assertions, different plumbing underneath.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, `go test ./... -race`. Report results.
Add a new `- [x]` entry to CLAUDE.md between slice 14 and slice 16, and
note explicitly whether `fakes_test.go` was extended here (it should be —
confirm for slices 16/17's benefit).

## 7. If anything is ambiguous

This should be purely mechanical. If `create_local_user.go` or
`fakes_test.go` don't fit slice 13/14's mechanism cleanly, that's feedback
on those slices — report it rather than inventing a workaround here.

---

**Recommended model / effort: Sonnet 5, medium effort.** One handler, one
shared fake extension, against a fully specified shape.
