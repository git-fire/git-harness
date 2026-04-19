# Feature specification: Polyglot git-harness (CLI bridge)

**Feature branch**: `001-polyglot-harness`  
**Status**: Implemented (canonical spec-kit baseline)

## Goal

Expose **git-harness** (`git` + `safety` packages) to Python and Java via **`cmd/git-harness-cli`**, using the same repository layout pattern as **git-testkit** (`testkit/python`, `testkit/java`, `testkit/.specify`).

## User scenarios

1. **Python / Java test authors** invoke scan, git metadata, and safety helpers without reimplementing subprocess orchestration.
2. **CI** validates spec-kit artifacts, builds the CLI once, and runs wrapper tests plus sample smoke flows on Linux and a cross-platform matrix.

## Acceptance

- `testkit/.specify/scripts/validate_specify.sh` passes.
- `go test ./...`, `pytest` under `testkit/python`, and `mvn test` under `testkit/java` pass when `GIT_HARNESS_CLI` points at a built binary.
