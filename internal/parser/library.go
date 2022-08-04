package parser

import (
	"errors"
	"github.com/chain710/manga/internal/log"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type BookMeta struct {
	Name    BookNameMeta
	Volumes []VolumeMeta
	Path    string
	ModTime time.Time
	Extras  []VolumeMeta
}

type classifiedEntries struct {
	volumes     []fs.DirEntry // all should be files
	directories []fs.DirEntry // other directories
}

type LibraryWalker struct {
	Predict func(*BookMeta) bool // before parse volumes
	Handle  func(*BookMeta)
}

type LibraryWalkerFactory func() LibraryWalker

func WalkLibrary(root string, wf LibraryWalkerFactory, options ...Option) error {
	opt := DefaultOptions()
	for _, apply := range options {
		apply(&opt)
	}

	info, err := os.Lstat(root)
	if err != nil {
		log.Errorf("stat path %s error: %s", root, err)
		return err
	} else if !info.IsDir() {
		return errors.New("not directory")
	}

	return walkDir(root, info, wf, &opt)
}

func walkDir(root string, info os.FileInfo, wf LibraryWalkerFactory, options *Options) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	cf := options.classifyFiles(root, entries)
	if len(cf.volumes) == 0 {
		for _, entry := range cf.directories {
			path1 := filepath.Join(root, entry.Name())
			dirInfo, err := os.Lstat(path1)
			if err != nil {
				return err
			}
			if err := walkDir(path1, dirInfo, wf, options); err != nil {
				return err
			}
		}

		return nil
	}

	book := BookMeta{
		Name:    BookNameMeta{},
		Volumes: []VolumeMeta{},
		Path:    root,
		ModTime: info.ModTime(),
		Extras:  []VolumeMeta{},
	}

	bookName, err := ParseBookName(filepath.Base(root))
	if err != nil {
		log.Errorf("parse book name (%s) error: %s", root, err)
		return err
	}
	book.Name = *bookName

	walker := wf()
	if !walker.Predict(&book) {
		return nil
	}
	for _, f := range cf.volumes {
		volPath := filepath.Join(root, f.Name())
		vol, err := ParseBookVolume(volPath, options)
		if err != nil {
			log.Errorf("parse vol %s error: %s", volPath, err)
			continue // parse remaining files
		}

		if len(vol.Files) == 0 {
			log.Debugf("volume file is empty: %s", volPath)
			continue
		}

		book.Volumes = append(book.Volumes, *vol)
	}

	if options.SortVolumes != nil {
		options.SortVolumes(book.Volumes)
		for i := range book.Volumes {
			book.Volumes[i].ID = i + 1 // assign order id here
		}
	}

	for i := range cf.directories {
		entry := cf.directories[i]
		path1 := filepath.Join(root, entry.Name())
		if err := walkBookExtraDir(path1, &book, options); err != nil {
			return err
		}
	}

	if options.SortVolumes != nil {
		options.SortVolumes(book.Extras)
	}

	walker.Handle(&book)
	return nil
}

func walkBookExtraDir(root string, book *BookMeta, options *Options) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	cf := options.classifyFiles(root, entries)
	for i := range cf.volumes {
		// add to book extra
		volPath := filepath.Join(root, cf.volumes[i].Name())
		vol, err := ParseBookVolume(volPath, options)
		if err != nil {
			log.Errorf("parse extra vol %s error: %s", volPath, err)
			continue
		}

		if len(vol.Files) == 0 {
			log.Debugf("extra file is empty: %s", volPath)
			continue
		}

		book.Extras = append(book.Extras, *vol)
	}

	for i := range cf.directories {
		entry := cf.directories[i]
		path1 := filepath.Join(root, entry.Name())
		if err := walkBookExtraDir(path1, book, options); err != nil {
			return err
		}
	}

	return nil
}
