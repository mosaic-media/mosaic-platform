# Slice 20 — SDK extraction readiness

> This is MEG-015 §12's own slice 14 ("SDK extraction readiness"), shifted
> to local position 20. The final slice of the Platform foundation build
> sequence.

---

## 1. Before you start

Read `CLAUDE.md` in full, especially the "Reference capability path" entry.
Confirm **Slice 19 — Reference capability path** is `[x]` and specifically
that its exit criteria were met without a private import — if slice 19
ended in a reported blocker rather than a working capability, this slice
**cannot proceed**; stop and report, since MEG-015 §12's Stop Point is
explicit that the contracts are not ready to generate or publish until the
reference capability actually clears it.

## 2. What this slice implements

- [MEG-015 §12 — Build Sequence](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/12-build-sequence.md) — the "SDK readiness" row: *"Import boundaries are enforced and the promoted `contracts/platform/v1` surface is confirmed to expose no private Platform internals."*
- [MEG-015 §02 — Repository Layout](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/02-repository-layout.md) — Public Surface Control: *"Only `contracts/platform/v1` may be treated as a candidate public contract source... Before SDK generation exists, tests should enforce this with import checks."*
- [MEG-015 §11 — Test Gates](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/guides/meg-015-platform-foundation-implementation/11-test-gates.md) — Import boundary gate: *"Modules and transports cannot import private Platform internals."*
- [MIP-004 — Platform–SDK Contract Protocol](https://github.com/mosaic-media/mosaic-architecture/blob/main/docs/engineering/protocols/mip-004-platform-sdk-contract-protocol/index.md) — read this now, in full, before deciding what "ready to generate" actually requires; this repo has not needed it until this slice.

## 3. Scope — what to build

This slice verifies and hardens, it does not add new capability surface.

1. **Audit every file under `contracts/platform/v1`** (populated in slice
   19) for any import of an `internal/` package, any reference to a
   PostgreSQL/pgx type, or any other leak of a private implementation
   detail. Fix anything found — the promoted surface must be genuinely
   self-contained.
2. **Add or strengthen a repo-wide static check** — following the existing
   `go/parser`-based boundary-test convention already used in
   `internal/transport/graphql/boundary_test.go` and
   `internal/transport/health` — that fails if anything outside
   `contracts/platform/v1` that claims to be "public-surface-facing" (the
   reference capability from slice 19, and `contracts/platform/v1` itself)
   imports any `internal/` package. Verify it the same way earlier boundary
   tests were verified: temporarily introduce a private import, confirm the
   check fails, then revert.
3. **Confirm `ContractID`/`ContractVersion`** (established in the original
   Core contracts slice) are exposed through `contracts/platform/v1` and
   are the version identity MIP-004 expects for compatibility checks — this
   slice does not need to implement MIP-004's full compatibility protocol,
   only confirm the metadata it depends on is present and correctly
   surfaced.
4. **Write up, in CLAUDE.md, exactly what is and is not in
   `contracts/platform/v1` as of this slice** — this becomes the
   authoritative record of the Platform's first candidate public contract
   surface for whoever picks up actual SDK generation next; do not leave
   that surface undocumented.

## 4. Explicitly out of scope

- Do not generate an actual SDK package, CLI, or codegen tooling — MEG-015
  §01 explicitly excludes "the full public SDK" from this build's scope;
  this slice confirms readiness, it doesn't do the generation.
- Do not add new capabilities or expand the reference capability from
  slice 19 — if it under-proved something, that's slice 19 needing more
  work, not something to patch here.
- Do not implement MIP-004's full compatibility-checking machinery — only
  confirm the metadata it needs already exists and is exposed correctly.

## 5. Exit criteria

MEG-015 §12: *"Import boundaries are enforced and the promoted
`contracts/platform/v1` surface is confirmed to expose no private Platform
internals."*

Concrete test: the boundary check from §3.2 passes and was verified against
a deliberately introduced violation; a full audit of
`contracts/platform/v1`'s actual import declarations (not a text grep —
`go/parser`, matching this repo's established convention) shows zero
`internal/` imports; the reference capability from slice 19 still builds
and passes its own tests against this now-audited surface.

## 6. Before declaring this slice done

Run `go build ./...`, `go vet ./...`, `go test ./... -race`. Report
results. Add a new `- [x]` entry to CLAUDE.md for this slice — this is the
last row in MEG-015 §12's build sequence, so this checklist entry also
marks the Platform foundation build itself as complete against MEG-015.
State that plainly in the writeup.

## 7. If anything is ambiguous

If the audit in §3.1 finds a leak that can't be fixed without reopening a
design decision from slices 13–19 (for example, if `StorageAdapter`'s
promoted shape turns out to reference a Postgres-specific type), that is
not this slice's decision to make alone — report it as a finding against
the earlier slice responsible, the same way slice 19 was instructed to
report rather than force a private import. Otherwise, for anything about
what MIP-004 actually expects from contract metadata, read it directly
rather than guessing what "SDK-ready" should mean.

---

**Recommended model / effort: Sonnet 5, high effort.** This is
well-specified verification and audit work — not mechanical scaffolding,
since it requires genuinely careful review of every file for leaks, but
also not touching transaction atomicity or a live architectural ambiguity,
since slice 19 already resolved the one ambiguity this build sequence had
left. Opus is not needed here.
