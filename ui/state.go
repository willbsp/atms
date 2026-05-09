package ui

import (
	"atms/git"

	"github.com/gdamore/tcell/v3"
)

type State struct {
	cursor        int
	query         string
	filteredItems []ListItem
}

type ListItem struct {
	Repo    git.Repo
	Session bool
	Depth   int
	IsLast  bool
}

func handleKey(e *tcell.EventKey, state *State, updateState func()) (bool, git.Repo) {
	switch e.Key() {
	case tcell.KeyUp:
		state.cursor = max(state.cursor-1, 0)
	case tcell.KeyDown:
		state.cursor = min(state.cursor+1, max(0, len(state.filteredItems)-1))
	case tcell.KeyRune:
		state.query += e.Str()
		state.cursor = 0
		updateState()
	case tcell.KeyBackspace:
		if len(state.query) > 0 {
			runes := []rune(state.query)
			state.query = string(runes[:len(runes)-1])
			state.cursor = 0
			updateState()
		}
	case tcell.KeyEnter:
		if state.cursor >= 0 && state.cursor < len(state.filteredItems) {
			return true, state.filteredItems[state.cursor].Repo
		}
	case tcell.KeyEsc:
		return true, git.Repo{}
	}

	return false, git.Repo{}
}
