# git-harness (Python)

Thin JSON-over-subprocess client for the `git-harness-cli` tool in this repository.

Set `GIT_HARNESS_CLI` to a prebuilt binary path (recommended in CI), or rely on `go run ./cmd/git-harness-cli` from the repository root.

## Development

```bash
cd wrappers/python
python -m pip install -e ".[dev]"
python -m pytest tests/ -v
```

Samples: see `samples/README.md`.
