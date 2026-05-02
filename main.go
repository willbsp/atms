package main

import (
	"atns/config"
	"atns/finder"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	repos := finder.FindRepos(cfg.SearchDirs)
	for i, r := range repos {
		fmt.Printf("Found repo #%d %v\n", i, r)
	}
}
