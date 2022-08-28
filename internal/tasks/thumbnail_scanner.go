package tasks

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/imagehelper"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/util"
	"github.com/chain710/workqueue"
	"github.com/disintegration/imaging"
)

type ThumbnailOption func(*ThumbnailScanner)

func ThumbWithSize(width, height int) ThumbnailOption {
	return func(scanner *ThumbnailScanner) {
		scanner.thumbWidth = width
		scanner.thumbHeight = height
	}
}

func ThumbWithArchiveCache(archiveCache *arc.ArchiveCache) ThumbnailOption {
	return func(scanner *ThumbnailScanner) {
		scanner.archiveCache = archiveCache
	}
}

func NewThumbnailScanner(db db.Interface, options ...ThumbnailOption) *ThumbnailScanner {
	d := &ThumbnailScanner{
		database: db,
		q:        workqueue.NewRetryQueue("thumb", clk),
		options: imagehelper.VolumeThumbnailOptions{
			SampleHeight:   100,
			HeadCandidates: 3,
			TailCandidates: 2,
		},
		thumbWidth:  210,
		thumbHeight: 297,
		notify:      make(chan struct{}, 1),
	}
	for _, apply := range options {
		apply(d)
	}

	return d
}

type ThumbnailScanner struct {
	database     db.Interface
	q            workqueue.RetryInterface
	archiveCache *arc.ArchiveCache
	options      imagehelper.VolumeThumbnailOptions
	thumbWidth   int
	thumbHeight  int
	notify       chan struct{}
}

type thumbOfVolumeIndex int64

type thumbOfVolume struct {
	volume db.Volume
}

func (i thumbOfVolume) IsReplaceable() bool {
	return true
}

func (i thumbOfVolume) Index() interface{} {
	return thumbOfVolumeIndex(i.volume.ID)
}

type thumbOfBookIndex int64

type thumbOfBook struct {
	book db.Book
}

func (i thumbOfBook) IsReplaceable() bool {
	return true
}

func (i thumbOfBook) Index() interface{} {
	return thumbOfBookIndex(i.book.ID)
}

func (d *ThumbnailScanner) Start(ctx context.Context) {
	d.Scan() // scan at startup
	go d.workloop(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Debugf("stopping thumb scanner")
			d.q.ShutDown()
			break
		case <-d.notify:
			_ = d.listVolumes(ctx, false)
			_ = d.listBooks(ctx)
		}
	}
}

func (d *ThumbnailScanner) Scan() {
	select {
	case d.notify <- struct{}{}:
		log.Debugf("notify thumb scan ok")
	default:
		log.Debugf("notify thumb scan miss")
	}
}

func (d *ThumbnailScanner) ScanBook(books ...db.Book) {
	for i := range books {
		_ = d.q.Add(thumbOfBook{book: books[i]})
	}
}

func (d *ThumbnailScanner) ScanVolumes(vols ...db.Volume) {
	for i := range vols {
		_ = d.q.Add(thumbOfVolume{volume: vols[i]})
	}
}

func (d *ThumbnailScanner) Once(ctx context.Context, all bool) error {
	if err := d.listVolumes(ctx, all); err != nil {
		return err
	}
	if err := d.listBooks(ctx); err != nil {
		return err
	}
	d.q.ShutDown()
	d.workloop(ctx)
	return nil
}

func (d *ThumbnailScanner) listVolumes(ctx context.Context, all bool) error {
	opt := db.ListVolumesOptions{Join: db.VolumeMustNotHaveThumb}
	if all {
		opt.Join = ""
	}
	volumes, err := d.database.ListVolumes(ctx, opt)
	if err != nil {
		log.Errorf("list volumes error: %s", err)
		return err
	}

	for i := range volumes {
		_ = d.q.Add(thumbOfVolume{volume: volumes[i]})
	}
	return nil
}

func (d *ThumbnailScanner) listBooks(ctx context.Context) error {
	books, _, err := d.database.ListBooks(ctx, db.ListBooksOptions{Join: db.ListBookWithoutThumbnail})
	if err != nil {
		log.Errorf("list books error: %s", err)
		return err
	}

	d.ScanBook(books...)
	return nil
}

func (d *ThumbnailScanner) workloop(ctx context.Context) {
	for {
		item, shutdown := d.q.Get()
		if shutdown {
			log.Infof("volume thumb shutdown")
			return
		}

		switch b := item.(type) {
		case thumbOfVolume:
			d.scanVolume(ctx, &b.volume)
		case thumbOfBook:
			d.scanBook(ctx, &b.book)
		default:
			panic(fmt.Sprintf("unknown type %v", b))
		}

	}
}

func (d *ThumbnailScanner) scanVolume(ctx context.Context, volume *db.Volume) {
	logger := log.With("volume", volume.Path)
	fh, err := util.FileHash(volume.Path)
	if err != nil {
		logger.Errorf("gen file hash error: %s", err)
		return
	}

	archive, err := arc.Open(volume.Path, arc.OpenWithCache(d.archiveCache, fh))
	if err != nil {
		logger.Errorf("open archive error: %s", err)
		return
	}

	//goland:noinspection GoUnhandledErrorResult
	defer archive.Close()
	img, err := imagehelper.GetVolumeCover(archive, volume.Files, d.options)
	if err != nil {
		logger.Errorf("generate vol thumb error: %s", err)
		return
	}

	img = imaging.Thumbnail(img, d.thumbWidth, d.thumbHeight, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG); err != nil {
		logger.Errorf("encode thumb to jpeg error: %s", err)
		return
	}

	vt := db.VolumeThumbnail{
		ID:        volume.ID,
		Hash:      util.ImageHash(buf.Bytes()),
		Thumbnail: buf.Bytes(),
	}
	if err := d.database.SetVolumeThumbnail(ctx, vt); err != nil {
		logger.Errorf("set volume thumb error: %s", err)
		return
	}

	logger.Debugf("set volume thumb ok, size=%d", buf.Len())
}

func (d *ThumbnailScanner) scanBook(ctx context.Context, book *db.Book) {
	logger := log.With("book", book.Path)
	thumb, err := d.database.GetVolumeThumbnail(ctx, db.GetVolumeThumbOptions{BookID: &book.ID})
	if err != nil {
		logger.Errorf("get volume thumb error: %s", err)
		return
	}

	if thumb == nil {
		logger.Debugf("volume thumb not found")
		return
	}

	bt := db.BookThumbnail{ID: book.ID, Hash: thumb.Hash, Thumbnail: thumb.Thumbnail}
	if err := d.database.SetBookThumbnail(ctx, bt); err != nil {
		logger.Errorf("set book thumb (vol %d) error: %s", thumb.ID, err)
		return
	}

	logger.Debugf("set book thumbnail ok, size=%d", len(thumb.Thumbnail))
}
