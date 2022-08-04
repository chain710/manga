package strings

import "sort"

type empty struct{}

type Set struct {
	m    map[string]empty
	conv func(string) string
}

func NewSet(conv func(string) string, strings ...string) Set {
	if conv == nil {
		conv = dummy
	}
	set := Set{
		m:    make(map[string]empty),
		conv: conv,
	}

	for _, str := range strings {
		set.m[conv(str)] = struct{}{}
	}
	return set
}

func (s Set) Add(values ...string) {
	for _, v := range values {
		s.m[s.conv(v)] = empty{}
	}
}

func (s Set) Remove(str string) {
	delete(s.m, s.conv(str))
}

func (s Set) Contains(str string) bool {
	_, ok := s.m[s.conv(str)]
	return ok
}

func (s Set) Len() int {
	return len(s.m)
}

func (s Set) SortedList() []string {
	ss := make([]string, 0, len(s.m))
	for k := range s.m {
		ss = append(ss, k)
	}

	sort.Strings(ss)
	return ss
}

func dummy(in string) string {
	return in
}
