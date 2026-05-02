package main

import (
	"atns/config"
	"atns/finder"
	"fmt"
)

func main() {
	cfg := config.Load()

	repos := finder.FindRepos(cfg.SearchDirs)
	for i, r := range repos {
		fmt.Printf("Found repo #%d %v\n", i, r)
	}
}
