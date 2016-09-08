package statictemplate

import (
	"io"
)

type constantWriterTo string

func (c constantWriterTo) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(c))
	return int64(n), err
}
