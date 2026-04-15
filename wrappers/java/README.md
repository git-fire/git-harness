# git-harness (Java)

JSON-over-subprocess client for `git-harness-cli`, using Gson to parse responses.

Set `GIT_HARNESS_CLI` to a prebuilt binary, or use `go run ./cmd/git-harness-cli` from the repository root (default when unset).

## Build and test

```bash
cd wrappers/java
mvn test
```

Sample smoke tests (also run in CI):

```bash
mvn -Dtest=SampleRepoFlowSmoke,SampleSafetyFlowSmoke test
```
