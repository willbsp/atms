package main

import (
	"atns/config"
	"atns/finder"
	"atns/tmux"
	"atns/ui"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	repos := finder.FindRepos(cfg.SearchDirs)
	sessions := tmux.LoadSessions()

	selected, err := ui.Run(repos, sessions.Has)
	if err != nil {
		log.Fatal(err)
	}

	tmux.SwitchOrAttach(selected)
}
