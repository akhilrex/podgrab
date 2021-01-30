// Copyright 2013 Michael Yang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package id3

import (
	"errors"
	"os"

	v1 "github.com/akhilrex/podgrab/internal/id3/v1"
	v2 "github.com/akhilrex/podgrab/internal/id3/v2"
)

const (
	LatestVersion = 3
)

// Tagger represents the metadata of a tag
type Tagger interface {
	Title() string
	Artist() string
	Album() string
	Year() string
	Genre() string
	Comments() []string
	SetTitle(string)
	SetArtist(string)
	SetAlbum(string)
	SetYear(string)
	SetComment(string)
	SetGenre(string)
	SetDate(string)        // Added
	SetReleaseYear(string) // Added
	AllFrames() []v2.Framer
	Frames(string) []v2.Framer
	Frame(string) v2.Framer
	DeleteFrames(string) []v2.Framer
	AddFrames(...v2.Framer)
	Bytes() []byte
	Dirty() bool
	Padding() uint
	Size() int
	Version() string
}

// File represents the tagged file
type File struct {
	Tagger
	originalSize int
	file         *os.File
}

// Parses an open file
func Parse(file *os.File, forcev2 bool) (*File, error) {
	res := &File{file: file}

	if forcev2 {
		res.Tagger = v2.NewTag(2)
	} else {
		if v2Tag := v2.ParseTag(file); v2Tag != nil {
			res.Tagger = v2Tag
			res.originalSize = v2Tag.Size()
		} else if v1Tag := v1.ParseTag(file); v1Tag != nil {
			res.Tagger = v1Tag
		} else {
			// Add a new tag if none exists
			res.Tagger = v2.NewTag(LatestVersion)
		}
	}

	return res, nil
}

// Opens a new tagged file
func Open(name string, forceV2 bool) (*File, error) {
	fi, err := os.OpenFile(name, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	file, err := Parse(fi, forceV2)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Saves any edits to the tagged file
func (f *File) Close() error {
	defer f.file.Close()

	if !f.Dirty() {
		return nil
	}

	switch f.Tagger.(type) {
	case (*v1.Tag):
		if _, err := f.file.Seek(-v1.TagSize, os.SEEK_END); err != nil {
			return err
		}
	case (*v2.Tag):
		if f.Size() > f.originalSize {
			start := int64(f.originalSize + v2.HeaderSize)
			offset := int64(f.Tagger.Size() - f.originalSize)

			if err := shiftBytesBack(f.file, start, offset); err != nil {
				return err
			}
		}

		if _, err := f.file.Seek(0, os.SEEK_SET); err != nil {
			return err
		}
	default:
		return errors.New("Close: unknown tag version")
	}

	if _, err := f.file.Write(f.Tagger.Bytes()); err != nil {
		return err
	}

	return nil
}
