package parser

import (
	"github.com/chain710/manga/internal/arc"
	"sort"
)

type sortByName []arc.File

func (s sortByName) Len() int {
	return len(s)
}

func (s sortByName) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

func (s sortByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func SortByName(files []arc.File) {
	sort.Sort(sortByName(files))
}
