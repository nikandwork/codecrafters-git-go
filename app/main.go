package main

import (
	"bufio"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

var ErrUsage = errors.New("usage error")

func main() {
	if len(os.Args) == 1 {
		help()
		os.Exit(1)
	}

	var err error

	switch command := os.Args[1]; command {
	case "help":
		help()
	case "init":
		err = initCmd()
	case "cat-file":
		err = catfileCmd(os.Args[1:])
	default:
		err = fmt.Errorf("%q is not a git command. See git --help", command)
	}

	if err != nil {
		if !errors.Is(err, ErrUsage) {
			fmt.Fprintf(os.Stderr, "git: %v\n", err)
		}

		os.Exit(1)
	}
}

func help() {
	fmt.Printf(`usage: git [<flags>] <command> [<args>]

There are commands:

start a working area
`)

	fmt.Printf("   init     Create an empty Git repository\n")
}

func initCmd() error {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err // os package wraps error for us, so we shouldn't
		}
	}

	headFileContents := []byte("ref: refs/heads/master\n")
	// ioutil package is deprecated
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("write %s: %w", ".git/HEAD", err)
	}

	fmt.Println("Initialized git directory")

	return nil
}

func catfileCmd(args []string) (err error) {
	if len(args) < 3 || args[1] != "-p" {
		fmt.Printf("usage: git cat-file -p <object>\n\n")
		fmt.Printf("    -p			pretty print object's content\n")

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

	return prettyPrintObject(bufr)
}

func prettyPrintObject(bufr *bufio.Reader) error {
	typ, size, err := parseObjectHeader(bufr)
	if err != nil {
		return fmt.Errorf("parse object header: %w", err)
	}

	if typ != "blob" {
		return fmt.Errorf("unsupported object type")
	}

	_, err = io.CopyN(os.Stdout, bufr, size)
	if err != nil {
		return fmt.Errorf("read content: %w", err)
	}

	return nil
}

func parseObjectHeader(br *bufio.Reader) (string, int64, error) {
	typ, err := br.ReadString(' ')
	if err != nil {
		return "", 0, fmt.Errorf("read type: %w", err)
	}

	typ = typ[:len(typ)-1] // cut ' '

	siz, err := br.ReadString('\000')
	if err != nil {
		return "", 0, fmt.Errorf("read size: %w", err)
	}

	siz = siz[:len(siz)-1] // cut '\000'

	size, err := strconv.ParseInt(siz, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("parse size: %w", err)
	}

	return typ, size, nil
}
