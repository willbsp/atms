package main

import (
	"atns/finder"
	"fmt"
)

func main() {
	repos := finder.FindRepos(([]string{"/Users/will/Developer/", "/Users/will/dotfiles"}))
	for i, r := range repos {
		fmt.Printf("Found repo #%d %v\n", i, r)
	}
}
