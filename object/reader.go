package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

type (
	Hash = [sha1.Size]byte
)

func ReadFromFile(path string) (typ string, content []byte, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, fmt.Errorf("open: %w", err)
	}

	defer func() {
		e := f.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close: %w", e)
		}
	}()

	return Read(f)
}

func Read(src io.Reader) (typ string, content []byte, err error) {
	zr, err := zlib.NewReader(src)
	if err != nil {
		return "", nil, fmt.Errorf("new zlib reader: %w", err)
	}

	defer func() {
		e := zr.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close zlib reader: %w", e)
		}
	}()

	content, err = io.ReadAll(zr)
	if err != nil {
		return "", nil, err
	}

	typeEnd := bytes.IndexByte(content, ' ')
	if typeEnd == -1 {
		return "", nil, errors.New("no object type delimiter")
	}

	typ = string(content[:typeEnd])

	typeEnd++

	sizeEnd := typeEnd + bytes.IndexByte(content[typeEnd:], '\000')
	if sizeEnd < typeEnd {
		return "", nil, errors.New("no object size delimiter")
	}

	size, err := strconv.ParseInt(string(content[typeEnd:sizeEnd]), 10, 64)
	if err != nil {
		return "", nil, fmt.Errorf("parse size: %w", err)
	}

	sizeEnd++

	if l := len(content[sizeEnd:]); l != int(size) {
		return "", nil, fmt.Errorf("object size mismatch: expected %d, read %d", size, l)
	}

	return typ, content[sizeEnd:], nil
}
