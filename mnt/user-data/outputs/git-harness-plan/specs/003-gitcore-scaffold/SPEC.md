# Spec 003 — git-harness Module Scaffold & Code Extraction

**Branch:** `spec/003-git-harness-scaffold`
**Depends on:** Spec 002 green
**Status:** Ready for implementation after 002

## Goal
Create `github.com/git-fire/git-harness` as a proper Go module and copy all `EXTRACT`-classified
code into it. Both repos still compile against their local copies at this stage.

## Acceptance Criteria

### Module structure
```
git-harness/
  go.mod          # module github.com/git-fire/git-harness, go 1.24
  go.sum
  README.md       # brief — "Core git primitives for the git-fire suite"
  LICENSE         # MIT
  git/            # runner, subprocess exec, error types
  safety/         # sanitize, redact, secret scanner primitives
  exec/           # executor primitives (if in EXTRACT scope)
  repo/           # repo introspection (if in EXTRACT scope)
```

### Code quality
- `go build ./...` passes in git-harness
- `go vet ./...` passes in git-harness
- `go test ./...` passes in git-harness (all tests ported from source)
- No imports of `github.com/git-fire/git-fire` or `github.com/git-fire/git-fcuk`
- No imports of Cobra, Viper, Bubble Tea, Lipgloss, or any TUI library

### Local wiring (temporary)
Add to `git-fire/go.mod`:
```
replace github.com/git-fire/git-harness => ../git-harness
```
Add to `git-fcuk/go.mod`:
```
replace github.com/git-fire/git-harness => ../git-harness
```
These are temporary and will be removed in Spec 005.

## Commit
```
feat(git-harness): scaffold module and extract primitives
```
