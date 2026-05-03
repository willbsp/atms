package ui

// TODO implement a fuzzy finder
import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"

	"github.com/sahilm/fuzzy"
)

var (
	normalStyle   = tcell.StyleDefault.Foreground(color.White)
	dimStyle      = tcell.StyleDefault.Foreground(color.Gray)
	selectedStyle = tcell.StyleDefault.Foreground(color.Black).Background(color.Teal).Bold(true)
	headerStyle   = tcell.StyleDefault.Foreground(color.Teal).Bold(true)
	promptStyle   = tcell.StyleDefault.Foreground(color.Green).Bold(true)
	queryStyle    = tcell.StyleDefault.Foreground(color.White)
	dividerStyle  = tcell.StyleDefault.Foreground(color.DarkCyan)
)

type State struct {
	cursor        int
	query         string
	filteredRepos []string
}

func Run(repos []string) (string, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return "", fmt.Errorf("Failed to create screen")
	}
	if err := s.Init(); err != nil {
		return "", fmt.Errorf("Failed to initialise screen")
	}
	defer s.Fini()

	s.EnableMouse()
	s.Clear()

	state := State{cursor: 0, query: "", filteredRepos: repos}
	for {
		s.Clear()
		draw(s, &state)
		s.Show()

		e := <-s.EventQ()
		switch e := e.(type) {
		case *tcell.EventKey:
			shouldExit, selectedRepo := handleKey(e, repos, &state)
			if shouldExit {
				return selectedRepo, nil
			}
		case *tcell.EventResize:
			s.Sync()
		}
	}
}

func draw(s tcell.Screen, state *State) {
	w, h := s.Size()

	dividerX := w / 2
	if dividerX < 30 {
		dividerX = w
	}

	listTop := 2
	listHeight := h - 3

	drawHeader(s, 1, 0)
	drawInputLine(s, 1, 1, state.query)
	drawList(s, 1, listTop, listHeight, state.filteredRepos, state.cursor, dividerX)
	drawFooter(s, 1, h-1)

	if dividerX < w {
		drawDivider(s, dividerX, h)
	}
}

func drawList(s tcell.Screen, x, y, h int, repos []string, cursor int, dividerX int) {

	for i := 0; i < h && i < len(repos); i++ {
		repo := filepath.Base(repos[i])
		y := y + i

		if i == cursor {
			for x := range dividerX {
				s.SetContent(x, y, ' ', nil, selectedStyle)
			}
			drawStr(s, x, y, selectedStyle, "▸ ")
			drawStr(s, x+2, y, selectedStyle, repo)
		} else {
			drawStr(s, x, y, dimStyle, "  ")
			drawStr(s, x+2, y, normalStyle, repo)
		}
	}
}

func drawHeader(s tcell.Screen, x, y int) {
	title := "atns"
	drawStr(s, x, y, headerStyle, "  ")
	drawStr(s, x+2, y, headerStyle, title)
}

func drawFooter(s tcell.Screen, x, y int) {
	footer := " ↑↓ navigate • enter select • esc quit"
	drawStr(s, x, y, dimStyle, footer)
}

func drawInputLine(s tcell.Screen, x, y int, query string) {
	drawStr(s, x, y, promptStyle, "❯ ")
	drawStr(s, x+2, y, queryStyle, query)
	s.ShowCursor(x+2+len(query), y)
}

func drawDivider(s tcell.Screen, x, h int) {
	for y := range h {
		s.SetContent(x, y, '│', nil, dividerStyle)
	}
}

func handleKey(e *tcell.EventKey, repos []string, state *State) (bool, string) {
	switch e.Key() {
	case tcell.KeyUp:
		state.cursor = max(state.cursor-1, 0)
	case tcell.KeyDown:
		state.cursor = min(state.cursor+1, len(state.filteredRepos)-1)
	case tcell.KeyRune:
		state.query += e.Str()
		state.filteredRepos = fuzzyFind(state.query, repos)
		state.cursor = 0
	case tcell.KeyBackspace:
		if len(state.query) > 0 {
			runes := []rune(state.query)
			state.query = string(runes[:len(runes)-1])
			state.filteredRepos = fuzzyFind(state.query, repos)
			state.cursor = 0
		}
	case tcell.KeyEnter:
		return true, state.filteredRepos[state.cursor]
	case tcell.KeyEsc:
		return true, ""
	}

	return false, ""
}

func fuzzyFind(query string, list []string) []string {
	if query == "" {
		return list
	}

	results := fuzzy.Find(query, list)
	filtered := make([]string, len(results))
	for i, r := range results {
		filtered[i] = list[r.Index]
	}

	return filtered
}

func drawStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for i, c := range str {
		s.SetContent(x+i, y, c, nil, style)
	}
}
