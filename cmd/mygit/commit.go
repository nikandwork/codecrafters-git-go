package main

import (
	"fmt"

	git "git.codecrafters.io/0c40c1d7ba1ab4a0"
)

func commitTreeCmd(g *git.Git, args []string) error {
	if len(args) < 2 {
		commitUsage()

		return ErrUsage
	}

	var parent, message string

	tree := args[1]
	name := "nik"
	email := "nik@gmail.com"

	for i := 2; i < len(args); {
		arg := args[i]

		if i+1 == len(args) {
			return flagError(arg)
		}

		switch arg {
		case "-m":
			message = args[i+1]
		case "-p":
			parent = args[i+1]
		case "--name":
			name = args[i+1]
		case "--email":
			email = args[i+1]
		default:
			return fmt.Errorf("provided but not defined flag: %v", arg)
		}

		i += 2
	}

	if message == "" {
		return fmt.Errorf("message is required")
	}

	key, err := g.CommitTree(tree, parent, name, email, message, true)
	if err != nil {
		return err
	}

	fmt.Printf("%x\n", key[:])

	return nil
}

func flagError(flag string) error {
	commitUsage()

	return fmt.Errorf("flag %v requires value", flag)
}

func commitUsage() {
	fmt.Printf("usage: git commit-tree <tree_hash> [-p <parent_commit_hash>] -m <message>\n\n")
	fmt.Printf("    -p          parent commit hash\n")
	fmt.Printf("    -m          commit commit message\n")
	fmt.Printf("    --name      author name (nik)\n")
	fmt.Printf("    --email     author name (nik@gmail.com)\n")
}
