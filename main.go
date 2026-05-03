package main

import (
	"atns/config"
	"atns/finder"
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
	selected, _ := ui.Run(repos)
	fmt.Println(selected)
}
