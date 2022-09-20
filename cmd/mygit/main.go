package main

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"

	git "git.codecrafters.io/0c40c1d7ba1ab4a0"
)

var ErrUsage = errors.New("usage error")

func main() {
	if len(os.Args) == 1 {
		help()
		os.Exit(1)
	}

	err := run(os.Args)
	if err != nil {
		if !errors.Is(err, ErrUsage) {
			fmt.Fprintf(os.Stderr, "git: %v\n", err)
		}

		os.Exit(1)
	}
}

func run(args []string) error {
	var err error
	command := args[1]

	switch command {
	case "help":
		help()

		return nil
	case "init":
		_, err = git.Init(".")

		return err

	case "zlib":
		err = zlibCmd(os.Args[1:])

		return err
	}

	g, err := git.Find(".")
	if err != nil {
		return err
	}

	switch command {
	case "cat-file":
		err = catFileCmd(g, os.Args[1:])
	case "hash-object":
		err = hashObjectCmd(g, os.Args[1:])
	case "ls-tree":
		err = lsTreeCmd(g, os.Args[1:])
	case "write-tree":
		err = writeTreeCmd(g, os.Args[1:])
	case "commit-tree":
		err = commitTreeCmd(g, os.Args[1:])
	default:
		err = fmt.Errorf("%q is not a git command. See git --help", command)
	}

	if err != nil {
		return fmt.Errorf("%v: %w", command, err)
	}

	return nil
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
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("write %s: %w", ".git/HEAD", err)
	}

	fmt.Println("Initialized git directory")

	return nil
}

func zlibCmd(args []string) (err error) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: git zlib -d [<file>]\n\nread from stdin if file not specified\n")

		return ErrUsage
	}

	switch f := args[1]; f {
	case "-d":
		// fine
	default:
		return fmt.Errorf("unsupported flag: %v", f)
	}

	obj := "-"

	if len(args) > 2 {
		obj = args[2]
	}

	var data []byte

	if obj == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(obj)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
	}

	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("new zlib reader: %w", err)
	}

	defer func() {
		e := r.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close zlib reader: %w", e)
		}
	}()

	decoded, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	_, err = os.Stdout.Write(decoded)
	if err != nil {
		return fmt.Errorf("write to output: %w", err)
	}

	return nil
}
