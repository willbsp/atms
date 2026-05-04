package finder

import (
	"atns/git"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func FindRepos(paths []string) []git.Repo {
	var repoPaths []string

	for _, p := range paths {
		if isGitRepo(p) {
			repoPaths = append(repoPaths, p)
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
					repoPaths = append(repoPaths, path)
				}
			}
		}
	}

	repos := make([]git.Repo, len(repoPaths))
	var wg sync.WaitGroup
	for i, p := range repoPaths {
		wg.Add(1)
		go func(i int, p string) {
			defer wg.Done()
			repos[i] = getRepo(p)
		}(i, p)
	}
	wg.Wait()

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
