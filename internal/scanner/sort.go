package scanner

import (
	"github.com/chain710/manga/internal/arc"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

func less(x, y []int, xs, ys string) bool {
	var i int
	for i = 0; i < len(x) && i < len(y); i++ {
		if x[i] == y[i] {
			continue
		} else if x[i] < y[i] {
			return true
		} else {
			return false
		}
	}

	if len(x) != len(y) {
		return len(x) < len(y)
	} else {
		return xs < ys
	}
}

type digitsExtractor func(int) ([]int, string)

func extractVolumeMetaDigits(vols []volumeMeta) digitsExtractor {
	cache := make(map[string][]int)
	r := regexp.MustCompile(`\d+`)
	return func(i int) ([]int, string) {
		p := filepath.Base(vols[i].Path)
		x, ok := cache[p]
		if ok {
			return x, p
		}

		dlist := r.FindAllString(p, -1)
		digits := make([]int, len(dlist))
		for j, d := range dlist {
			val, err := strconv.Atoi(d)
			if err != nil {
				panic(err) // should not happen
			}
			digits[j] = val
		}
		cache[p] = digits
		return digits, p
	}
}

func extractArchiveFileDigits(files []arc.File) digitsExtractor {
	cache := make(map[string][]int)
	r := regexp.MustCompile(`\d+`)
	return func(i int) ([]int, string) {
		p := files[i].Name()
		x, ok := cache[p]
		if ok {
			return x, p
		}

		dlist := r.FindAllString(p, -1)
		digits := make([]int, len(dlist))
		for j, d := range dlist {
			val, err := strconv.Atoi(d)
			if err != nil {
				panic(err) // should not happen
			}
			digits[j] = val
		}
		cache[p] = digits
		return digits, p
	}
}

func SortSliceByDigit(slice interface{}, extract digitsExtractor) {
	sort.Slice(slice, func(i, j int) bool {
		di, si := extract(i)
		dj, sj := extract(j)
		return less(di, dj, si, sj)
	})
}
