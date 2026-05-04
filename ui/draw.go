package ui

import (
	"atns/git"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
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
		fmt.Sprint(repo.Path),
		strings.Repeat("─", 40),
		"",
		fmt.Sprintf("Branch:  %s", repo.Branch),
		fmt.Sprintf("Commit:  %s", repo.LastCommit.Description),
		fmt.Sprintf("Author:  %s", repo.LastCommit.Author),
	}
	if !repo.Status.IsClean() {
		lines = append(lines, fmt.Sprintf("Status:  %d staged, %d unstaged, %d untracked", repo.Status.Staged, repo.Status.Unstaged, repo.Status.Untracked))
	} else {
		lines = append(lines, "Status:  Clean ✓")
	}
	if len(repo.Remotes) > 0 {
		lines = append(lines, "", "Remotes: ")
		for _, r := range repo.Remotes {
			lines = append(lines, fmt.Sprintf("• %v -> %v", r.Name, r.Url))
		}
	}
	if len(repo.RecentBranches) > 0 {
		lines = append(lines, "", "Recent branches: ")
		for _, b := range repo.RecentBranches {
			lines = append(lines, fmt.Sprintf("• %v", b))
		}
	}
	if len(repo.Worktrees) > 0 {
		lines = append(lines, "", "Worktrees: ")
		for _, w := range repo.Worktrees {
			lines = append(lines, fmt.Sprintf("• %v (%v)", w.Path, w.Branch))
		}
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

	name := item.Repo.Name
	if item.Repo.IsWorktree {
		name = fmt.Sprintf("%v (%v)", item.Repo.Name, item.Repo.Branch)
	}
	if item.Session {
		name = fmt.Sprintf("%v •", name)
	}

	s.PutStrStyled(x+indent*(item.Depth+1), y, name, style)
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
