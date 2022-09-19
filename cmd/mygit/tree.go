package main

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Hash = [sha1.Size]byte

const (
	ModeTree = 1 << 14
	ModeBlob = 1 << 15
)

func lsTreeCmd(args []string) (err error) {
	if len(args) != 2 && (args[1] != "--name-only" || len(args) != 3) {
		fmt.Printf("usage: git ls-tree [--name-only] <hash>\n\n")
		fmt.Printf("    --name-only	 list only filenames\n")

		return ErrUsage
	}

	var nameOnly bool
	var obj string

	if args[1] == "--name-only" {
		nameOnly = true
		obj = args[2]
	} else {
		obj = args[1]
	}

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

	return prettyPrintObject(bufr, nameOnly)
}

func lsTreeObject(bufr *bufio.Reader, nameOnly bool) error {
	sum := make([]byte, 20)

	for {
		modeStr, err := bufr.ReadString(' ')
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		modeStr = modeStr[:len(modeStr)-1]

		name, err := bufr.ReadString('\000')
		if err != nil {
			return err
		}

		name = name[:len(name)-1]

		_, err = bufr.Read(sum)
		if err != nil {
			return err
		}

		if nameOnly {
			fmt.Printf("%s\n", name)
		} else {
			mode, err := strconv.ParseInt(modeStr, 8, 64)
			if err != nil {
				return fmt.Errorf("parse mode")
			}

			typ := "????"

			switch {
			case mode&ModeBlob != 0:
				typ = "blob"
			case mode&ModeTree != 0:
				typ = "tree"
			}

			fmt.Printf("%06o %s %x    %-15s\n", mode, typ, sum, name)
		}
	}

	return nil
}

func writeTreeCmd(args []string) error {
	sum, err := writeTree(".")
	if err != nil {
		return err
	}

	fmt.Printf("%x\n", sum[:])

	return nil
}

func writeTree(root string) (Hash, error) {
	files, err := os.ReadDir(root)
	if err != nil {
		return Hash{}, fmt.Errorf("read dir: %w", err)
	}

	var table []byte

	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}

		filePath := filepath.Join(root, f.Name())

		inf, err := f.Info()
		if err != nil {
			return Hash{}, fmt.Errorf("file info: %v: %w", filePath, err)
		}

		mode := inf.Mode()

		if !mode.IsDir() && !mode.IsRegular() {
			continue
		}

		var gitMode int
		var sum Hash

		if mode.IsDir() {
			gitMode |= ModeTree

			sum, err = writeTree(filePath)
			if err != nil {
				return Hash{}, err
			}
		} else if mode.IsRegular() {
			gitMode = ModeBlob | int(mode)&0xfff

			sum, err = hashObjectFile("blob", filePath, true)
			if err != nil {
				return Hash{}, fmt.Errorf("hash object %v: %w", filePath, err)
			}
		}

		table = fmt.Appendf(table, "%06o %v\x00%s", gitMode, f.Name(), sum[:])
	}

	sum, err := hashObject("tree", table, true)
	if err != nil {
		return Hash{}, fmt.Errorf("hash tree %v: %w", root, err)
	}

	return sum, nil
}
