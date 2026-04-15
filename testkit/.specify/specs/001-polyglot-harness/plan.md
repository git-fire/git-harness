# Implementation plan: Polyglot git-harness

**Feature**: `001-polyglot-harness`  
**Input**: `testkit/.specify/specs/001-polyglot-harness/spec.md`  
**Status**: Implemented (canonical spec-kit baseline)

## Summary

Ship **`cmd/git-harness-cli`** (JSON stdin → JSON stdout) and thin clients under **`testkit/python`** (`git_harness`) and **`testkit/java`** (`io.gitfire.harness`), mirroring [git-testkit](https://github.com/git-fire/git-testkit).

## Artifact map

- Bridge: `cmd/git-harness-cli/main.go`
- Python: `testkit/python/git_harness/`, `testkit/python/tests/`, `testkit/python/samples/`
- Java: `testkit/java/` (Maven, Gson)
- Spec-kit: `testkit/.specify/**`
