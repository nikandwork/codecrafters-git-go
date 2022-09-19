package main

import (
	"errors"
	"fmt"
	"os"
)

var ErrUsage = errors.New("usage error")

func main() {
	if len(os.Args) == 1 {
		help()
		os.Exit(1)
	}

	var err error
	command := os.Args[1]

	switch command {
	case "help":
		help()
	case "init":
		err = initCmd()
	case "cat-file":
		err = catFileCmd(os.Args[1:])
	case "hash-object":
		err = hashObjectCmd(os.Args[1:])
	case "ls-tree":
		err = lsTreeCmd(os.Args[1:])
	default:
		err = fmt.Errorf("%q is not a git command. See git --help", command)
	}

	if err != nil {
		if !errors.Is(err, ErrUsage) {
			fmt.Fprintf(os.Stderr, "git: %v: %v\n", command, err)
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
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("write %s: %w", ".git/HEAD", err)
	}

	fmt.Println("Initialized git directory")

	return nil
}
