package finder

import (
	"log"
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

		entries, err := os.ReadDir(p)
		if err != nil {
			log.Printf("warning: unable to read directory %v\n", p)
			continue
		}
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
