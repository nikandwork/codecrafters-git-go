package main

import (
	"fmt"

	git "git.codecrafters.io/0c40c1d7ba1ab4a0"
)

func lsTreeCmd(g *git.Git, args []string) (err error) {
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

	return g.LsTree(obj, nameOnly)
}

func writeTreeCmd(g *git.Git, args []string) error {
	sum, err := g.WriteTree(".", true)
	if err != nil {
		return err
	}

	fmt.Printf("%x\n", sum[:])

	return nil
}
