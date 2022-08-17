package scanner

import "time"

type Option func(p *Type)

func IgnoreBookModTime(v bool) Option {
	return func(p *Type) {
		p.ignoreBookModTime = v
	}
}

type ScanBookOption func(o *ScanBookOptions)

type ScanBookOptions struct {
	libraryID int64
	path      string
	modTime   time.Time
	entries   *classifiedEntries
}

func scanBookOptions(option ScanBookOptions) ScanBookOption {
	return func(o *ScanBookOptions) {
		*o = option
	}
}
