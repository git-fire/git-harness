package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/git-fire/git-harness/safety"
)

// CommitOptions configures auto-commit behavior.
type CommitOptions struct {
	Message          string // Commit message
	AddAll           bool   // Run git add -A; must be set explicitly (default: false)
	UseDualBranch    bool   // Use staged/unstaged dual branch strategy (default: true)
	ReturnToOriginal bool   // Reset to original state after creating branches (default: true)
}

// AutoCommitResult contains information about branches created.
type AutoCommitResult struct {
	StagedBranch string // Empty if no staged changes
	FullBranch   string // Empty if no unstaged changes
	BothCreated  bool   // True if both branches were created
}

// Worktree represents a git worktree.
type Worktree struct {
	Path   string // Absolute path to worktree
	Branch string // Current branch in this worktree
	Head   string // Current HEAD SHA
	IsMain bool   // True if this is the main worktree
}

// AutoCommitDirty commits all uncommitted changes in a repo.
// Returns nil if repo is already clean.
func AutoCommitDirty(repoPath string, opts CommitOptions) error {
	isDirty, err := IsDirty(repoPath)
	if err != nil {
		return fmt.Errorf("failed to check repo status: %w", err)
	}
	if !isDirty {
		return nil
	}

	// Refuse to commit in detached HEAD — the commit would be unreachable.
	if _, err := GetCurrentBranch(repoPath); err != nil {
		return fmt.Errorf("cannot auto-commit: %w", err)
	}

	if opts.AddAll {
		cmd := exec.Command("git", "add", "-A")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return commandError("git add", err, output)
		}
	}

	message := opts.Message
	if message == "" {
		message = fmt.Sprintf("git-fire emergency backup - %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return commandError("git commit", err, output)
	}
	return nil
}

// IsDirty checks if a repo has uncommitted changes.
func IsDirty(repoPath string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return len(output) > 0, nil
}

// DetectConflict checks if local and remote branches have diverged.
// Returns: hasConflict, localSHA, remoteSHA, error.
func DetectConflict(repoPath, branch, remote string) (bool, string, string, error) {
	cmd := exec.Command("git", "fetch", remote)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return false, "", "", commandError("git fetch", err, output)
	}

	localSHA, err := getCommitSHA(repoPath, branch)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get local SHA: %w", err)
	}

	remoteBranch := fmt.Sprintf("%s/%s", remote, branch)
	remoteSHA, err := getCommitSHA(repoPath, remoteBranch)
	if err != nil {
		// Remote branch might not exist yet.
		return false, localSHA, "", nil
	}

	mergeBaseSHA, hasMergeBase, err := getMergeBaseSHA(repoPath, branch, remoteBranch)
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get merge-base: %w", err)
	}
	if !hasMergeBase {
		return true, localSHA, remoteSHA, nil
	}
	hasConflict := mergeBaseSHA != remoteSHA && mergeBaseSHA != localSHA
	return hasConflict, localSHA, remoteSHA, nil
}

// GetCommitSHA returns the SHA for a ref in the repository.
func GetCommitSHA(repoPath, ref string) (string, error) {
	return getCommitSHA(repoPath, ref)
}

func getCommitSHA(repoPath, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed for %s: %w", ref, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func getMergeBaseSHA(repoPath, leftRef, rightRef string) (string, bool, error) {
	cmd := exec.Command("git", "merge-base", leftRef, rightRef)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", false, nil
		}
		return "", false, commandError("git merge-base", err, output)
	}
	return strings.TrimSpace(string(output)), true, nil
}

// RefIsAncestor reports whether ancestorRef is an ancestor of descendantRef
// (git merge-base --is-ancestor).
func RefIsAncestor(repoPath, ancestorRef, descendantRef string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", ancestorRef, descendantRef)
	cmd.Dir = repoPath
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, fmt.Errorf("git merge-base --is-ancestor %s %s failed: %w", ancestorRef, descendantRef, err)
}

// FetchRemote runs git fetch for a single remote name.
func FetchRemote(repoPath, remote string) error {
	cmd := exec.Command("git", "fetch", remote)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return commandError("git fetch", err, output)
	}
	return nil
}

