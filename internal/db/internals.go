package db

type BookTable struct {
	ID        int64  `db:"id" json:"id"`
	LibraryID int64  `db:"library_id" json:"library_id"`
	CreateAt  Time   `db:"create_at" json:"create_at"`
	UpdateAt  Time   `db:"update_at" json:"update_at"`
	PathModAt Time   `db:"path_mod_at" json:"-"`
	Path      string `db:"path" json:"path"`
	Name      string `db:"name" json:"name"`
	Writer    string `db:"writer" json:"writer"`
	Volume    int    `db:"volume" json:"volume"`
	Summary   string `db:"summary" json:"summary"`
}

type bookProgressAlias struct {
	VolumeID    *int64  `db:"read_volume_id"`
	VolumeNo    *int    `db:"read_volume"`
	VolumeTitle *string `db:"read_volume_title"`
	FirstReadAt *Time   `db:"read_volume_begin_at"`
	LastReadAt  *Time   `db:"read_volume_at"`
	Page        *int    `db:"read_volume_page"`
	PageCount   *int    `db:"read_volume_page_count"`
}

type bookJoin struct {
	BookTable
	bookProgressAlias
}
