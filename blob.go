package git

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"git.codecrafters.io/0c40c1d7ba1ab4a0/object"
)

var ErrShortHash = errors.New("too short hash")

func (g *Git) CatFile(obj string) error {
	path, err := g.findPath(obj)
	if err != nil {
		return fmt.Errorf("find object: %w", err)
	}

	typ, content, err := object.ReadFromFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	switch typ {
	case "blob":
		_, err = os.Stdout.Write(content)
	case "tree":
		err = g.lsTree(content, false)
	default:
		return fmt.Errorf("unsupported object type: %v", typ)
	}

	if err != nil {
		return fmt.Errorf("print: %w", err)
	}

	return nil
}

func (g *Git) HashObject(file string, write bool) (object.Hash, error) {
	return g.writeObject("blob", file, write)
}

func (g *Git) writeObject(typ, file string, write bool) (object.Hash, error) {
	var buf bytes.Buffer

	key, err := object.WriteFile(&buf, typ, file)
	if err != nil {
		return object.Hash{}, fmt.Errorf("encode object: %w", err)
	}

	if !write {
		return key, nil
	}

	path := g.objectPath(key)
	dir := filepath.Dir(path)

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return object.Hash{}, fmt.Errorf("create object dir: %w", err)
	}

	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		return object.Hash{}, fmt.Errorf("write file: %w", err)
	}

	return key, nil
}

func (g *Git) findPath(obj string) (string, error) {
	// TODO: deciding of branch names, tags and hash prefixes should be here

	if len(obj) < 4 {
		return "", errors.New("short object-hash")
	}

	return filepath.Join(g.gitRoot, "objects", obj[:2], obj[2:]), nil
}

func (g *Git) objectPath(key object.Hash) string {
	hashStr := hex.EncodeToString(key[:])

	return filepath.Join(g.gitRoot, "objects", hashStr[:2], hashStr[2:])
}
