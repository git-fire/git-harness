package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/git-fire/git-harness/git"
	"github.com/git-fire/git-harness/safety"
)

type request struct {
	Op string `json:"op"`

	// scan_repositories
	ScanOptions *scanOptionsInput `json:"scanOptions,omitempty"`

	// analyze_repository, git_* ops
	RepoPath string `json:"repoPath,omitempty"`

	// git_get_commit_sha, git_ref_is_ancestor
	Ref string `json:"ref,omitempty"`
	// git_ref_is_ancestor
	AncestorRef      string `json:"ancestorRef,omitempty"`
	DescendantRef    string `json:"descendantRef,omitempty"`
	Branch           string `json:"branch,omitempty"`
	Remote           string `json:"remote,omitempty"`
	OriginalBranch   string `json:"originalBranch,omitempty"`
	LocalSHA         string `json:"localSHA,omitempty"`
	Message          string `json:"message,omitempty"`
	AddAll           *bool  `json:"addAll,omitempty"`
	UseDualBranch    *bool  `json:"useDualBranch,omitempty"`
	ReturnToOriginal *bool  `json:"returnToOriginal,omitempty"`

	// safety
	Text            string                `json:"text,omitempty"`
	Files           []string              `json:"files,omitempty"`
	FilesSuspicious []suspiciousFileInput `json:"suspiciousFiles,omitempty"`
}

type scanOptionsInput struct {
	RootPath    string          `json:"rootPath,omitempty"`
	Exclude     []string        `json:"exclude,omitempty"`
	MaxDepth    int             `json:"maxDepth,omitempty"`
	UseCache    *bool           `json:"useCache,omitempty"`
	CacheFile   string          `json:"cacheFile,omitempty"`
	CacheTTL    string          `json:"cacheTTL,omitempty"`
	Workers     int             `json:"workers,omitempty"`
	KnownPaths  map[string]bool `json:"knownPaths,omitempty"`
	DisableScan *bool           `json:"disableScan,omitempty"`
}

type suspiciousFileInput struct {
	Path        string   `json:"path"`
	Reason      string   `json:"reason"`
	Patterns    []string `json:"patterns,omitempty"`
	LineNumbers []int    `json:"lineNumbers,omitempty"`
}