// CreateFireBranch creates a new fire backup branch.
func CreateFireBranch(repoPath, originalBranch, localSHA string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	shortSHA := localSHA
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}

	branchName := fmt.Sprintf("git-fire-backup-%s-%s-%s", originalBranch, timestamp, shortSHA)
	// Point the backup ref at localSHA — never at current HEAD. Callers may be on
	// a different branch while backing up a diverged or inactive local branch.
	cmd := exec.Command("git", "branch", branchName, localSHA)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", commandError("git branch "+branchName, err, output)
	}
	return branchName, nil
}

// PushBranch pushes a specific branch to a remote.
func PushBranch(repoPath, remote, branch string) error {
	cmd := exec.Command("git", "push", remote, branch)
	cmd.Dir = repoPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return commandError("git push", err, stderr.Bytes())
	}
	return nil
}

// PushAllBranches pushes all branches to a remote.
func PushAllBranches(repoPath, remote string) error {
	cmd := exec.Command("git", "push", remote, "--all")
	cmd.Dir = repoPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return commandError("git push --all", err, stderr.Bytes())
	}
	return nil
}

// ListRemoteBranches returns short branch names under remote/ (excluding HEAD).
func ListRemoteBranches(repoPath, remote string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git branch -r failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	branches := make([]string, 0, len(lines))
	prefix := remote + "/"
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, prefix) {
			branch := strings.TrimPrefix(line, prefix)
			if branch != "HEAD" && !strings.Contains(branch, "->") {
				branches = append(branches, branch)
			}
		}
	}
	return branches, nil
}

// ListLocalBranches returns local branch short names.
func ListLocalBranches(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git branch failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// GetCurrentBranch returns the currently checked out branch.
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current failed: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "", fmt.Errorf("not on any branch (detached HEAD?)")
	}
	return branch, nil
}

// HasStagedChanges checks if there are staged changes.
func HasStagedChanges(repoPath string) (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = repoPath
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return true, nil
		}
		return false, fmt.Errorf("git diff --cached --quiet failed: %w", err)
	}
	return false, nil
}

// HasUnstagedChanges checks if there are unstaged changes (including untracked files).
func HasUnstagedChanges(repoPath string) (bool, error) {
	cmd := exec.Command("git", "diff", "--quiet")
	cmd.Dir = repoPath
	err := cmd.Run()
	hasModified := false
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			hasModified = true
		} else {
			return false, fmt.Errorf("git diff --quiet failed: %w", err)
		}
	}

	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git ls-files failed: %w", err)
	}

	hasUntracked := len(strings.TrimSpace(string(output))) > 0
	return hasModified || hasUntracked, nil
}

// ListWorktrees returns all worktrees for a repository.
func ListWorktrees(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var worktrees []Worktree
	var current Worktree
	isFirst := true

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" {
				current.IsMain = isFirst
				worktrees = append(worktrees, current)
				current = Worktree{}
				isFirst = false
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		key, value := parts[0], parts[1]
		switch key {
		case "worktree":
			current.Path = value
		case "HEAD":
			current.Head = value
		case "branch":
			current.Branch = strings.TrimPrefix(value, "refs/heads/")
		}
	}

	if current.Path != "" {
		current.IsMain = isFirst
		worktrees = append(worktrees, current)
	}
	return worktrees, nil
}

