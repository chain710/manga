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

type progressAlias struct {
	VolumeID    *int64  `db:"read_volume_id"`
	VolumeNo    *int    `db:"read_volume"`
	VolumeTitle *string `db:"read_volume_title"`
	FirstReadAt *Time   `db:"read_volume_begin_at"`
	LastReadAt  *Time   `db:"read_volume_at"`
	Page        *int    `db:"read_volume_page"`
	PageCount   *int    `db:"read_volume_page_count"`
}

func (p *progressAlias) convertBook(bookID int64) *BookProgress {
	if p.VolumeID == nil {
		return nil
	}
	return &BookProgress{
		BookID:    bookID,
		UpdateAt:  *p.LastReadAt,
		Volume:    *p.VolumeNo,
		VolumeID:  *p.VolumeID,
		Title:     *p.VolumeTitle,
		Page:      *p.Page,
		PageCount: *p.PageCount,
	}
}

func (p *progressAlias) convertVolume() *VolumeProgress {
	if p.VolumeID == nil {
		return nil
	}
	return &VolumeProgress{
		CreateAt: *p.FirstReadAt,
		UpdateAt: *p.LastReadAt,
		Page:     *p.Page,
	}
}

type bookJoin struct {
	BookTable
	progressAlias
}

func (j *bookJoin) convert(volumes []volumeJoin) *Book {
	var ret Book
	ret.BookTable = j.BookTable
	ret.Progress = j.progressAlias.convertBook(j.ID)
	for _, vol := range volumes {
		if vol.Volume > 0 {
			ret.Volumes = append(ret.Volumes, *vol.convert())
		} else {
			ret.Extras = append(ret.Extras, *vol.convert())
		}
	}
	return &ret
}

type VolumeTable struct {
	ID        int64          `db:"id" json:"id"`
	BookID    int64          `db:"book_id" json:"book_id"`
	CreateAt  Time           `db:"create_at" json:"create_at"`
	Path      string         `db:"path" json:"path"`
	Title     string         `db:"title" json:"title"`
	Volume    int            `db:"volume" json:"volume"` // 0 = extra
	PageCount int            `db:"page_count" json:"page_count"`
	Files     VolumeFileList `db:"files" json:"files,omitempty"`
}

type volumeJoin struct {
	VolumeTable
	progressAlias
	BookName string `db:"book_name" json:"book_name,omitempty"`
	Writer   string `db:"writer" json:"writer,omitempty"`
}

func (j *volumeJoin) convert() *Volume {
	return &Volume{
		VolumeTable: j.VolumeTable,
		BookName:    j.BookName,
		Writer:      j.Writer,
		Progress:    j.progressAlias.convertVolume(),
	}
}

type VolumeProgressTable struct {
	CreateAt Time  `db:"create_at" json:"create_at"`
	UpdateAt Time  `db:"update_at" json:"update_at"`
	BookID   int64 `db:"book_id" json:"book_id"`
	VolumeID int64 `db:"volume_id" json:"volume_id"`
	Complete bool  `db:"complete" json:"complete"`
	Page     int   `db:"page" json:"page"`
}
