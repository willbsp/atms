package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Repo struct {
	Name       string
	Path       string
	Branch     string
	Status     Status
	LastCommit LastCommit
	Remotes    []Remote
	Worktrees  []Repo
	IsWorktree bool
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

func GetRepo(repoPath string) Repo {
	return Repo{
		Name:       filepath.Base(repoPath),
		Path:       repoPath,
		Branch:     getCurrentBranch(repoPath),
		LastCommit: getLastCommitInfo(repoPath),
		Remotes:    getRemotes(repoPath),
		Status:     getStatusSummary(repoPath),
	}
}

func GetWorktrees(repoPath string) []Repo {
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
			tree := GetRepo(p)
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
	description, _ := gitCmd(repoPath, "log", "-1", "--format=%h %s (%cr)")
	author, _ := gitCmd(repoPath, "log", "-1", "--format=%an")
	return LastCommit{Description: description, Author: author}
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
	status, _ := gitCmd(repoPath, "status", "--short")
	lines := strings.Split(strings.TrimSpace(status), "\n")
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

func gitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	o, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git command has failed %w", err)
	}

	return strings.TrimSpace(string(o)), nil
}
