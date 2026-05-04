package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Repo struct {
	Name           string
	Path           string
	Branch         string
	RecentBranches []string
	Status         Status
	LastCommit     LastCommit
	Remotes        []Remote
	Worktrees      []Repo
	IsWorktree     bool
}

type LastCommit struct {
	Description string
	Author      string
}

type Status struct {
	Staged    int
	Unstaged  int
	Untracked int
}

type Remote struct {
	Name string
	Url  string
}

func (s Status) IsClean() bool {
	return s.Staged == 0 && s.Unstaged == 0 && s.Untracked == 0
}

func GetRepo(repoPath string) Repo {
	return getRepo(repoPath, true)
}

func getRepo(repoPath string, fetchWorktrees bool) Repo {
	repo := Repo{
		Name:       filepath.Base(repoPath),
		Path:       repoPath,
		IsWorktree: false,
	}
	var wg sync.WaitGroup
	wg.Go(func() { repo.Branch = getCurrentBranch(repoPath) })
	wg.Go(func() { repo.RecentBranches = getRecentBranches(repoPath, 5) })
	wg.Go(func() { repo.LastCommit = getLastCommitInfo(repoPath) })
	wg.Go(func() { repo.Remotes = getRemotes(repoPath) })
	wg.Go(func() { repo.Status = getStatusSummary(repoPath) })
	if fetchWorktrees {
		wg.Go(func() { repo.Worktrees = getWorktrees(repoPath) })
	}
	wg.Wait()
	return repo
}

func getWorktrees(repoPath string) []Repo {
	o, _ := gitCmd(repoPath, "worktree", "list", "--porcelain")
	var treePaths []string

	for line := range strings.SplitSeq(string(o), "\n") {
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			if path == repoPath {
				continue
			}
			treePaths = append(treePaths, path)
		}
	}

	trees := make([]Repo, len(treePaths))
	var wg sync.WaitGroup
	for i, p := range treePaths {
		wg.Add(1)
		go func(i int, p string) {
			defer wg.Done()
			tree := getRepo(p, false)
			tree.IsWorktree = true
			trees[i] = tree
		}(i, p)
	}
	wg.Wait()

	return trees
}

func getCurrentBranch(repoPath string) string {
	branch, _ := gitCmd(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	return branch
}

func getLastCommitInfo(repoPath string) LastCommit {
	o, _ := gitCmd(repoPath, "log", "-1", "--format=%h %s (%cr)%x1f%an")
	parts := strings.SplitN(o, "\x1f", 2)
	if len(parts) != 2 {
		return LastCommit{}
	}
	return LastCommit{Description: parts[0], Author: parts[1]}
}

func getRemotes(repoPath string) []Remote {
	o, _ := gitCmd(repoPath, "remote", "-v")
	if o == "" {
		return []Remote{}
	}
	lines := strings.Split(strings.TrimSpace(o), "\n")
	seen := make(map[string]bool)
	var remotes []Remote
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && !seen[parts[0]] {
			seen[parts[0]] = true
			remotes = append(remotes, Remote{parts[0], parts[1]})
		}
	}
	return remotes
}

func getStatusSummary(repoPath string) Status {
	o, _ := gitCmd(repoPath, "status", "--short")
	lines := strings.Split(strings.TrimSpace(o), "\n")
	staged, unstaged, untracked := 0, 0, 0
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		x, y := line[0], line[1]
		if x == '?' && y == '?' {
			untracked++
			continue
		}
		if x != ' ' && x != '?' {
			staged++
		}
		if y != ' ' && y != '?' {
			unstaged++
		}
	}
	return Status{staged, unstaged, untracked}
}

func getRecentBranches(repoPath string, limit int) []string {
	o, _ := gitCmd(repoPath, "branch", "--sort=-committerdate", "--format=%(refname:short)")
	if strings.TrimSpace(o) == "" {
		return []string{}
	}
	branches := strings.Split(strings.TrimSpace(o), "\n")
	if len(branches) > limit {
		return branches[:limit]
	}
	return branches
}

func gitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	o, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git command has failed %w", err)
	}

	return strings.TrimSpace(string(o)), nil
}