type response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`

	Repositories []repositoryOut `json:"repositories,omitempty"`
	Repository   *repositoryOut  `json:"repository,omitempty"`

	Dirty           *bool                  `json:"dirty,omitempty"`
	Output          *string                `json:"output,omitempty"`
	SHA             string                 `json:"sha,omitempty"`
	Branches        []string               `json:"branches,omitempty"`
	HasConflict     *bool                  `json:"hasConflict,omitempty"`
	LocalSHA        string                 `json:"localSHA,omitempty"`
	RemoteSHA       string                 `json:"remoteSHA,omitempty"`
	IsAncestor      *bool                  `json:"isAncestor,omitempty"`
	Branch          string                 `json:"branch,omitempty"`
	Staged          *bool                  `json:"staged,omitempty"`
	Unstaged        *bool                  `json:"unstaged,omitempty"`
	Paths           []string               `json:"paths,omitempty"`
	Worktrees       []worktreeOut          `json:"worktrees,omitempty"`
	FireBranch      string                 `json:"fireBranch,omitempty"`
	StagedBranch    string                 `json:"stagedBranch,omitempty"`
	FullBranch      string                 `json:"fullBranch,omitempty"`
	BothCreated     *bool                  `json:"bothCreated,omitempty"`
	Text            string                 `json:"text,omitempty"`
	Lines           []string               `json:"lines,omitempty"`
	Warning         string                 `json:"warning,omitempty"`
	Notice          string                 `json:"notice,omitempty"`
	SuspiciousFiles []suspiciousFileOutput `json:"suspiciousFiles,omitempty"`
}

type suspiciousFileOutput struct {
	Path        string   `json:"path"`
	Reason      string   `json:"reason"`
	Patterns    []string `json:"patterns"`
	LineNumbers []int    `json:"lineNumbers"`
}

type remoteOut struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type repositoryOut struct {
	Path         string      `json:"path"`
	Name         string      `json:"name"`
	Remotes      []remoteOut `json:"remotes"`
	Branches     []string    `json:"branches"`
	IsDirty      bool        `json:"isDirty"`
	LastModified time.Time   `json:"lastModified"`
	Selected     bool        `json:"selected"`
	Mode         string      `json:"mode"`
}

type worktreeOut struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
	Head   string `json:"head"`
	IsMain bool   `json:"isMain"`
}

func main() {
	req, err := parseRequest()
	if err != nil {
		writeResponse(response{OK: false, Error: err.Error()})
		os.Exit(1)
	}

	res, err := handle(req)
	if err != nil {
		writeResponse(response{OK: false, Error: err.Error()})
		os.Exit(1)
	}
	if err := writeResponse(res); err != nil {
		os.Exit(1)
	}
}

func parseRequest() (request, error) {
	var req request
	dec := json.NewDecoder(os.Stdin)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return request{}, fmt.Errorf("invalid JSON request: %w", err)
	}
	if strings.TrimSpace(req.Op) == "" {
		return request{}, fmt.Errorf("missing required field: op")
	}
	return req, nil
}

func handle(req request) (response, error) {
	switch req.Op {
	case "scan_repositories":
		opts, err := mergeScanOptions(req.ScanOptions)
		if err != nil {
			return response{}, err
		}
		repos, err := git.ScanRepositories(opts)
		if err != nil {
			return response{}, err
		}
		out := make([]repositoryOut, 0, len(repos))
		for _, r := range repos {
			out = append(out, repoToOut(r))
		}
		return response{OK: true, Repositories: out}, nil

	case "analyze_repository":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		r, err := git.AnalyzeRepository(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		ro := repoToOut(r)
		return response{OK: true, Repository: &ro}, nil

	case "git_is_dirty":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		d, err := git.IsDirty(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Dirty: &d}, nil

	case "git_get_current_branch":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		b, err := git.GetCurrentBranch(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Branch: b}, nil

	case "git_get_commit_sha":
		if req.RepoPath == "" || req.Ref == "" {
			return response{}, fmt.Errorf("missing repoPath or ref")
		}
		sha, err := git.GetCommitSHA(req.RepoPath, req.Ref)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, SHA: sha}, nil

	case "git_list_local_branches":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		br, err := git.ListLocalBranches(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Branches: br}, nil

	case "git_list_remote_branches":
		if req.RepoPath == "" || req.Remote == "" {
			return response{}, fmt.Errorf("missing repoPath or remote")
		}
		br, err := git.ListRemoteBranches(req.RepoPath, req.Remote)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Branches: br}, nil

	case "git_ref_is_ancestor":
		if req.RepoPath == "" || req.AncestorRef == "" || req.DescendantRef == "" {
			return response{}, fmt.Errorf("missing repoPath, ancestorRef, or descendantRef")
		}
		ok, err := git.RefIsAncestor(req.RepoPath, req.AncestorRef, req.DescendantRef)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, IsAncestor: &ok}, nil

	case "git_detect_conflict":
		if req.RepoPath == "" || req.Branch == "" || req.Remote == "" {
			return response{}, fmt.Errorf("missing repoPath, branch, or remote")
		}
		has, local, remote, err := git.DetectConflict(req.RepoPath, req.Branch, req.Remote)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, HasConflict: &has, LocalSHA: local, RemoteSHA: remote}, nil

	case "git_has_staged_changes":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		v, err := git.HasStagedChanges(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Staged: &v}, nil

	case "git_has_unstaged_changes":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		v, err := git.HasUnstagedChanges(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Unstaged: &v}, nil

	case "git_get_uncommitted_files":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		paths, err := git.GetUncommittedFiles(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, Paths: paths}, nil

	case "git_list_worktrees":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		wts, err := git.ListWorktrees(req.RepoPath)
		if err != nil {
			return response{}, err
		}
		out := make([]worktreeOut, 0, len(wts))
		for _, w := range wts {
			out = append(out, worktreeOut{
				Path:   w.Path,
				Branch: w.Branch,
				Head:   w.Head,
				IsMain: w.IsMain,
			})
		}
		return response{OK: true, Worktrees: out}, nil

	case "git_auto_commit_dirty":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		if req.UseDualBranch != nil || req.ReturnToOriginal != nil {
			return response{}, fmt.Errorf("git_auto_commit_dirty does not support useDualBranch or returnToOriginal; use git_auto_commit_dirty_with_strategy")
		}
		co := git.CommitOptions{Message: req.Message}
		if req.AddAll != nil {
			co.AddAll = *req.AddAll
		}
		if err := git.AutoCommitDirty(req.RepoPath, co); err != nil {
			return response{}, err
		}
		return response{OK: true}, nil

	case "git_auto_commit_dirty_with_strategy":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		co := git.CommitOptions{Message: req.Message}
		if req.AddAll != nil {
			co.AddAll = *req.AddAll
		}
		if req.UseDualBranch != nil {
			co.UseDualBranch = *req.UseDualBranch
		}
		if req.ReturnToOriginal != nil {
			co.ReturnToOriginal = *req.ReturnToOriginal
		}
		res, err := git.AutoCommitDirtyWithStrategy(req.RepoPath, co)
		if err != nil {
			return response{}, err
		}
		bc := res.BothCreated
		return response{
			OK:           true,
			StagedBranch: res.StagedBranch,
			FullBranch:   res.FullBranch,
			BothCreated:  &bc,
		}, nil

	case "git_create_fire_branch":
		if req.RepoPath == "" || req.OriginalBranch == "" || req.LocalSHA == "" {
			return response{}, fmt.Errorf("missing repoPath, originalBranch, or localSHA")
		}
		name, err := git.CreateFireBranch(req.RepoPath, req.OriginalBranch, req.LocalSHA)
		if err != nil {
			return response{}, err
		}
		return response{OK: true, FireBranch: name}, nil

	case "git_fetch_remote":
		if req.RepoPath == "" || req.Remote == "" {
			return response{}, fmt.Errorf("missing repoPath or remote")
		}
		if err := git.FetchRemote(req.RepoPath, req.Remote); err != nil {
			return response{}, err
		}
		return response{OK: true}, nil

	case "git_push_branch":
		if req.RepoPath == "" || req.Remote == "" || req.Branch == "" {
			return response{}, fmt.Errorf("missing repoPath, remote, or branch")
		}
		if err := git.PushBranch(req.RepoPath, req.Remote, req.Branch); err != nil {
			return response{}, err
		}
		return response{OK: true}, nil

	case "git_push_all_branches":
		if req.RepoPath == "" || req.Remote == "" {
			return response{}, fmt.Errorf("missing repoPath or remote")
		}
		if err := git.PushAllBranches(req.RepoPath, req.Remote); err != nil {
			return response{}, err
		}
		return response{OK: true}, nil

	case "safety_sanitize_text":
		return response{OK: true, Text: safety.SanitizeText(req.Text)}, nil

	case "safety_recommended_gitignore_patterns":
		p := safety.RecommendedGitignorePatterns()
		return response{OK: true, Lines: p}, nil

	case "safety_security_notice":
		return response{OK: true, Notice: safety.SecurityNotice()}, nil

	case "safety_format_warning":
		files := make([]safety.SuspiciousFile, 0, len(req.FilesSuspicious))
		for _, f := range req.FilesSuspicious {
			files = append(files, safety.SuspiciousFile{
				Path:        f.Path,
				Reason:      f.Reason,
				Patterns:    f.Patterns,
				LineNumbers: f.LineNumbers,
			})
		}
		return response{OK: true, Warning: safety.FormatWarning(files)}, nil

	case "safety_scan_files":
		if req.RepoPath == "" {
			return response{}, fmt.Errorf("missing repoPath")
		}
		sc := safety.NewSecretScanner()
		found, err := sc.ScanFiles(req.RepoPath, req.Files)
		if err != nil {
			return response{}, err
		}
		out := make([]suspiciousFileOutput, 0, len(found))
		for _, f := range found {
			patterns := f.Patterns
			if patterns == nil {
				patterns = []string{}
			}
			lines := f.LineNumbers
			if lines == nil {
				lines = []int{}
			}
			out = append(out, suspiciousFileOutput{
				Path:        f.Path,
				Reason:      f.Reason,
				Patterns:    patterns,
				LineNumbers: lines,
			})
		}
		return response{OK: true, SuspiciousFiles: out}, nil

	default:
		return response{}, fmt.Errorf("unsupported op: %s", req.Op)
	}
}

func mergeScanOptions(in *scanOptionsInput) (git.ScanOptions, error) {
	opts := git.DefaultScanOptions()
	if in == nil {
		return opts, nil
	}
	if in.RootPath != "" {
		opts.RootPath = in.RootPath
	}
	if in.Exclude != nil {
		opts.Exclude = in.Exclude
	}
	if in.MaxDepth > 0 {
		opts.MaxDepth = in.MaxDepth
	}
	if in.UseCache != nil {
		opts.UseCache = *in.UseCache
	}
	if in.CacheFile != "" {
		opts.CacheFile = in.CacheFile
	}
	if in.CacheTTL != "" {
		d, err := time.ParseDuration(in.CacheTTL)
		if err != nil {
			return git.ScanOptions{}, fmt.Errorf("invalid scanOptions.cacheTTL %q: %w", in.CacheTTL, err)
		}
		opts.CacheTTL = d
	}
	if in.Workers > 0 {
		opts.Workers = in.Workers
	}
	if in.KnownPaths != nil {
		opts.KnownPaths = in.KnownPaths
	}
	if in.DisableScan != nil {
		opts.DisableScan = *in.DisableScan
	}
	return opts, nil
}

func repoToOut(r git.Repository) repositoryOut {
	rem := make([]remoteOut, 0, len(r.Remotes))
	for _, x := range r.Remotes {
		rem = append(rem, remoteOut{Name: x.Name, URL: x.URL})
	}
	branches := r.Branches
	if branches == nil {
		branches = []string{}
	}
	return repositoryOut{
		Path:         r.Path,
		Name:         r.Name,
		Remotes:      rem,
		Branches:     branches,
		IsDirty:      r.IsDirty,
		LastModified: r.LastModified,
		Selected:     r.Selected,
		Mode:         r.Mode.String(),
	}
}

func writeResponse(res response) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(res); err != nil {
		fallback := response{
			OK:    false,
			Error: fmt.Sprintf("failed writing response: %s", err.Error()),
		}
		stderrEnc := json.NewEncoder(os.Stderr)
		stderrEnc.SetEscapeHTML(false)
		if encodeErr := stderrEnc.Encode(fallback); encodeErr != nil {
			fmt.Fprintf(os.Stderr, "failed writing fallback response: %v\n", encodeErr)
		}
		return err
	}
	return nil
}