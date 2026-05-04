package main

import (
	"atns/config"
	"atns/finder"
	"atns/tmux"
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
	sessions := tmux.ListSessions()

	selected, err := ui.Run(repos, sessions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(selected)
}
