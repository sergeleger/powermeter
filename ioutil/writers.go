package ioutil

import (
	"bufio"
	"encoding/json"
	"io"
)

// WriteJSONL writes the contents of T as a jsonl file.
func WriteJSONL[T ~[]E, E any](w io.Writer, source T) error {
	bufW := bufio.NewWriter(w)
	for _, e := range source {
		buf, err := json.Marshal(e)
		if err != nil {
			return err
		}

		bufW.Write(buf)
		bufW.WriteByte('\n')
	}

	return bufW.Flush()
}
