package finder

import (
	"log"
	"os"
	"path/filepath"
)

func FindRepoPaths(paths []string) <-chan string {
	ch := make(chan string)
	go walkPaths(paths, ch)
	return ch
}

func walkPaths(paths []string, pathCh chan<- string) {
	defer close(pathCh)
	for _, p := range paths {
		if isGitRepo(p) {
			pathCh <- p
			continue
		}

		entries, err := os.ReadDir(p)
		if err != nil {
			log.Printf("warning: unable to read directory %v\n", p)
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			child := filepath.Join(p, e.Name())
			if isGitRepo(child) {
				pathCh <- child
			}
		}
	}
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}