// AutoCommitDirtyWithStrategy commits changes using the staged/unstaged dual branch strategy.
func AutoCommitDirtyWithStrategy(repoPath string, opts CommitOptions) (result *AutoCommitResult, retErr error) {
	result = &AutoCommitResult{}

	if !opts.UseDualBranch {
		opts.UseDualBranch = true
	}
	if !opts.ReturnToOriginal {
		opts.ReturnToOriginal = true
	}

	currentBranch, err := GetCurrentBranch(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	originalHeadSHA, hasOriginalHead, err := getOptionalHeadSHA(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get original HEAD SHA: %w", err)
	}

	hasStaged, err := HasStagedChanges(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check staged changes: %w", err)
	}

	hasUnstaged, err := HasUnstagedChanges(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check unstaged changes: %w", err)
	}

	if !hasStaged && !hasUnstaged {
		return result, nil
	}

	originalIndexTree, err := captureIndexTree(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to capture original index state: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	commitsCreated := 0
	successRestoreAttempted := false

	defer func() {
		if retErr == nil || commitsCreated == 0 || successRestoreAttempted {
			return
		}
		if cleanupErr := restoreOriginalState(repoPath, hasOriginalHead, originalHeadSHA, originalIndexTree); cleanupErr != nil {
			retErr = fmt.Errorf("%w; failed cleanup restore of original staged/unstaged state: %v", retErr, cleanupErr)
		}
	}()

	if hasStaged && !hasUnstaged {
		message := opts.Message
		if message == "" {
			message = fmt.Sprintf("git-fire staged backup - %s", time.Now().Format("2006-01-02 15:04:05"))
		}

		if err := commitChanges(repoPath, message, false); err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to commit staged changes: %w", err))
		}
		commitsCreated++

		sha, err := getCommitSHA(repoPath, "HEAD")
		if err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to get commit SHA: %w", err))
		}
		shortSHA := sha
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}

		branchName := fmt.Sprintf("git-fire-staged-%s-%s-%s", currentBranch, timestamp, shortSHA)
		if err := createBranch(repoPath, branchName); err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to create staged branch: %w", err))
		}
		result.StagedBranch = branchName
	}

	if !hasStaged && hasUnstaged {
		message := opts.Message
		if message == "" {
			message = fmt.Sprintf("git-fire full backup - %s", time.Now().Format("2006-01-02 15:04:05"))
		}

		if err := commitChanges(repoPath, message, true); err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to commit unstaged changes: %w", err))
		}
		commitsCreated++

		sha, err := getCommitSHA(repoPath, "HEAD")
		if err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to get commit SHA: %w", err))
		}
		shortSHA := sha
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}

		branchName := fmt.Sprintf("git-fire-full-%s-%s-%s", currentBranch, timestamp, shortSHA)
		if err := createBranch(repoPath, branchName); err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to create full branch: %w", err))
		}
		result.FullBranch = branchName
	}

	if hasStaged && hasUnstaged {
		message1 := opts.Message
		if message1 == "" {
			message1 = fmt.Sprintf("git-fire staged backup - %s", time.Now().Format("2006-01-02 15:04:05"))
		}
		if err := commitChanges(repoPath, message1, false); err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to commit staged changes: %w", err))
		}
		commitsCreated++

		sha1, err := getCommitSHA(repoPath, "HEAD")
		if err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to get staged commit SHA: %w", err))
		}
		shortSHA1 := sha1
		if len(shortSHA1) > 7 {
			shortSHA1 = shortSHA1[:7]
		}
		stagedBranchName := fmt.Sprintf("git-fire-staged-%s-%s-%s", currentBranch, timestamp, shortSHA1)
		if err := createBranch(repoPath, stagedBranchName); err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to create staged branch: %w", err))
		}
		result.StagedBranch = stagedBranchName

		message2 := opts.Message
		if message2 == "" {
			message2 = fmt.Sprintf("git-fire full backup - %s", time.Now().Format("2006-01-02 15:04:05"))
		}
		if err := commitChanges(repoPath, message2, true); err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to commit unstaged changes: %w", err))
		}
		commitsCreated++

		sha2, err := getCommitSHA(repoPath, "HEAD")
		if err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to get full commit SHA: %w", err))
		}
		shortSHA2 := sha2
		if len(shortSHA2) > 7 {
			shortSHA2 = shortSHA2[:7]
		}
		fullBranchName := fmt.Sprintf("git-fire-full-%s-%s-%s", currentBranch, timestamp, shortSHA2)
		if err := createBranch(repoPath, fullBranchName); err != nil {
			return returnResultOnError(result, fmt.Errorf("failed to create full branch: %w", err))
		}
		result.FullBranch = fullBranchName
		result.BothCreated = true
	}

	if opts.ReturnToOriginal && commitsCreated > 0 {
		successRestoreAttempted = true
		if err := restoreOriginalState(repoPath, hasOriginalHead, originalHeadSHA, originalIndexTree); err != nil {
			return returnResultOnError(result, err)
		}
	}

	return result, nil
}

