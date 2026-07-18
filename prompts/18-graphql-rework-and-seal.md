# Slice 18 — GraphQL rework and seal (MAD-001, contract step 6 of 6)

> Last of six rework slices, and the only one that seals `Tx`. Slices 13–17
> expanded: `Store[T]`/`StorageAdapter` were added and every other
> `tx.Foo()`-calling command handler was migrated onto them, while `Tx`'s
> six accessor methods stayed in place throughout. This slice migrates the
> one remaining handler and fake, confirms nothing else in the repository
> still depends on the old accessors, then removes them — completing MAD-001
> in code. After this slice, "Reference capability path" (local slice 19)
> and "SDK extraction readiness" (local slice 20) are the only work left.

---

## 1. Before you start

Read `CLAUDE.md`. Confirm all five prior rework slices are `[x]`: **13
(Core contracts rework)**, **14 (PostgreSQL adapter rework)**, **15
(Application service skeleton rework)**, **16 (Identity, sessions and
policy rework)**, **17 (Configuration versioning rework)**. This slice
depends on every one of them — it is the only place all their work
converges. If any is missing, **stop and report**.

## 2. What this slice implements

- [MEG-015 §03 — Platform Contracts](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md) — the sealed `Tx` shape becomes real in this slice, not just documented.
- [MEG-015 §09 — GraphQL and Diagnostics](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/09-graphql-and-diagnostics.md) — resolvers call services only; confirm this slice's `setUserStatus` resolver still does nothing but call `app.SetUserStatus`.
- [MAD-001 §04 — Consequences](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/04-consequences.md): *"Every `tx.Users()`-style call site across the thirteen built slices moves to uniform resolution."* This slice is where that sentence becomes completely true.
- [MEG-015 §12 — Build Sequence](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/12-build-sequence.md) — re-read "Contract Promotion Within Slice 13" and "Storage Contract Correction" so you don't pull local-slice-19 work forward into this one.

## 3. Scope — what to build

1. Migrate `internal/platform/app/set_user_status.go` from `tx.Foo()` to
   `contracts.Store[contracts.FooStore](tx)`.
2. Migrate `internal/transport/graphql/fakes_test.go`'s `fakeTx` (six
   methods today) to satisfy `Store[T]` resolution the same way the other
   fakes were extended in slices 13/15, returning the same fake store
   instances it already constructs.
3. **Confirm zero remaining callers.** Run `go build ./...` and search the
   whole repository for `.Users(`, `.Sessions(`, `.Permissions(`,
   `.Config(`, `.Outbox(`, `.Credentials(` called on any `contracts.Tx`
   value. The only matches should be the six method *definitions* on `Tx`
   itself and on its implementations (postgres `tx`, `mockTx`, the two
   `fakeTx`s). If any caller remains, migrate it before continuing — do not
   seal with a live caller still depending on the old accessors.
4. **Seal `Tx`.** Remove the six accessor methods from
   `internal/platform/contracts/unit_of_work.go`, leaving the marker-only
   shape from MEG-015 §03's example. Remove the now-redundant six-method
   implementations from `internal/modules/postgres/unit_of_work.go`'s `tx`
   struct, `contracts_test.go`'s `mockTx`, `app/fakes_test.go`'s `fakeTx`,
   and `graphql/fakes_test.go`'s `fakeTx` — each should already resolve
   everything through `Store[T]` internally, so this is deletion of dead
   code, not new work.
5. Add a structural boundary test proving `Tx` has no exported methods,
   following this repo's existing `go/parser`-based convention (see
   `internal/transport/graphql/boundary_test.go` and
   `internal/transport/health`'s equivalent). Verify it the same way those
   were verified: temporarily add a `Users()` method back to `Tx`, confirm
   the new test fails, remove it again before committing.

## 4. Explicitly out of scope

- **Do not populate `contracts/platform/v1`.** MEG-015 §12 places that as
  the first step of local slice 19 (Reference capability path), which comes
  after this one.
- **Do not build the reference capability itself, or a second storage
  adapter.**
- **Do not touch the query-only app handlers** (`get_user_by_id.go`,
  `list_users.go`, `get_active_config_version.go`, `get_config_version.go`,
  `get_roles_for_user.go`, `get_grants_for_user.go`,
  `get_effective_permissions.go`) — confirm they still bypass `Tx` entirely;
  don't add work here if so.
- **Do not redesign anything slices 13–17 already decided.** If sealing
  reveals one of them got something wrong (a caller that doesn't fit the
  mechanism, a fake that can't cleanly drop its six methods), report that
  explicitly rather than quietly patching around it.

## 5. Exit criteria

No MEG-015 §12 table row exists for this slice, but this is where MAD-001's
decision fully lands. Near-verbatim from MAD-001 §02: *"The transaction
scope exposes no fixed list of stores. Every store — Core Platform or
capability — is resolved the same way, so nothing is privileged by being
named on the transaction handle and adding a store never edits it."*

Scoped exit criteria: **`Tx` is sealed (no exported accessor methods),
every call site in the repository resolves stores through `Store[T]`, and
every test suite from every earlier slice — atomicity, policy denial,
config activation, auth, GraphQL routing — still passes with unchanged
intent.**

Concrete tests:
- Full `go test ./... -race` green against real PostgreSQL, with no test's
  *assertions* changed from what they proved before this migration.
- The structural boundary test (§3.5) passes, and was verified to actually
  fail when `Users()` is temporarily added back to `Tx`.
- A repo-wide grep for `tx.Users(`, `tx.Sessions(`, `tx.Permissions(`,
  `tx.Config(`, `tx.Outbox(`, `tx.Credentials(` returns nothing.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, `go test ./... -race` against real
PostgreSQL. Report results. Add a new `- [x]` entry to CLAUDE.md between
slice 17 and "Reference capability path," and update the existing
"Reference capability path" bullet to state plainly that its blocker is now
cleared on the code side, not just the docs side — `contracts/platform/v1`
can now be populated with a `Tx`/`Store[T]`/`StorageAdapter` surface that
has no private-internals dependency baked into its shape.

## 7. If anything is ambiguous

By this point ambiguity should mean one of slices 13–17 left something
inconsistent, not that this slice needs to make a new design call. If step
4 (deleting the six-method implementations) doesn't cleanly fall out of
what those slices built, report the inconsistency rather than resolving it
silently. For the `ErrorCategory` mapping on `set_user_status.go`, check
[MEG-015 §03's Error Categories table](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md) and how slice 15/16/17 already
handled the same question before picking independently.

---

**Recommended model / effort: Sonnet 5, medium effort**, with a bump to
**high** if step 3's zero-remaining-callers confirmation turns up anything
unexpected, or if deleting the dead six-method implementations in step 4
doesn't fall out cleanly — both would indicate a real problem worth careful
attention rather than a mechanical fix.
