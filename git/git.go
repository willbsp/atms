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
	LastCommit LastCommit
	Worktrees  []Repo
	IsWorktree bool
}

type LastCommit struct {
	Info   string
	Author string
}

func GetRepo(repoPath string) Repo {
	return Repo{
		Name:       filepath.Base(repoPath),
		Path:       repoPath,
		Branch:     getCurrentBranch(repoPath),
		LastCommit: getLastCommitInfo(repoPath),
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
	info, _ := gitCmd(repoPath, "log", "-1", "--format=%h %s (%cr)")
	author, _ := gitCmd(repoPath, "log", "-1", "--format=%an")
	return LastCommit{Info: info, Author: author}
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
