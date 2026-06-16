package git

import (
	"os/exec"
	"strings"
	"testing"

	testutil "github.com/git-fire/git-testkit"
)

func TestNonInteractiveGitEnv_OverridesExistingPrompt(t *testing.T) {
	env := nonInteractiveGitEnv([]string{
		"HOME=/tmp",
		"GIT_TERMINAL_PROMPT=1",
		"PATH=/bin",
	})
	if !containsEnv(env, "GIT_TERMINAL_PROMPT=0") {
		t.Fatalf("expected GIT_TERMINAL_PROMPT=0 in env, got %#v", env)
	}
	for _, e := range env {
		if e == "GIT_TERMINAL_PROMPT=1" {
			t.Fatalf("did not override existing GIT_TERMINAL_PROMPT: %#v", env)
		}
	}
}

func TestPrepareNetworkGit_SetsEnvOnCommand(t *testing.T) {
	cmd := exec.Command("git", "version")
	PrepareNetworkGit(cmd)
	if !containsEnv(cmd.Env, "GIT_TERMINAL_PROMPT=0") {
		t.Fatalf("prepareNetworkGit did not set GIT_TERMINAL_PROMPT=0: %#v", cmd.Env)
	}
}

func TestFetchRemote_UnauthenticatedHTTPSFailsWithoutPrompt(t *testing.T) {
	repo := testutil.CreateTestRepo(t, testutil.RepoOptions{
		Name: "https-fetch-repo",
		Remotes: map[string]string{
			"origin": "https://github.com/git-fire/nonexistent-repo-auth-test.git",
		},
	})

	err := FetchRemote(repo, "origin")
	if err == nil {
		t.Fatal("expected fetch to fail without credentials")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "terminal prompts disabled") &&
		!strings.Contains(msg, "could not read username") {
		t.Fatalf("expected non-interactive auth failure, got: %v", err)
	}
}

func containsEnv(env []string, want string) bool {
	for _, e := range env {
		if e == want {
			return true
		}
	}
	return false
}
