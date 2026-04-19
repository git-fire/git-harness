## Polyglot testkit (git-harness)

This tree mirrors [git-testkit](https://github.com/git-fire/git-testkit): **Go core** at the repository root, **`cmd/git-harness-cli`** as the JSON stdin/stdout bridge, and thin **Python** / **Java** clients under `testkit/python` and `testkit/java`.

Spec-kit style metadata lives under `testkit/.specify/` (validated in CI).

### Run conformance locally

- Python: `cd testkit/python && python3 -m pip install -e ".[dev]" && python3 -m pytest tests/ -v`
- Java: `cd testkit/java && mvn test`
- Go: from repository root, `go test ./...`

Set `GIT_HARNESS_CLI` to a prebuilt `./bin/git-harness-cli` when you want to avoid `go run` during wrapper tests.
