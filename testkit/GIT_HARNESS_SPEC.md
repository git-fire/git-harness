# git-harness polyglot spec (summary)

Polyglot layout matches **git-testkit**: repository root Go module, **`cmd/git-harness-cli`** JSON bridge, and language clients under **`testkit/`**.

## Layout

| Path | Role |
|------|------|
| `cmd/git-harness-cli/` | stdin JSON → stdout JSON |
| `testkit/python/` | `git_harness` package (subprocess client) |
| `testkit/java/` | `io.gitfire.harness.CliBridge` (Maven) |
| `testkit/.specify/` | Spec-kit artifacts + `validate_specify.sh` |

## Environment

- **`GIT_HARNESS_CLI`**: path to the built `git-harness-cli` binary (recommended in CI). If unset, clients use `go run ./cmd/git-harness-cli` from the repository root.

## Canonical contract

Machine-readable op list: `testkit/.specify/specs/001-polyglot-harness/contracts/cli-protocol.json`.
