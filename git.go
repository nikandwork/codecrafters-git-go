package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type (
	Git struct {
		dataRoot string // project folder
		gitRoot  string // git folder
	}
)

var (
	ErrNotGit = errors.New("not in git repository")
	ErrBadGit = errors.New("invalid gitfile format")
)

func Init(root string) (*Git, error) {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		full := filepath.Join(root, dir)

		err := os.MkdirAll(full, 0755)
		if err != nil {
			return nil, err // os package wraps error for us, so we shouldn't
		}
	}

	headFileContents := []byte("ref: refs/heads/master\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return nil, fmt.Errorf("write %s: %w", ".git/HEAD", err)
	}

	return &Git{
		dataRoot: root,
		gitRoot:  filepath.Join(root, ".git"),
	}, nil
}

func Find(path string) (*Git, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	for p := path; ; p = filepath.Dir(p) {
		err := checkGitDir(p)
		if os.IsNotExist(err) {
			if p == "/" {
				break
			}

			continue
		}
		if err != nil {
			return nil, err
		}

		return &Git{
			dataRoot: p,
			gitRoot:  filepath.Join(p, ".git"),
		}, nil
	}

	return nil, ErrNotGit
}

func checkGitDir(path string) error {
	gitPath := filepath.Join(path, ".git")

	inf, err := os.Stat(gitPath)
	if os.IsNotExist(err) {
		return err
	}

	if !inf.IsDir() {
		return ErrBadGit
	}

	ok := checkDirExists(filepath.Join(gitPath, "objects"))
	if !ok {
		return fmt.Errorf("%w: no %v folder", ErrBadGit, "objects")
	}

	ok = checkDirExists(filepath.Join(gitPath, "refs"))
	if !ok {
		return fmt.Errorf("%w: no %v folder", ErrBadGit, "refs")
	}

	return nil
}

func checkDirExists(path string) bool {
	inf, err := os.Stat(path)

	return err == nil && inf.IsDir()
}
