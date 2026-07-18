# Slice 16 — Identity, sessions and policy rework (MAD-001, expand step 4 of 6)

> Fourth of six rework slices. Mechanical: migrate the two command handlers
> this original slice introduced.

---

## 1. Before you start

Read `CLAUDE.md`. Confirm **Slice 13 — Core contracts rework**, **Slice 14
— PostgreSQL adapter rework**, and **Slice 15 — Application service
skeleton rework** are all `[x]`. Slice 15 should already have extended the
shared `internal/platform/app/fakes_test.go` fake to support `Store[T]` —
confirm that from its writeup before assuming you need to touch that file
again. If any prerequisite is missing, **stop and report**.

## 2. What this slice implements

- [MEG-015 §04 — Application Boundaries](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/04-application-boundaries.md) — command order unchanged.
- [MEG-015 §07 — Identity, Policy and Sessions](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/07-identity-policy-and-sessions.md) — confirm this slice's original scope (password login → session issuance; session revocation) is unaffected in behavior, only in store acquisition.
- [MAD-001 §04 — Consequences](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/04-consequences.md).

## 3. Scope — what to build

1. Migrate `internal/platform/app/authenticate_local_user.go` from its
   current `tx.Foo()` calls to `contracts.Store[contracts.FooStore](tx)`.
   Confirm exactly which stores it resolves today (at minimum users,
   credentials, sessions, outbox) before migrating each call site.
2. Migrate `internal/platform/app/revoke_session.go` the same way.
3. If, contrary to slice 15's expectation, `internal/platform/app/fakes_test.go`
   does not yet support `Store[T]` resolution, extend it now — but this
   should be a fallback, not the expected path; note in your report if you
   had to do this.

## 4. Explicitly out of scope

- `create_local_user.go` (slice 15, already done), the config version
  handlers (slice 17), anything in `internal/transport/graphql/` (slice 18).
- `internal/platform/sessions/` and `internal/platform/policy/` themselves —
  they take stores as constructor parameters directly, not through `Tx`,
  and are unaffected by this migration. Confirm this is still true; do not
  touch them if so.
- Do not remove `Tx`'s six accessor methods.

## 5. Exit criteria

No MEG-015 §12 table row exists for this slice. Scoped exit criteria:
**`authenticate_local_user.go` and `revoke_session.go` resolve their stores
through `Store[T]`; every existing test covering authentication and session
revocation still passes with unchanged assertions.**

Concrete test: `TestCreateLocalUserThenAuthenticateSucceeds`,
`TestAuthenticateLocalUserRejectsWrongPassword`,
`TestAuthenticateLocalUserRejectsUnknownUsername`,
`TestSessionIssuedValidatedAndRevoked`, and
`TestRevokeSessionDeniedByPolicyDoesNotMutateState` all pass unmodified in
intent.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, `go test ./... -race`. Report results.
Add a new `- [x]` entry to CLAUDE.md between slice 15 and slice 17.

## 7. If anything is ambiguous

Purely mechanical against slice 13/14's already-decided shape. If something
doesn't fit, report it as feedback on those slices rather than improvising.

---

**Recommended model / effort: Sonnet 5, medium effort.** Two handlers,
against a fully specified shape and an already-extended fake.
