# git-harness

Go library extracted from [git-fire](https://github.com/git-fire/git-fire): subprocess-oriented git helpers and small safety utilities.

**Module:** `github.com/git-fire/git-harness`

## Packages

- **`git`** — repository scanning, status, commits, pushes, worktrees, and related helpers.
- **`safety`** — redaction and secret-pattern scanning helpers used by git error paths.

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
