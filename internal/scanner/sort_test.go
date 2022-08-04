package scanner

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"strconv"
	"testing"
)

func extractStringDigits(slice []string) digitsExtractor {
	r := regexp.MustCompile(`\d+`)
	return func(i int) ([]int, string) {
		dlist := r.FindAllString(slice[i], -1)
		digits := make([]int, len(dlist))
		for j, d := range dlist {
			val, err := strconv.Atoi(d)
			if err != nil {
				panic(err) // should not happen
			}
			digits[j] = val
		}
		return digits, slice[i]
	}
}

func TestSortByDigit(t *testing.T) {
	tests := []struct {
		name   string
		in     []string
		expect []string
	}{
		{
			name:   "normal",
			in:     []string{"2", "1", "1.1"},
			expect: []string{"1", "1.1", "2"},
		},
		{
			name:   "with alphabet",
			in:     []string{"a2", "a1", "a1.1"},
			expect: []string{"a1", "a1.1", "a2"},
		},
		{
			name:   "mixed",
			in:     []string{"a2", "a1", "z", "b", "f"},
			expect: []string{"b", "f", "z", "a1", "a2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortSliceByDigit(tt.in, extractStringDigits(tt.in))
			require.Equal(t, tt.expect, tt.in)
		})
	}
}
