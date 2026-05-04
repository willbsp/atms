package ui

import (
	"slices"

	"github.com/sahilm/fuzzy"
)

func fuzzyFind(query string, listItems []ListItem) []ListItem {
	if query == "" {
		return listItems
	}

	mapListItemStrings := func(repos []ListItem) []string {
		result := make([]string, len(repos))
		for i, r := range repos {
			result[i] = r.Repo.Name
		}
		return result
	}

	results := fuzzy.Find(query, mapListItemStrings(listItems))

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
