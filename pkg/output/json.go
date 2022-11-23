package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"registry-cli/pkg/errors"
)

type JSONArrayWriter struct {
	count  int
	stdout io.Writer
}

func (w *JSONArrayWriter) Write(obj interface{}) error {
	const intent = "    "
	if w.count == 0 {
		if _, err := fmt.Fprintln(w.stdout); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w.stdout, ",\n"); err != nil {
			return err
		}
	}
	t, err := json.MarshalIndent(obj, intent, intent)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w.stdout, "%s%s", intent, string(t)); err != nil {
		return err
	}
	w.count++
	return nil
}

func (w *JSONArrayWriter) Finish() error {
	if w.count > 0 {
		if _, err := fmt.Fprintln(w.stdout); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w.stdout, "]")
	return err
}

func NewJSONArrayWriter(stdout io.Writer) (*JSONArrayWriter, error) {
	w := &JSONArrayWriter{
		stdout: stdout,
	}
	if _, err := fmt.Fprintf(w.stdout, "["); err != nil {
		return nil, err
	}
	return w, nil
}

func WriteJSON(stdout io.Writer, obj interface{}) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stdout == nil {
		return errors.ErrNeedStdOut
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(obj); err != nil {
		return err
	}
	return nil
}
