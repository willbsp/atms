package ui

// TODO implement a fuzzy finder
import (
	"atns/git"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"

	"github.com/sahilm/fuzzy"
)

var (
	normalStyle        = tcell.StyleDefault.Foreground(color.White)
	dimStyle           = tcell.StyleDefault.Foreground(color.Gray)
	selectedStyle      = tcell.StyleDefault.Foreground(color.Black).Background(color.Teal).Bold(true)
	headerStyle        = tcell.StyleDefault.Foreground(color.Teal).Bold(true)
	promptStyle        = tcell.StyleDefault.Foreground(color.Green).Bold(true)
	queryStyle         = tcell.StyleDefault.Foreground(color.White)
	dividerStyle       = tcell.StyleDefault.Foreground(color.DarkCyan)
	previewHeaderStyle = tcell.StyleDefault.Foreground(color.Teal).Bold(true)
)

type State struct {
	cursor        int
	query         string
	filteredRepos []git.Repo
}

func Run(repos []git.Repo) (string, error) {
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

	repoNames := mapRepoStrings(state.filteredRepos, func(r git.Repo) string {
		return filepath.Base(r.Path)
	})

	selectedRepo := state.filteredRepos[state.cursor]

	drawHeader(s, 1, 0)
	drawInputLine(s, 1, 1, state.query)
	drawList(s, 1, listTop, listHeight, repoNames, state.cursor, dividerX)
	drawFooter(s, 1, h-1)

	if dividerX < w {
		drawDivider(s, dividerX, h)
	}

	drawPreview(s, dividerX+3, 0, selectedRepo)
}

func drawPreview(s tcell.Screen, x, y int, repo git.Repo) {
	drawStr(s, x, y, previewHeaderStyle, "Preview")
	lines := []string{
		"",
		fmt.Sprintf("📂 %s", repo.Path),
		strings.Repeat("─", 40),
		"",
		fmt.Sprintf("Branch:  %s", repo.CurrentBranch),
		fmt.Sprintf("Commit:  %s", repo.LastCommit.Info),
		fmt.Sprintf("Author:  %s", repo.LastCommit.Author),
	}
	for i, line := range lines {
		drawStr(s, x+1, y+i+1, normalStyle, line)
	}
}

func drawList(s tcell.Screen, x, y, h int, repos []string, cursor int, dividerX int) {

	for i := 0; i < h && i < len(repos); i++ {
		repo := repos[i]
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

func handleKey(e *tcell.EventKey, repos []git.Repo, state *State) (bool, string) {
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
		return true, state.filteredRepos[state.cursor].Path
	case tcell.KeyEsc:
		return true, ""
	}

	return false, ""
}

func fuzzyFind(query string, repos []git.Repo) []git.Repo {
	if query == "" {
		return repos
	}

	paths := mapRepoStrings(repos, func(r git.Repo) string {
		return r.Path
	})

	results := fuzzy.Find(query, paths)
	filtered := make([]git.Repo, len(results))
	for i, r := range results {
		filtered[i] = repos[r.Index]
	}

	return filtered
}

func drawStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for i, c := range str {
		s.SetContent(x+i, y, c, nil, style)
	}
}

func mapRepoStrings(repos []git.Repo, fn func(git.Repo) string) []string {
	result := make([]string, len(repos))
	for i, r := range repos {
		result[i] = fn(r)
	}
	return result
}
