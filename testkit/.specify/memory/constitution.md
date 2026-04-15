# git-harness testkit constitution

## Core principles

### I. Real git only

Conformance tests run against the real `git` binary on `PATH`. Wrappers delegate to **`git-harness-cli`**, which shells out to git for repository operations.

### II. Single behavior source

The **Go** packages (`git`, `safety`) and **`cmd/git-harness-cli`** are the behavior source. Python and Java clients are thin bridges over the JSON protocol.

### III. Deterministic and bounded

Samples and tests use temporary directories under the OS temp root. No network remotes beyond local bare repos created on disk.

### IV. Executable proof

Every polyglot change keeps **Go tests**, **Python pytest**, and **Java Maven** smoke paths green in CI.
