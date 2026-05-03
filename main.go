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

	dirs := finder.FindRepoDirs(cfg.SearchDirs)
	repos := make([]git.Repo, len(dirs))
	for i, dir := range dirs {
		repos[i] = git.GetRepoInfo(dir)
	}

	selected, err := ui.Run(repos)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(selected)
}
