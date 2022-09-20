package main

import (
	"fmt"

	git "git.codecrafters.io/0c40c1d7ba1ab4a0"
)

func catFileCmd(g *git.Git, args []string) (err error) {
	if len(args) != 3 || args[1] != "-p" {
		fmt.Printf("usage: git cat-file -p <object>\n\n")
		fmt.Printf("    -p			pretty print object's content\n")

		return ErrUsage
	}

	obj := args[2]

	return g.CatFile(obj)
}

func hashObjectCmd(g *git.Git, args []string) error {
	if len(args) < 2 || args[1] == "-w" && len(args) < 3 {
		fmt.Printf("usage: git hash-object [-w] <file>\n\n")
		fmt.Printf("    -w			write the onject into the object database\n")

		return ErrUsage
	}

	var doWrite bool

	if args[1] == "-w" {
		doWrite = true
		args = args[2:] // args now is a list of files to hash
	} else {
		args = args[1:]
	}

	for _, file := range args {
		sum, err := g.HashObject(file, doWrite)
		if err != nil {
			return fmt.Errorf("%v: %w", file, err)
		}

		fmt.Printf("%x\n", sum)
	}

	return nil
}
