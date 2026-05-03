package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
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
	var trees []Repo

	for line := range strings.SplitSeq(string(o), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			if path == repoPath {
				continue
			}
			repo := GetRepo(path)
			repo.IsWorktree = true
			trees = append(trees, repo)
		}
	}

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
