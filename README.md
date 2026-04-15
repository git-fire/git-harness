# git-harness

Go library extracted from [git-fire](https://github.com/git-fire/git-fire): subprocess-oriented git helpers and small safety utilities.

**Module:** `github.com/git-fire/git-harness`

## Packages

- **`git`** — repository scanning, status, commits, pushes, worktrees, and related helpers.
- **`safety`** — redaction and secret-pattern scanning helpers used by git error paths.

## Polyglot wrappers

Python and Java clients use the same layout as [git-testkit](https://github.com/git-fire/git-testkit) under **`testkit/`**: build `cmd/git-harness-cli`, set **`GIT_HARNESS_CLI`** to that binary (or rely on `go run ./cmd/git-harness-cli` from the repo root). Code lives in `testkit/python` and `testkit/java`; runnable samples are `testkit/python/samples/` and the Java `Sample*Smoke` tests.

## Requirements

- Go **1.24**+ (see `go.mod`).
- **`git`** on `PATH` for tests and runtime (package shells out to the git binary).

## Development

```bash
go build ./...
go vet ./...
go test -race -count=1 ./...
```

## Stability

`v0.x` releases may change APIs; pin a minor or patch version in consumers.

## License

MIT — see [LICENSE](LICENSE).
