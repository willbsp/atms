package main

import (
	"atns/config"
	"atns/finder"
	"atns/git"
	"atns/ui"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	repos := finder.FindRepos(cfg.SearchDirs)
	for _, repo := range repos {
		worktrees := git.GetWorktrees(repo)
		for _, worktree := range worktrees {
			if worktree.Path != repo {
				repos = append(repos, worktree.Path)
			}
		}
	}

	selected, _ := ui.Run(repos)
	fmt.Println(selected)
}
