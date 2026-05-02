package finder

import (
	"os"
	"path/filepath"
)

func FindRepos(paths []string) []string {
	var repos []string

	for _, p := range paths {
		if isGitRepo(p) {
			repos = append(repos, p)
			continue
		}

		entries, _ := os.ReadDir(p)
		for _, e := range entries {
			if e.IsDir() {
				path := filepath.Join(p, e.Name())
				if isGitRepo(path) {
					repos = append(repos, path)
				}
			}
		}
	}

	return repos
}

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}
