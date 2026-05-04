package ui

import (
	"atns/git"
	"fmt"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v3"
)

type RepoDiscoveredEvent struct {
	tcell.EventTime
	Repo git.Repo
}

func Run(repoCh <-chan git.Repo, hasSession func(repoPath string) bool) (string, error) {
	s, err := initScreen()
	if err != nil {
		return "", err
	}
	defer s.Fini()
	streamRepos(s, repoCh)

	var discoveredRepos []git.Repo
	var listItems []ListItem
	state := State{}
	updateFilteredItems := func() {
		state.filteredItems = fuzzyFind(state.query, listItems)
	}
	for {
		s.Clear()
		draw(s, &state)
		s.Show()

		e := <-s.EventQ()
		switch e := e.(type) {
		case *tcell.EventKey:
			shouldExit, selectedRepo := handleKey(e, &state, updateFilteredItems)
			if shouldExit {
				return selectedRepo, nil
			}
		case *tcell.EventResize:
			s.Sync()
		case *RepoDiscoveredEvent:
			idx, _ := slices.BinarySearchFunc(discoveredRepos, e.Repo, func(a, b git.Repo) int {
				return strings.Compare(
					strings.ToLower(a.Name),
					strings.ToLower(b.Name),
				)
			})
			discoveredRepos = slices.Insert(discoveredRepos, idx, e.Repo)
			listItems = flattenReposToListItems(discoveredRepos, hasSession)
			updateFilteredItems()
		}
	}
}

func initScreen() (tcell.Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen")
	}
	if err := s.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialise screen")
	}
	s.EnableMouse()
	s.Clear()
	return s, nil
}

func streamRepos(s tcell.Screen, repoCh <-chan git.Repo) {
	go func() {
		for repo := range repoCh {
			ev := RepoDiscoveredEvent{Repo: repo}
			ev.SetEventNow()
			s.EventQ() <- &ev
		}
	}()
}

func flattenReposToListItems(repos []git.Repo, hasSession func(repoPath string) bool) []ListItem {
	var result []ListItem
	for _, r := range repos {
		result = append(result, ListItem{
			Repo:    r,
			Depth:   0,
			IsLast:  false,
			Session: hasSession(r.Path),
		})
		for i, w := range r.Worktrees {
			result = append(result, ListItem{
				Repo:    w,
				Depth:   1,
				IsLast:  i == len(r.Worktrees)-1,
				Session: hasSession(w.Path),
			})
		}
	}
	return result
}
