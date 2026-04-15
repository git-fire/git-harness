#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SPEC_DIR="$ROOT_DIR/testkit/.specify/specs/001-polyglot-harness"

required_files=(
  "$ROOT_DIR/testkit/.specify/memory/constitution.md"
  "$SPEC_DIR/spec.md"
  "$SPEC_DIR/plan.md"
  "$SPEC_DIR/tasks.md"
  "$SPEC_DIR/contracts/cli-protocol.json"
  "$SPEC_DIR/checklists/quality.md"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "missing required spec-kit artifact: $file" >&2
    exit 1
  fi
done

grep -q "Status\\*\\*: Implemented (canonical spec-kit baseline)" "$SPEC_DIR/spec.md"
grep -q "Status\\*\\*: Implemented (canonical spec-kit baseline)" "$SPEC_DIR/plan.md"
grep -q "T015 Add spec-kit command workflow doc + shell helper" "$SPEC_DIR/tasks.md"
grep -q "\\[x\\] T015" "$SPEC_DIR/tasks.md"
grep -q "\"supported_ops\"" "$SPEC_DIR/contracts/cli-protocol.json"
grep -q "\\[x\\] Smoke test coverage exists for Go, Python wrapper, and Java wrapper paths." "$SPEC_DIR/checklists/quality.md"

echo "spec-kit artifacts validated"
