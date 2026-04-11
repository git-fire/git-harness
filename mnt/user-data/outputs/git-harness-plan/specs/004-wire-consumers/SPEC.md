# Spec 004 — Wire Consumers to git-harness

**Branch:** `spec/004-wire-consumers`
**Depends on:** Spec 003 — git-harness builds and tests pass
**Status:** Ready for implementation after 003

## Goal
Update `git-fire` and `git-fcuk` to import from `git-harness` instead of their local internal packages.
Delete the now-redundant local copies. All tests must remain green throughout.

## Step-by-step (must follow this order)

### Step 1 — Rewrite imports in git-fire
For every file importing an `EXTRACT`-classified internal package:
- Replace `github.com/git-fire/git-fire/internal/<pkg>` → `github.com/git-fire/git-harness/<pkg>`
- Do not touch `KEEP` packages

Run: `go build ./... && go test ./...`
**Must be green before Step 2.**
Commit: `feat(git-harness): consume git-harness in git-fire`

### Step 2 — Rewrite imports in git-fcuk
Same as Step 1 for git-fcuk.
Run: `go build ./... && go test ./...`
**Must be green before Step 3.**
Commit: `feat(git-harness): consume git-harness in git-fcuk`

### Step 3 — Delete extracted code from git-fire
Only after Steps 1 and 2 are green:
- Delete `git-fire/internal/<pkg>` directories that are fully replaced by git-harness
- Run `go build ./... && go test ./...`
Commit: `refactor(git-harness): remove extracted internals from git-fire`

### Step 4 — Delete extracted code from git-fcuk (if applicable)
Same as Step 3 for git-fcuk.
Commit: `refactor(git-harness): remove extracted internals from git-fcuk`

## Acceptance Criteria
- `go test ./...` green in git-fire ✓
- `go test ./...` green in git-fcuk ✓
- No remaining imports of deleted internal packages in either repo
- `go vet ./...` clean in both repos

## Do NOT do in this spec
- Do not remove `replace` directives yet (that's Spec 005)
- Do not tag git-harness yet (that's Spec 005)
