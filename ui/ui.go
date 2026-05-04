package ui

import (
	"atns/git"
	"fmt"
	"slices"
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
	treeStyle          = tcell.StyleDefault.Foreground(color.DarkCyan)
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

type RepoDiscoveredEvent struct {
	tcell.EventTime
	Repo git.Repo
}

func Run(repoCh <-chan git.Repo) (string, error) {
	s, err := initScreen()
	if err != nil {
		return "", err
	}
	defer s.Fini()
	streamRepos(s, repoCh)

	var discoveredRepos []git.Repo
	var listItems []ListItem
	state := State{cursor: 0, query: "", filteredItems: listItems}
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
			listItems = flattenReposToListItems(discoveredRepos)
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
	drawList(s, 0, listTop, listHeight, state.filteredItems, state.cursor, dividerX)
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
	const indent = 3

	style := normalStyle
	if selected {
		style = selectedStyle
		for i := range w {
			s.SetContent(x+i, y, ' ', nil, style)
		}
		s.PutStrStyled(x+1, y, "▸", style)
	}

	connector := "├─"
	if item.IsLast {
		connector = "└─"
	}

	s.PutStrStyled(x+indent*(item.Depth+1), y, item.Repo.Name, style)
	if item.Depth > 0 {
		s.PutStrStyled(x+(indent*item.Depth), y, connector, style)
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

func handleKey(e *tcell.EventKey, state *State, updateState func()) (bool, string) {
	switch e.Key() {
	case tcell.KeyUp:
		state.cursor = max(state.cursor-1, 0)
	case tcell.KeyDown:
		state.cursor = min(state.cursor+1, len(state.filteredItems)-1)
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
	results := fuzzy.Find(query, mapListItemStrings(listItems, func(r ListItem) string {
		return r.Repo.Name
	}))

	parentOf := func(idx int) int {
		for i := idx; i >= 0; i-- {
			if listItems[i].Depth == 0 {
				return i
			}
		}
		return idx
	}

	type child struct {
		index int
		score int
	}
	type group struct {
		bestScore int
		children  []child
	}
	groups := map[int]*group{}
	groupOrder := make([]int, 0, len(results))

	for _, r := range results {
		parentIdx := parentOf(r.Index)
		g, ok := groups[parentIdx]
		if !ok {
			g = &group{bestScore: r.Score}
			groups[parentIdx] = g
			groupOrder = append(groupOrder, parentIdx)
		}
		if r.Score > g.bestScore {
			g.bestScore = r.Score
		}
		if r.Index != parentIdx {
			g.children = append(g.children, child{index: r.Index, score: r.Score})
		}
	}

	slices.SortFunc(groupOrder, func(a, b int) int {
		return groups[b].bestScore - groups[a].bestScore
	})

	filtered := make([]ListItem, 0, len(results))
	for _, parentIdx := range groupOrder {
		g := groups[parentIdx]
		filtered = append(filtered, listItems[parentIdx])
		slices.SortFunc(g.children, func(a, b child) int {
			return b.score - a.score
		})
		for i, c := range g.children {
			item := listItems[c.index]
			item.IsLast = i == len(g.children)-1
			filtered = append(filtered, item)
		}
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