func resetMixedToCommit(repoPath, sha string) error {
	cmd := exec.Command("git", "reset", "--mixed", sha)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return commandError("git reset --mixed", err, output)
	}
	return nil
}

func resetToUnborn(repoPath string) error {
	branch, err := GetCurrentBranch(repoPath)
	if err != nil {
		return fmt.Errorf("resolve current branch for unborn reset: %w", err)
	}
	ref := fmt.Sprintf("refs/heads/%s", branch)
	cmd := exec.Command("git", "update-ref", "-d", ref)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return commandError("git update-ref -d "+ref, err, output)
	}
	return nil
}

func clearIndexForUnborn(repoPath string) error {
	cmd := exec.Command("git", "read-tree", "--empty")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return commandError("git read-tree --empty", err, output)
	}
	return nil
}

func captureIndexTree(repoPath string) (string, error) {
	cmd := exec.Command("git", "write-tree")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", commandError("git write-tree", err, output)
	}
	tree := strings.TrimSpace(string(output))
	if tree == "" {
		return "", fmt.Errorf("git write-tree returned empty tree SHA")
	}
	return tree, nil
}

func restoreIndexTree(repoPath, treeSHA string) error {
	cmd := exec.Command("git", "read-tree", treeSHA)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return commandError("git read-tree "+treeSHA, err, output)
	}
	return nil
}

func restoreOriginalState(repoPath string, hasOriginalHead bool, originalHeadSHA, originalIndexTree string) error {
	if hasOriginalHead {
		if err := resetMixedToCommit(repoPath, originalHeadSHA); err != nil {
			return err
		}
	} else {
		if err := resetToUnborn(repoPath); err != nil {
			return err
		}
		if err := clearIndexForUnborn(repoPath); err != nil {
			return err
		}
	}
	return restoreIndexTree(repoPath, originalIndexTree)
}

func returnResultOnError(result *AutoCommitResult, err error) (*AutoCommitResult, error) {
	if result != nil && (result.StagedBranch != "" || result.FullBranch != "" || result.BothCreated) {
		return result, err
	}
	return nil, err
}

func getOptionalHeadSHA(repoPath string) (string, bool, error) {
	cmd := exec.Command("git", "rev-parse", "-q", "--verify", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", false, nil
		}
		return "", false, commandError("git rev-parse -q --verify HEAD", err, output)
	}

	sha := strings.TrimSpace(string(output))
	if sha == "" {
		return "", false, fmt.Errorf("git rev-parse -q --verify HEAD returned empty output")
	}
	return sha, true, nil
}

func commitChanges(repoPath, message string, addAll bool) error {
	if addAll {
		cmd := exec.Command("git", "add", "-A")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return commandError("git add", err, output)
		}
	}

	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return commandError("git commit", err, output)
	}
	return nil
}

// GetUncommittedFiles returns the relative paths of all uncommitted files.
func GetUncommittedFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "-z")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	var files []string
	entries := strings.Split(string(output), "\x00")
	i := 0
	for i < len(entries) {
		entry := entries[i]
		i++
		if len(entry) < 4 {
			continue
		}
		xy := entry[:2]
		path := entry[3:]
		if (xy[0] == 'R' || xy[0] == 'C') && i < len(entries) {
			i++ // consume old path token
		}
		if path != "" {
			files = append(files, path)
		}
	}
	return files, nil
}

func createBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "branch", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return commandError("git branch "+branchName, err, output)
	}
	return nil
}

func commandError(action string, err error, output []byte) error {
	out := strings.TrimSpace(string(output))
	if out == "" {
		return fmt.Errorf("%s failed: %w", action, err)
	}
	return fmt.Errorf("%s failed: %w (stderr: %s)", action, err, safety.SanitizeText(out))
}
