# Slice 19 — Reference capability path

> This is MEG-015 §12's own slice 13 ("Reference capability path"), shifted
> to local position 19 because six rework slices (13–18 in this repo's local
> numbering) had to land first to make the transaction contract shape this
> slice depends on actually exist in code, not just in `mosaic-architecture`
> docs. Two prior attempts at this slice are already recorded in CLAUDE.md —
> read both before starting, they are not just history, they set the bar
> for what "actually proves it" means here.

---

## 1. Before you start

Read `CLAUDE.md` in full, especially the "Transaction contract shape
correction (MAD-001)" section and every existing "Reference capability
path" note. Confirm **Slice 18 — GraphQL rework and seal** is `[x]` — this
slice depends on `Tx` actually being sealed and `Store[T]`/`StorageAdapter`
existing and being used by every command handler in the repo. If it is not
`[x]`, **stop and report**.

## 2. What this slice implements

- [MEG-015 §12 — Build Sequence](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/12-build-sequence.md) — the "Reference capability" row: *"Proven contracts are promoted into `contracts/platform/v1`, then one non-media capability proves the registration path using only those packages."* Also "Contract Promotion Within Slice 13" (promotion is step one of this slice, not deferred) and "Stop Point Before SDK."
- [MEG-015 §03 — Platform Contracts](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/03-platform-contracts.md) — **Storage Extensibility Boundary** section specifically; read §7 below before assuming what it permits.
- [MEG-015 §02 — Repository Layout](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/02-repository-layout.md) — Public Surface Control (`contracts/platform/v1` is the only candidate public contract source).
- [MAD-001 — Transactional Store Extensibility](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/architecture/mad-001-transactional-store-extensibility/index.md) §04's Deferred Follow-Ups (the content-agnostic object model is *not yet built* — this matters directly, see §7) and §05's Verification section.
- [MEG-006 — Module Platform §13](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-006-module-platform/13-platform-guidelines.md) — this slice is also where the `internal/modules/`-shaped built-in-module registration pattern gets validated against something that isn't infrastructure, per the open question CLAUDE.md flagged when this slice was first attempted.

## 3. Scope — what to build

1. **Populate `contracts/platform/v1`** with the candidate public surface
   proven by slices 1–18: `UnitOfWork`, the sealed `Tx`, `Store[T]`,
   `StorageAdapter`, `ErrorCategory`/`Error`, `ContractID`/`ContractVersion`,
   `Clock`, `IDGenerator`, and `EventOutbox` (a capability needs to emit its
   own outbox events the same way Core Platform commands do). Do not
   promote `UserStore`/`SessionStore`/`PermissionStore`/`ConfigStore`/
   `CredentialStore`/`SecretBroker`/`HealthProbe` unless the capability you
   build genuinely needs them — promote what's proven necessary, not
   everything that exists.
2. **Pick one genuinely simple, non-media capability** — the two prior
   attempts used a notes/tags idea (`Note{ID,UserID,Text,Tags,CreatedAt}`);
   reusing that is fine, or pick another equally simple one (e.g. a minimal
   audit-log viewer). The choice matters less than proving the boundary
   holds.
3. **Build it as a built-in module** under `internal/modules/` (not
   `internal/adapters/` — per the tier model), registered through
   `internal/composition/builtin/` the same way the Postgres module is,
   validating that the registration pattern actually generalizes to
   non-infrastructure capabilities.
4. **Its application/command logic must import only `contracts/platform/v1`**
   — no `internal/platform/contracts`, no other `internal/platform/*`
   package. Its own storage needs go through the mechanism §7 discusses.
5. Add a static import-boundary test for this capability's own command/
   service code, following the existing `go/parser`-based convention (see
   `internal/transport/graphql/boundary_test.go`), proving it does not
   import anything under `internal/` beyond what's unavoidable for
   registration (see §7 on where the registration-vs-contracts line falls).

## 4. Explicitly out of scope

- Do not build a GraphQL surface, media/product behavior, or anything
  beyond the one capability's core command/query path.
- Do not attempt SDK generation or extraction itself — that's slice 20,
  and only after this slice's Stop Point is actually satisfied.
- Do not promote contracts into `contracts/platform/v1` "just in case" —
  each addition should trace to something this specific capability needed.

## 5. Exit criteria

MEG-015 §12: *"One non-media capability proves registration path."* Stop
Point Before SDK: *"The Platform is ready for SDK work when the reference
capability uses only candidate contract packages and no private Platform
internals. If the reference capability requires private imports, the
Platform contracts are not ready to generate or publish."*

If, after slices 13–18's work, the capability's core logic still requires a
private `internal/` import, **stop and report that as a finding** — per
MEG-015 §12, that is a real result, not a failure to paper over, exactly as
it was the first time this slice was attempted. Do not force it past with
an import that shouldn't be there.

Concrete test: the capability's command handler(s) pass a real test exercising
create/read against a real transaction (through `Store[T]` and the promoted
`EventOutbox`), and the import-boundary test from §3.5 passes — verify it
the way earlier boundary tests were verified, by temporarily adding a
private import and confirming the test fails, then reverting.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, `go test ./... -race` against real
PostgreSQL. Report results. Update CLAUDE.md's "Reference capability path"
checklist entry — mark it `[x]` if the capability builds against
`contracts/platform/v1` alone, or record the specific private import that
blocked it if not, per §5.

## 7. If anything is ambiguous

**This is the one place in this slice where you should stop and think
before writing code, not improvise:** MEG-015 §03's Storage Extensibility
Boundary says capabilities "do not define their own tables, modify Core
Platform schema or open parallel databases," and MAD-001 gestures at a
future "content-agnostic object model" that new content capabilities are
meant to map onto instead of inventing new storage — but MAD-001 §04's own
Deferred Follow-Ups section says that object model is **not yet built**
("canonising it is a separate effort... out of scope for this record").
Read MAD-001 `05-implementation-implications.md`'s "For Module Authors"
section closely: it also says *"A Module that needs a genuinely new
data-owning domain... is proposing Platform and SDK evolution and should be
treated as such"* — which is not a prohibition, it's a requirement that the
new store be added deliberately, through the reviewed candidate contract
surface, not smuggled in as a private table a capability reaches for
directly. That is almost certainly the correct reading for this slice —
your capability's own store contract (if it needs one) belongs in
`contracts/platform/v1` alongside the others, implemented by a built-in
module going through the same `UnitOfWork`/`Store[T]`/migration path as
everything else, not a bespoke table hidden inside the capability's own
package. If you land somewhere different after reading both sections
yourself, say so explicitly in your report rather than resolving the
tension silently — this is genuinely unsettled territory, not something
this prompt can hand you a final answer on.

---

**Recommended model / effort: Opus 4.8, high effort.** This slice both
touches transaction/storage design again (the capability's own persistence
path) and requires resolving the genuine architectural ambiguity in §7 —
both independently call for Opus at high effort per the standing rule.
