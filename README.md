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

## Rewrite orchestration

The `git` package now includes bounded multi-pass rewrite orchestration via
`RunRewriteScenario`, designed for flows that must:

1. Detect whether intervention is needed.
2. Intervene (run one rewrite pass).
3. Verify the repo is clean.
4. Rerun from detect until clean or attempts are exhausted.

Use it for safety-first rewrite loops where you need deterministic stop
conditions and pass-by-pass telemetry (`RewriteScenarioResult.Passes`).

### Validate locally

```bash
# Targeted rewrite orchestration tests
go test -count=1 ./git -run RunRewriteScenario

# Full module sanity
go test -count=1 ./...
```

Expected outcome:

- Targeted run: `ok   github.com/git-fire/git-harness/git ...`
- Full run: all packages return `ok` with no failing tests
- If a rewrite loop never reaches clean within `MaxAttempts`, callers receive
  `ErrRewriteAttemptsExceeded`

## Stability

`v0.x` releases may change APIs; pin a minor or patch version in consumers.

## License

MIT — see [LICENSE](LICENSE).
