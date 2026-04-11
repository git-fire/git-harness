# Spec 005 — Harden, Tag, and Publish git-harness

**Branch:** `spec/005-harden-and-release`
**Depends on:** Spec 004 — all consumers wired and green
**Status:** Ready for implementation after 004

## Goal
Add CI to git-harness, tag v0.1.0, publish, remove `replace` directives from consumers,
and do a final green check across all three repos.

## Steps

### Step 1 — CI for git-harness
Create `.github/workflows/ci.yml` in git-harness:
```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go build ./...
      - run: go vet ./...
      - run: go test -race -coverprofile=coverage.out ./...
```
Commit: `ci: add GitHub Actions for git-harness`

### Step 2 — Tag v0.1.0
```bash
cd git-harness
git tag v0.1.0
git push origin v0.1.0
```

### Step 3 — Remove replace directives
In `git-fire/go.mod` and `git-fcuk/go.mod`:
- Remove the `replace github.com/git-fire/git-harness => ...` lines
- Run: `go get github.com/git-fire/git-harness@v0.1.0` in each consumer repo
- Run: `go mod tidy` in each consumer repo

### Step 4 — Final green check
```bash
# In git-fire:
go build ./... && go test ./... && go vet ./...

# In git-fcuk:
go build ./... && go test ./... && go vet ./...

# In git-harness:
go build ./... && go test ./... && go vet ./...
```
All three must be fully green.

### Step 5 — Commit consumers
```
chore(git-harness): pin to published git-harness v0.1.0
```
One commit per consumer repo.

## Acceptance Criteria
- git-harness CI workflow exists and passes on push
- `v0.1.0` tag exists on git-harness remote
- No `replace` directives remain in git-fire or git-fcuk go.mod
- `go test ./...` green in all three repos
- `go vet ./...` clean in all three repos
