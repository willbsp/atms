package main

import (
	"atns/config"
	"atns/finder"
	"atns/git"
	"atns/tmux"
	"atns/ui"
	"log"
	"path/filepath"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	repos := finder.FindRepos(cfg.SearchDirs)
	sessions := tmux.LoadSessions()

	hasSession := func(r git.Repo) bool {
		return sessions.Has(sessionTarget{r})
	}

	selected, err := ui.Run(repos, hasSession)
	if err != nil {
		log.Fatal(err)
	}

	if selected.Path != "" {
		err = tmux.SwitchOrAttach(sessionTarget{selected})
		if err != nil {
			log.Fatal(err)
		}
	}
}

type sessionTarget struct {
	git.Repo
}

func (r sessionTarget) Name() string {
	if r.IsWorktree && r.Parent != "" && r.Branch != "" {
		return r.Parent + "/" + r.Branch
	}
	return filepath.Base(r.Path)
}

func (r sessionTarget) Dir() string {
	return r.Path
}
