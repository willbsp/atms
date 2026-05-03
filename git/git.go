package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Worktree struct {
	Path   string
	Head   string
	Branch string
}

func GetCurrentBranch(repoPath string) string {
	branch, _ := gitCmd(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	return branch
}

func GetLastCommitInfo(repoPath string) (string, string) {
	lastCommit, _ := gitCmd(repoPath, "log", "-1", "--format=%h %s (%cr)")
	author, _ := gitCmd(repoPath, "log", "-1", "--format=%an")
	return lastCommit, author
}

func GetWorktrees(repoPath string) []Worktree {
	o, _ := gitCmd(repoPath, "worktree", "list", "--porcelain")
	var trees []Worktree
	var current Worktree

	for line := range strings.SplitSeq(string(o), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "":
			if current.Path != "" {
				trees = append(trees, current)
				current = Worktree{}
			}
		}
	}

	if current.Path != "" {
		trees = append(trees, current)
	}

	return trees
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
