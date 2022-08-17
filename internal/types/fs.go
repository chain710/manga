package types

type Directory struct {
	Parent  string     `json:"parent,omitempty"`
	Entries []DirEntry `json:"entries"`
}

const EntryTypeDirectory = "directory"

type DirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}
