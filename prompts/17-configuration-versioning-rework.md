# Slice 17 — Configuration versioning rework (MAD-001, expand step 5 of 6)

> Fifth of six rework slices. Mechanical: migrate the three command handlers
> this original slice introduced.

---

## 1. Before you start

Read `CLAUDE.md`. Confirm slices **13 (Core contracts rework)**, **14
(PostgreSQL adapter rework)**, **15 (Application service skeleton rework)**,
and **16 (Identity, sessions and policy rework)** are all `[x]`. If any is
missing, **stop and report**.

## 2. What this slice implements

- [MEG-015 §04 — Application Boundaries](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/04-application-boundaries.md) — command order unchanged.
- [MEG-015 §08 — Configuration and Secrets](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/08-configuration-and-secrets.md) — confirm the Draft→Validated→Active→Superseded state machine and reload-class classification are unaffected in behavior; only store acquisition changes.
- [MAD-001 §04 — Consequences](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/04-consequences.md).

## 3. Scope — what to build

1. Migrate `internal/platform/app/draft_config_version.go`,
   `validate_config_version.go`, and `activate_config_version.go` from
   their current `tx.Config()`/`tx.Outbox()` calls to
   `contracts.Store[contracts.FooStore](tx)`. `ActivateConfigVersion`'s
   Hot-only-hot-applies / Generation-class-deferred logic (from the original
   Configuration versioning slice) does not change — only how it obtains
   the `ConfigStore` and `EventOutbox` inside its transaction.
2. `internal/platform/app/fakes_test.go` should already support `Store[T]`
   from slice 15 — confirm rather than assume; extend only if it genuinely
   doesn't.

## 4. Explicitly out of scope

- Nothing in `internal/transport/graphql/` (slice 18) — note that
  `GetActiveConfigVersion`/`GetConfigVersion` are query-only and already
  bypass `Tx` entirely; do not touch them.
- `internal/platform/config/`'s `Manager` (`Draft`/`Validate`/`Activate`)
  itself — it takes a `ConfigStore` parameter directly, not through `Tx`,
  and is unaffected. Confirm this is still true; do not touch it if so.
- Do not remove `Tx`'s six accessor methods.

## 5. Exit criteria

No MEG-015 §12 table row exists for this slice. Scoped exit criteria:
**all three config version command handlers resolve their stores through
`Store[T]`; the full Draft→Validated→Active→Superseded round trip and every
policy-denial/rejection test still pass with unchanged assertions.**

Concrete test: `TestConfigVersionDraftValidateActivateHotChange`,
`TestActivateConfigVersionRejectsGenerationClassChangeFromHotApplying`,
`TestActivateConfigVersionSupersedesPreviousActive`,
`TestValidateConfigVersionRejectsUnregisteredField`, and
`TestDraftConfigVersionDeniedByPolicyDoesNotMutateState` all pass unmodified
in intent.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, `go test ./... -race`. Report results.
Add a new `- [x]` entry to CLAUDE.md between slice 16 and slice 18.

## 7. If anything is ambiguous

Purely mechanical. Report anything that doesn't fit slice 13/14's shape as
feedback on those slices.

---

**Recommended model / effort: Sonnet 5, medium effort.** Three handlers,
against a fully specified shape and an already-extended fake.
