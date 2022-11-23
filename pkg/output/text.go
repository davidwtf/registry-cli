package output

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/liggitt/tabwriter"
)

type TextWriter struct {
	writer *tabwriter.Writer
	lock   sync.Mutex
}

func (w *TextWriter) Write(col ...string) error {
	w.lock.Lock()
	defer w.lock.Unlock()
	_, err := fmt.Fprintln(w.writer, strings.Join(col, "\t"))
	if err != nil {
		return err
	}
	return nil
}

func (w *TextWriter) Flush() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.writer.Flush()
}

func NewTextWriter(stdout io.Writer, headers ...string) (*TextWriter, error) {
	w := &TextWriter{
		writer: tabwriter.NewWriter(stdout, 6, 4, 3, ' ', tabwriter.RememberWidths),
	}
	if err := w.Write(headers...); err != nil {
		return nil, err
	}
	return w, nil
}
