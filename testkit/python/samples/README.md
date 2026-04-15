## Python sample smoke implementations

Runnable examples that exercise the bridge end-to-end. They exit non-zero on failure.

From the repository root (with `GIT_HARNESS_CLI` pointing at a built binary, recommended):

```bash
cd testkit/python
python -m pip install -e ".[dev]"
python -m samples.smoke_repo_flow
python -m samples.smoke_safety_flow
```

Or from repo root using `PYTHONPATH`:

```bash
PYTHONPATH=testkit/python python3 testkit/python/samples/smoke_repo_flow.py
```
