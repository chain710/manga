package types

type PageKey struct {
	Volume string // volume's hash
	Page   int
}

type Image struct {
	Data   []byte
	Hash   string
	Format string
	H, W   int
}
