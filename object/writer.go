package object

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func WriteFile(dst io.Writer, typ, file string) (key Hash, err error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return Hash{}, fmt.Errorf("read data file: %w", err)
	}

	return Write(dst, typ, content)
}

func Write(dst io.Writer, typ string, content []byte) (key Hash, err error) {
	zw := zlib.NewWriter(dst)
	defer func() {
		e := zw.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close zlib writer: %w", e)
		}
	}()

	hash := sha1.New()

	w := io.MultiWriter(hash, zw)

	_, err = fmt.Fprintf(w, "%s %d\000", typ, len(content))
	if err != nil {
		return Hash{}, err
	}

	_, err = w.Write(content)
	if err != nil {
		return Hash{}, err
	}

	_ = hash.Sum(key[:0])

	return key, nil
}
