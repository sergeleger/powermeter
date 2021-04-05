package main

import (
	"io"
	"os"
)

type ReaderFrom interface {
	ReadFrom(r io.Reader) error
}

type WriterTo interface {
	WriteTo(w io.Writer) error
}

func ReadFrom(dest ReaderFrom, file string) error {
	var r io.Reader = os.Stdin
	if file != "-" {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	return dest.ReadFrom(r)
}

func WriteTo(src WriterTo, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return src.WriteTo(f)
}
