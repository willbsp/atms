package finder

import (
	"atns/git"
	"log"
	"os"
	"path/filepath"
)

func FindRepos(paths []string) []git.Repo {
	var repos []git.Repo

	for _, p := range paths {
		if isGitRepo(p) {
			repo := getRepo(p)
			repos = append(repos, repo)
			continue
		}

		entries, err := os.ReadDir(p)
		if err != nil {
			log.Printf("warning: unable to read directory %v\n", p)
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				path := filepath.Join(p, e.Name())
				if isGitRepo(path) {
					repo := getRepo(path)
					repos = append(repos, repo)
				}
			}
		}
	}

	return repos
}

func getRepo(path string) git.Repo {
	repo := git.GetRepo(path)
	repo.IsWorktree = false
	repo.Worktrees = git.GetWorktrees(path)
	return repo
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}
