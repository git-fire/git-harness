# Spec 002 — Pre-Extraction Test Coverage

**Branch:** `spec/002-git-harness-test-coverage`
**Depends on:** Spec 001 audit.md complete
**Status:** Ready for implementation after 001

## Goal
Bring all `EXTRACT`-classified packages to ≥ 80% test coverage before any code is moved.
Tests are the contract that proves extraction didn't break anything.

## Scope
Packages identified as `EXTRACT` in `specs/001-git-harness-audit/audit.md`.

## Acceptance Criteria
- `go test -coverprofile=coverage.out ./internal/<pkg>/...` reports ≥ 80% for each EXTRACT package
- Every exported function has at minimum: one happy-path test, one error-path test
- Tests use `git-testkit` for any tests that require a real git repo on disk
- Boundary integration tests exist for every `BOUNDARY` symbol (call from CLI layer → primitive)
- `go test ./...` is green in both repos at the end of this spec
- No implementation code is changed — tests only

## Test naming conventions (match existing git-fire style)
```go
func TestRunnerExec_Success(t *testing.T) { ... }
func TestRunnerExec_NonZeroExit(t *testing.T) { ... }
func TestSanitize_RemovesAnsiCodes(t *testing.T) { ... }
```

## Commit cadence
One commit per package:
```
test(git-harness-prep): coverage for internal/git
test(git-harness-prep): coverage for internal/safety
test(git-harness-prep): boundary tests for executor
```
