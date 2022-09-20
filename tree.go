package git

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.codecrafters.io/0c40c1d7ba1ab4a0/object"
)

const (
	ModeTree = 1 << 14
	ModeBlob = 1 << 15
)

func (g *Git) LsTree(obj string, nameOnly bool) error {
	path, err := g.findPath(obj)
	if err != nil {
		return fmt.Errorf("find object: %w", err)
	}

	typ, content, err := object.ReadFromFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if typ != "tree" {
		return fmt.Errorf("not a tree object: %v", typ)
	}

	return g.lsTree(content, nameOnly)
}

func (g *Git) lsTree(content []byte, nameOnly bool) error {
	for i := 0; i < len(content); {
		space := i + bytes.IndexByte(content[i:], ' ')
		if space < i { // IndexByte == -1
			return errors.New("malformed tree")
		}

		mode := content[i:space]

		zero := space + bytes.IndexByte(content[space:], '\000')
		if zero < space { // IndexByte == -1
			return errors.New("malformed tree")
		}

		name := content[space+1 : zero]
		sum := content[zero+1 : zero+1+20]

		if nameOnly {
			fmt.Printf("%v\n", name)
		} else {
			modeInt, err := parseMode(string(mode))
			if err != nil {
				return fmt.Errorf("parse mode: %w", err)
			}

			typ := "????"

			switch {
			case modeInt&ModeTree != 0:
				typ = "tree"
			case modeInt&ModeBlob != 0:
				typ = "blob"
			}

			fmt.Printf("%06o %s %x    %s\n", modeInt, typ, sum, name)
		}

		i = zero + 1 + 20 // '\000' + 20 byte hash
	}

	return nil
}

func (g *Git) WriteTree(dir string, write bool) (object.Hash, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return object.Hash{}, fmt.Errorf("read dir: %w", err)
	}

	var table []byte

	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}

		filePath := filepath.Join(dir, f.Name())

		inf, err := f.Info()
		if err != nil {
			return object.Hash{}, fmt.Errorf("file info: %v: %w", filePath, err)
		}

		mode := inf.Mode()

		if !mode.IsDir() && !mode.IsRegular() {
			continue
		}

		var gitMode int
		var sum object.Hash

		if mode.IsDir() {
			gitMode |= ModeTree

			sum, err = g.WriteTree(filePath, write)
			if err != nil {
				return object.Hash{}, err
			}
		} else if mode.IsRegular() {
			gitMode = ModeBlob | int(mode)&0xfff

			sum, err = g.writeObject("blob", filePath, write)
			if err != nil {
				return object.Hash{}, fmt.Errorf("hash object %v: %w", filePath, err)
			}
		}

		table = fmt.Appendf(table, "%06o %v\x00%s", gitMode, f.Name(), sum[:])
	}

	return g.hashTree("tree", table, write)
}

func (g *Git) hashTree(typ string, content []byte, write bool) (object.Hash, error) {
	var buf bytes.Buffer

	key, err := object.Write(&buf, typ, content)
	if err != nil {
		return object.Hash{}, fmt.Errorf("encode object")
	}

	if !write {
		return key, nil
	}

	objPath := g.objectPath(key)
	dir := filepath.Dir(objPath)

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return object.Hash{}, fmt.Errorf("create object dir: %w", err)
	}

	err = os.WriteFile(objPath, buf.Bytes(), 0644)
	if err != nil {
		return object.Hash{}, fmt.Errorf("write file: %w", err)
	}

	return key, nil
}

func parseMode(s string) (int, error) {
	x, err := strconv.ParseInt(s, 8, 64)

	return int(x), err
}
