package main

import (
	"bufio"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func lsTreeCmd(args []string) (err error) {
	if len(args) != 3 || args[1] != "--name-only" {
		fmt.Printf("usage: git ls-tree --name-only <hash>\n\n")
		fmt.Printf("    --name-only	 list only filenames\n")

		return ErrUsage
	}

	obj := args[2]

	if len(obj) != 40 {
		return fmt.Errorf("bad object hash")
	}

	fname := filepath.Join(".git", "objects", obj[:2], obj[2:])

	f, err := os.Open(fname)
	if err != nil {
		return err // os package wraps error for us
	}

	defer func() {
		e := f.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close file: %w", e)
		}
	}()

	zr, err := zlib.NewReader(f)
	if err != nil {
		return fmt.Errorf("new zlib reader: %w", err)
	}

	defer func() {
		e := zr.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close zlib reader: %w", e)
		}
	}()

	bufr := bufio.NewReader(zr)

	return lsTreeObject(bufr)
}

func lsTreeObject(bufr *bufio.Reader) error {
	typ, _, err := parseObjectHeader(bufr)
	if err != nil {
		return fmt.Errorf("parse object header: %w", err)
	}

	if typ != "tree" {
		return fmt.Errorf("unsupported object type: %v", typ)
	}

	for {
		_, err := bufr.ReadString(' ')
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		name, err := bufr.ReadString('\000')
		if err != nil {
			return err
		}

		name = name[:len(name)-1]

		_, err = bufr.Discard(20)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", name)
	}

	return nil
}
