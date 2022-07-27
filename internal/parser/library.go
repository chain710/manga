package parser

import (
	"errors"
	"github.com/chain710/manga/internal/log"
	internalstrings "github.com/chain710/manga/internal/strings"
	"io/fs"
	"os"
	"path/filepath"
)

type BookMeta struct {
	Name    BookNameMeta
	Volumes []BookVolumeBasicMeta
	Path    string
	Extras  []BookVolumeBasicMeta
}

type LibraryOptions struct {
	AcceptFileTypes   internalstrings.Set
	AcceptHiddenFiles bool
}

func (p *LibraryOptions) filterAcceptFiles(root string, files []fs.DirEntry) (accepted []fs.DirEntry, notAccepted []fs.DirEntry) {
	for i := range files {
		file := files[i]
		path := filepath.Join(root, file.Name())
		if !p.AcceptHiddenFiles {
			if isHidden, err := isHiddenFile(path); err != nil || isHidden {
				if err != nil {
					log.Errorf("determine file %s hidden error: %s", path, err)
				}
				notAccepted = append(notAccepted, file)
				continue
			}
		}

		ext := filepath.Ext(file.Name())
		if !file.IsDir() && (p.AcceptFileTypes.Len() == 0 || p.AcceptFileTypes.Contains(ext)) {
			accepted = append(accepted, file)
		} else {
			notAccepted = append(notAccepted, file)
		}
	}

	return
}

type LibraryOption func(*LibraryOptions)

func LibraryOptionDefault(options *LibraryOptions) {
	options.AcceptFileTypes = internalstrings.NewSet(".zip", ".rar", ".7z")
}

func LibraryOptionAcceptFileTypes(types ...string) LibraryOption {
	return func(options *LibraryOptions) {
		options.AcceptFileTypes.Add(types...)
	}
}

type WalkBookFunc func(b *BookMeta)

func WalkLibrary(root string, fn WalkBookFunc, options ...LibraryOption) error {
	opt := LibraryOptions{
		AcceptFileTypes: internalstrings.NewSet(),
	}
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

	return walkDir(root, fn, &opt)
}

func walkDir(root string, fn WalkBookFunc, options *LibraryOptions) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	acceptFiles, notAccept := options.filterAcceptFiles(root, entries)
	if len(acceptFiles) == 0 {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			path1 := filepath.Join(root, entry.Name())
			if err := walkDir(path1, fn, options); err != nil {
				return err
			}
		}

		return nil
	}

	book := BookMeta{
		Name:    BookNameMeta{},
		Volumes: []BookVolumeBasicMeta{},
		Path:    root,
		Extras:  []BookVolumeBasicMeta{},
	}

	bookName, err := ParseBookName(filepath.Base(root))
	if err != nil {
		log.Errorf("parse book name (%s) error: %s", root, err)
		return err
	}

	book.Name = *bookName
	for _, f := range acceptFiles {
		volPath := filepath.Join(root, f.Name())
		vol := ParseBookVolumeBasic(volPath)
		book.Volumes = append(book.Volumes, vol)
	}

	for i := range notAccept {
		entry := notAccept[i]
		if entry.IsDir() {
			path1 := filepath.Join(root, entry.Name())
			if err := walkBookExtraDir(path1, &book, options); err != nil {
				return err
			}
		}
	}

	if fn != nil {
		fn(&book)
	}
	return nil
}

func walkBookExtraDir(root string, book *BookMeta, options *LibraryOptions) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	acceptFiles, notAccept := options.filterAcceptFiles(root, entries)
	for i := range acceptFiles {
		// add to book extra
		volPath := filepath.Join(root, acceptFiles[i].Name())
		vol := ParseBookVolumeBasic(volPath)
		book.Extras = append(book.Extras, vol)
	}

	for i := range notAccept {
		entry := notAccept[i]
		if !entry.IsDir() {
			continue
		}
		path1 := filepath.Join(root, entry.Name())
		if err := walkBookExtraDir(path1, book, options); err != nil {
			return err
		}
	}

	return nil
}
