package scanner

type Option func(p *Type)

func IgnoreBookModTime(v bool) Option {
	return func(p *Type) {
		p.ignoreBookModTime = v
	}
}
