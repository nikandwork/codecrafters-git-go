package git

import (
	"fmt"
	"time"

	"git.codecrafters.io/0c40c1d7ba1ab4a0/object"
)

func (g *Git) CommitTree(tree, parent, name, email, message string, write bool) (object.Hash, error) {
	var content []byte

	content = fmt.Appendf(content, "tree %s\n", tree)

	if parent != "" {
		content = fmt.Appendf(content, "parent %s\n", parent)
	}

	now := time.Now()

	content = fmt.Appendf(content, "author %s <%s> %d", name, email, now.Unix())
	content = now.AppendFormat(content, " -0700\n")

	content = fmt.Appendf(content, "\n%s", message)
	if message != "" && message[len(message)-1] != '\n' {
		content = append(content, '\n')
	}

	return g.hashObject("commit", content, write)
}
