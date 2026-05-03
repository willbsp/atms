package ui

// TODO implement a fuzzy finder
import (
	"atns/git"
	"fmt"
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
	filteredItems []ListItem
}

type ListItem struct {
	Repo   git.Repo
	Depth  int
	IsLast bool
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

	listItems := flattenReposToListItems(repos)

	state := State{cursor: 0, query: "", filteredItems: listItems}
	for {
		s.Clear()
		draw(s, &state)
		s.Show()

		e := <-s.EventQ()
		switch e := e.(type) {
		case *tcell.EventKey:
			shouldExit, selectedRepo := handleKey(e, listItems, &state)
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
	drawList(s, 1, listTop, listHeight, state.filteredItems, state.cursor, dividerX)
	drawFooter(s, 1, h-1)

	if dividerX < w {
		drawDivider(s, dividerX, h)
	}

	if len(state.filteredItems) > 0 {
		selectedItem := state.filteredItems[state.cursor]
		drawPreview(s, dividerX+3, 0, selectedItem.Repo)
	}
}

func drawPreview(s tcell.Screen, x, y int, repo git.Repo) {
	s.PutStrStyled(x, y, "Preview", previewHeaderStyle)
	lines := []string{
		"",
		fmt.Sprintf("📂 %s", repo.Path),
		strings.Repeat("─", 40),
		"",
		fmt.Sprintf("Branch:  %s", repo.Branch),
		fmt.Sprintf("Commit:  %s", repo.LastCommit.Info),
		fmt.Sprintf("Author:  %s", repo.LastCommit.Author),
	}
	for i, line := range lines {
		s.PutStrStyled(x+1, y+i+1, line, normalStyle)
	}
}

func drawList(s tcell.Screen, x, y, h int, items []ListItem, cursor int, dividerX int) {
	for i := 0; i < h && i < len(items); i++ {
		drawListItem(s, x, y+i, dividerX, items[i], i == cursor)
	}
}

func drawListItem(s tcell.Screen, x, y, w int, item ListItem, selected bool) {
	if selected {
		for x := range w {
			s.SetContent(x, y, ' ', nil, selectedStyle)
		}
		s.PutStrStyled(x, y, "▸", selectedStyle)
		s.PutStrStyled(x+2, y, item.Repo.Name, selectedStyle)
	} else {
		s.PutStrStyled(x+2, y, item.Repo.Name, normalStyle)
	}
}

func drawHeader(s tcell.Screen, x, y int) {
	title := "atns"
	s.PutStrStyled(x+2, y, title, headerStyle)
}

func drawFooter(s tcell.Screen, x, y int) {
	footer := " ↑↓ navigate • enter select • esc quit"
	s.PutStrStyled(x, y, footer, dimStyle)
}

func drawInputLine(s tcell.Screen, x, y int, query string) {
	s.PutStrStyled(x, y, "❯ ", promptStyle)
	s.PutStrStyled(x+2, y, query, queryStyle)
	s.ShowCursor(x+2+len(query), y)
}

func drawDivider(s tcell.Screen, x, h int) {
	for y := range h {
		s.SetContent(x, y, '│', nil, dividerStyle)
	}
}

func handleKey(e *tcell.EventKey, listItems []ListItem, state *State) (bool, string) {
	switch e.Key() {
	case tcell.KeyUp:
		state.cursor = max(state.cursor-1, 0)
	case tcell.KeyDown:
		state.cursor = min(state.cursor+1, len(state.filteredItems)-1)
	case tcell.KeyRune:
		state.query += e.Str()
		state.filteredItems = fuzzyFind(state.query, listItems)
		state.cursor = 0
	case tcell.KeyBackspace:
		if len(state.query) > 0 {
			runes := []rune(state.query)
			state.query = string(runes[:len(runes)-1])
			state.filteredItems = fuzzyFind(state.query, listItems)
			state.cursor = 0
		}
	case tcell.KeyEnter:
		return true, state.filteredItems[state.cursor].Repo.Path
	case tcell.KeyEsc:
		return true, ""
	}

	return false, ""
}

func fuzzyFind(query string, listItems []ListItem) []ListItem {
	if query == "" {
		return listItems
	}

	paths := mapListItemStrings(listItems, func(r ListItem) string {
		return r.Repo.Path
	})

	results := fuzzy.Find(query, paths)
	filtered := make([]ListItem, len(results))
	for i, r := range results {
		filtered[i] = listItems[r.Index]
	}

	return filtered
}

func mapListItemStrings(repos []ListItem, fn func(ListItem) string) []string {
	result := make([]string, len(repos))
	for i, r := range repos {
		result[i] = fn(r)
	}
	return result
}

func flattenReposToListItems(repos []git.Repo) []ListItem {
	var result []ListItem
	for _, r := range repos {
		result = append(result, ListItem{
			Repo:   r,
			Depth:  0,
			IsLast: false,
		})
		for i, w := range r.Worktrees {
			result = append(result, ListItem{
				Repo:   w,
				Depth:  1,
				IsLast: i == len(r.Worktrees)-1,
			})
		}
	}
	return result
}
