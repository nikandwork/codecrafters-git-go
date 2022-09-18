package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		help()
		os.Exit(1)
	}

	var err error

	switch command := os.Args[1]; command {
	case "init":
		err = initCmd()
	default:
		err = fmt.Errorf("%q is not a git command. See git --help", command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "git: %v\n", err)
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
