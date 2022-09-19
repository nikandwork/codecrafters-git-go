package main

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func catFileCmd(args []string) (err error) {
	if len(args) != 3 || args[1] != "-p" {
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

	return prettyPrintObject(bufr, false)
}

func prettyPrintObject(bufr *bufio.Reader, nameOnly bool) error {
	typ, size, err := parseObjectHeader(bufr)
	if err != nil {
		return fmt.Errorf("parse object header: %w", err)
	}

	if typ == "tree" {
		return lsTreeObject(bufr, nameOnly)
	}

	if typ != "blob" {
		return fmt.Errorf("unsupported object type: %v", typ)
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

	sizeStr, err := br.ReadString('\000')
	if err != nil {
		return "", 0, fmt.Errorf("read size: %w", err)
	}

	sizeStr = sizeStr[:len(sizeStr)-1] // cut '\000'

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("parse size: %w", err) // ParseInt already includes the sizeStr in error
	}

	return typ, size, nil
}

func hashObjectCmd(args []string) error {
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
		sum, err := hashObjectFile("blob", file, doWrite)
		if err != nil {
			return fmt.Errorf("%v: %w", file, err)
		}

		fmt.Printf("%s\n", sum)
	}

	return nil
}

func hashObjectFile(typ, file string, doWrite bool) (Hash, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Hash{}, fmt.Errorf("read file: %w", err)
	}

	return hashObject(typ, data, doWrite)
}

func hashObject(typ string, data []byte, doWrite bool) (sum Hash, err error) {
	object := make([]byte, 0, len(typ)+1+21+len(data)) // typ + ' ' + <size> + '\000' + data

	object = append(object, typ...)
	object = append(object, ' ')

	object = strconv.AppendInt(object, int64(len(data)), 10)
	object = append(object, '\000')

	object = append(object, data...)

	sum = sha1.Sum(object)

	if !doWrite {
		return sum, nil
	}

	sumStr := hex.EncodeToString(sum[:])

	fname := filepath.Join(".git", "objects", sumStr[:2], sumStr[2:])
	dir := filepath.Dir(fname)

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return sum, fmt.Errorf("create dir: %w", err)
	}

	f, err := os.Create(fname)
	if err != nil {
		return sum, fmt.Errorf("create file: %w", err)
	}

	defer func() {
		e := f.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close file: %w", e)
		}
	}()

	w := zlib.NewWriter(f)

	_, err = w.Write(object)
	if err != nil {
		return sum, fmt.Errorf("write file: %w", err)
	}

	err = w.Close()
	if err != nil {
		return sum, fmt.Errorf("close zlib writer: %w", err)
	}

	return sum, nil
}
