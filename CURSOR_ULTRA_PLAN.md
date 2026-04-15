# Ultra Plan: Extract `git-harness` from `git-fire`

## Mission
Extract reusable git primitives from `github.com/git-fire/git-fire` into a new standalone module
`github.com/git-fire/git-harness`. Consume it from both `git-fire` and `git-fcuk` without
breaking any existing behavior. Then follow the same multi-language pattern established by
`git-testkit` to ship Python and Java wrappers.

Tests are the contract — every deletion must be covered before it happens.

## Repo context (you can read all of these)
- `git-fire/git-fire` — main CLI; has `internal/git`, `internal/safety`, `internal/executor`, `internal/repo`
- `git-fire/git-fcuk` — history surgery CLI; likely duplicates git-fire primitives
- `git-fire/git-testkit` — READ THIS FIRST. Mirror its repo structure, wrapper pattern, CI setup,
  and language interface conventions exactly for git-harness.
- `git-harness` — does NOT exist yet; you will create it

## Go version floor
Align to `go 1.24` across all modules. Bump git-fcuk if it's behind.

## Naming
- Go module: `github.com/git-fire/git-harness`
- Python package: `git-harness` on PyPI (mirror git-testkit's Python wrapper structure)
- Java package: `io.gitfire.harness` on Maven (mirror git-testkit's Java wrapper structure)

---

## Phase 0 — Study git-testkit

Before writing a single line, read the following in `git-fire/git-testkit`:
- Root `go.mod` and directory structure
- How the Python wrapper is structured (bindings, build system, packaging)
- How the Java wrapper is structured (JNI/FFI or subprocess, Maven/Gradle setup)
- CI workflows (what triggers what, how wrappers are built and tested)
- Any `Makefile` or build scripts

Produce a one-paragraph summary of the pattern in your scratchpad before proceeding.
This is the canonical blueprint — git-harness must follow it exactly.

---

## Phase 1 — Audit & Extraction Candidates

### 1.1 Codebase scan
In **git-fire**, read every file under:
- `internal/git/`
- `internal/safety/`
- `internal/executor/`
- `internal/repo/`

In **git-fcuk**, read every file under:
- `internal/git/`
- `internal/repo/`
- any other non-UI, non-CLI-specific internals

### 1.2 Classify every symbol
For each exported and unexported type/function/interface, classify as:

| Class | Criteria | Destination |
|-------|----------|-------------|
| `EXTRACT` | Pure git/safety primitive; no Cobra, Viper, TUI, or app-specific config deps | `git-harness` |
| `KEEP` | References app-layer types (config structs, TUI models, CLI flags) | stays in caller |
| `BOUNDARY` | Called by both `EXTRACT` and `KEEP` code | needs adapter or interface split |

Quick filter: `grep -r "cobra\|viper\|bubbletea\|lipgloss\|charmbracelet"` — anything that matches is `KEEP`.

Pay special attention to symbols that exist in BOTH repos — these are highest-priority extractions.

Produce `specs/001-git-harness-audit/audit.md` inside git-fire. Do not move any code yet.

---

## Phase 2 — Test Coverage Before Surgery

### 2.1 Coverage baseline
Run `go test ./...` in git-fire and git-fcuk. Record coverage per package.
Any `EXTRACT`-classified package below **80% coverage** must be brought to 80% before Phase 3.

### 2.2 Write missing tests
For each under-covered `EXTRACT` package:
- Add unit tests in `<package>/<package>_test.go`
- Use `git-testkit` for any tests requiring a real git repo on disk
- Cover: happy path, error path, and edge cases for every exported function
- Do NOT alter implementation code — tests only
- Commit: `test(git-harness-prep): coverage for internal/git` (one commit per package)

### 2.3 Boundary tests
For every `BOUNDARY` symbol, write an integration test that calls through the full stack
from the CLI command down to the primitive. These are the regression guards for import rewrites.

Commit: `test(git-harness-prep): boundary integration tests`

### 2.4 Verify
`go test ./...` must be green in both repos before proceeding.

---

## Phase 3 — Create `git-harness` Go Module

### 3.1 Scaffold — mirror git-testkit's structure exactly
Read git-testkit's root layout and replicate it. At minimum:

```
git-harness/
  go.mod              # module github.com/git-fire/git-harness, go 1.24
  go.sum
  README.md
  LICENSE             # MIT, same as git-fire
  git/                # subprocess runner, errors, types
  safety/             # sanitize, redact, secret detection
  exec/               # executor primitives (if EXTRACT)
  repo/               # repo introspection helpers (if EXTRACT)
  .github/
    workflows/
      ci.yml          # Go build + test + vet
  testkit/            # polyglot layout (mirror git-testkit)
    python/
    java/
    .specify/
```

> **Local wiring (temporary):** Add `replace github.com/git-fire/git-harness => ../git-harness`
> to git-fire/go.mod and git-fcuk/go.mod. Remove before tagging v0.1.0.

### 3.2 Copy EXTRACT code
For each `EXTRACT`-classified file/symbol:
1. Copy into the appropriate `git-harness/<subpackage>/` directory
2. Update `package` declaration
3. Update any internal cross-references to use the new import path
4. Do NOT delete from origin yet

### 3.3 Verify git-harness builds and tests pass
```bash
cd git-harness && go build ./... && go test ./... && go vet ./...
```

Commit: `feat(git-harness): scaffold Go module and extract primitives`

---

## Phase 4 — Wire Go Consumers

### Step 1 — Rewrite imports in git-fire
Replace `github.com/git-fire/git-fire/internal/<pkg>` → `github.com/git-fire/git-harness/<pkg>`
for every `EXTRACT`-classified import. Leave `KEEP` imports untouched.

`go build ./... && go test ./...` must be green.
Commit: `feat(git-harness): consume git-harness in git-fire`

### Step 2 — Rewrite imports in git-fcuk
Same as Step 1.
`go build ./... && go test ./...` must be green.
Commit: `feat(git-harness): consume git-harness in git-fcuk`

### Step 3 — Delete dead code from git-fire
Only after both consumers compile and test green:
- Delete `git-fire/internal/<pkg>` directories fully replaced by git-harness
- `go build ./... && go test ./...`
Commit: `refactor(git-harness): remove extracted internals from git-fire`

### Step 4 — Delete dead code from git-fcuk (if applicable)
Same as Step 3.
Commit: `refactor(git-harness): remove extracted internals from git-fcuk`

---

## Phase 5 — Python Wrapper

> Follow git-testkit's Python wrapper structure exactly. Read it before writing anything here.

### 5.1 Scaffold
Mirror whatever structure git-testkit uses under `testkit/python/`. This likely means:
- A Python package under `testkit/python/git_harness/`
- Build tooling (cffi, ctypes, subprocess bridge, or whatever git-testkit uses)
- `pyproject.toml` / `setup.py`
- `testkit/python/README.md`

### 5.2 Implement
Expose the same surface area as the Go module — subprocess runner, safety/sanitize, repo introspection.
The Python API should feel Pythonic (snake_case, context managers where appropriate, exceptions not error tuples).

### 5.3 Tests
Mirror git-testkit's Python test setup. At minimum one test per exported function. Tests run in CI.

### 5.4 CI
Add a `python-wrapper.yml` workflow (or extend `ci.yml`) mirroring git-testkit's Python CI:
install deps → run pytest (or whatever git-testkit uses) → build the package.

### 5.5 Package
Mirror git-testkit's PyPI publishing setup.
Commit: `feat(git-harness): Python wrapper`

---

## Phase 6 — Java Wrapper

> Follow git-testkit's Java wrapper structure exactly. Read it before writing anything here.

### 6.1 Scaffold
Mirror whatever structure git-testkit uses under `testkit/java/`. Likely:
- Maven or Gradle project under `testkit/java/`
- `src/main/java/io/gitfire/harness/`
- `testkit/java/README.md`

### 6.2 Implement
Expose the same surface area as the Go module.
The Java API should feel idiomatic (camelCase, checked exceptions or Result types, Builder pattern where appropriate).

### 6.3 Tests
Mirror git-testkit's Java test setup (JUnit or TestNG, matching whatever git-testkit uses).

### 6.4 CI
Add a `java-wrapper.yml` workflow mirroring git-testkit's Java CI.

### 6.5 Package
Mirror git-testkit's Maven publishing setup.
Commit: `feat(git-harness): Java wrapper`

---

## Phase 7 — Release

### 7.1 Tag Go module
```bash
cd git-harness
git tag v0.1.0
git push origin v0.1.0
```

### 7.2 Remove replace directives
In git-fire/go.mod and git-fcuk/go.mod:
- Remove `replace github.com/git-fire/git-harness => ...`
- Run `go get github.com/git-fire/git-harness@v0.1.0` + `go mod tidy` in each

### 7.3 Final green check — all repos
```bash
# git-harness
go build ./... && go test -race ./... && go vet ./...

# git-fire
go build ./... && go test ./... && go vet ./...

# git-fcuk
go build ./... && go test ./... && go vet ./...
```

Commit consumers: `chore(git-harness): pin to published git-harness v0.1.0`

---

## Guardrails (non-negotiable)

1. **Read git-testkit first** — it is the canonical pattern for everything.
2. **Never delete from origin before the consumer compiles and tests pass.**
3. **Never add app-layer imports to git-harness** (no Cobra, Viper, Bubble Tea, Lipgloss).
4. **One concern per commit.** Audit → tests → copy → wire → delete → wrappers → release.
5. **`go vet` must pass at every commit.**
6. If a symbol's classification is unclear, mark it `BOUNDARY` in `audit.md` and stop for human review. Do not guess.
7. Python and Java wrappers must mirror git-testkit's approach — do not invent a new pattern.

---

## Deliverables checklist
- [ ] git-testkit wrapper pattern studied and documented
- [ ] `specs/001-git-harness-audit/audit.md` written
- [ ] Coverage ≥ 80% on all EXTRACT packages (both repos)
- [ ] `git-harness` Go module builds and tests pass
- [ ] `git-fire` updated, builds, tests pass
- [ ] `git-fcuk` updated, builds, tests pass
- [ ] Dead code deleted from both consumers
- [ ] Python wrapper implemented, tested, CI'd
- [ ] Java wrapper implemented, tested, CI'd
- [ ] `v0.1.0` tagged on git-harness
- [ ] `replace` directives removed from consumers
- [ ] All three Go repos green on final check
