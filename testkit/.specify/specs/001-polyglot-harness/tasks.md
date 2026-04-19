# Tasks: Polyglot git-harness CLI bridge

## Phase 1 — Bridge

- [x] T001 Add `cmd/git-harness-cli` JSON protocol for `git` and `safety` operations

## Phase 2 — Wrappers

- [x] T002 Add Python client under `testkit/python/git_harness/`
- [x] T003 Add Java client under `testkit/java/` (Maven + Gson)

## Phase 3 — Smoke and CI

- [x] T004 Add Python pytest coverage and sample modules under `testkit/python/samples/`
- [x] T005 Add Java JUnit tests and `Sample*Smoke` flows
- [x] T006 Wire `.github/workflows/ci.yml` (Go + testkit jobs + cross-platform matrix)

## Phase 4 — Spec-kit alignment

- [x] T011 Add `testkit/.specify/memory/constitution.md`
- [x] T012 Add spec in `testkit/.specify/specs/001-polyglot-harness/spec.md`
- [x] T013 Add plan in `testkit/.specify/specs/001-polyglot-harness/plan.md`
- [x] T014 Add tasks ledger (this file)
- [x] T015 Add spec-kit command workflow doc + shell helper
- [x] T016 Add CLI contract JSON and quality checklist
