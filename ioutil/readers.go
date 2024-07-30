package ioutil

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

// Reads the content of the reader as a jsonl file.
func ReadJSONL[T ~[]E, E any](dest T, r io.Reader) (_ T, err error) {
	bufR := bufio.NewReader(r)
	for {
		buf, err := bufR.ReadBytes('\n')
		if err != nil {
			break
		}

		var e E
		if err = json.Unmarshal(buf, &e); err != nil {
			break
		}

		dest = append(dest, e)
	}

	return dest, err
}

// ReadFrom opens the file for reading and calls the parsing function. If the file is the special
// "-" name, ReadFrom reads from standard input.
func ReadFrom[T ~[]E, E any](file string, parse func(r io.Reader) (T, error)) (T, error) {
	if file == "-" {
		return parse(os.Stdin)
	}

	f, err := os.Open(file)
	if err != nil {
		var t T
		return t, err
	}
	defer f.Close()

	return parse(f)
}
