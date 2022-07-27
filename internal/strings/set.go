package strings

type empty struct{}

type Set map[string]empty

func NewSet(strings ...string) Set {
	set := make(map[string]empty, len(strings))
	for _, str := range strings {
		set[str] = struct{}{}
	}
	return set
}

func (s Set) Add(values ...string) {
	for _, v := range values {
		s[v] = empty{}
	}
}

func (s Set) Remove(str string) {
	delete(s, str)
}

func (s Set) Contains(str string) bool {
	_, ok := s[str]
	return ok
}

func (s Set) Len() int {
	return len(s)
}
