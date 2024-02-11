package gonf

import (
	"bytes"
	"crypto/sha256"
	"io"
	"log/slog"
	"os"
)

// ----- Globals ---------------------------------------------------------------
var files []*FilePromise

// ----- Init ------------------------------------------------------------------
func init() {
	files = make([]*FilePromise, 0)
}

// ----- Public ----------------------------------------------------------------
type FilePromise struct {
	chain    []Promise
	contents Value
	err      error
	filename Value
	status   Status
}

func File(filename any, contents string) *FilePromise {
	return &FilePromise{
		chain:    nil,
		contents: interfaceToValue(contents),
		err:      nil,
		filename: interfaceToTemplateValue(filename),
		status:   PROMISED,
	}
}

func Template(filename any, contents any) *FilePromise {
	return &FilePromise{
		chain:    nil,
		contents: interfaceToTemplateValue(contents),
		err:      nil,
		filename: interfaceToTemplateValue(filename),
		status:   PROMISED,
	}
}

func (f *FilePromise) IfRepaired(p ...Promise) Promise {
	f.chain = p
	return f
}

func (f *FilePromise) Promise() Promise {
	files = append(files, f)
	return f
}

func (f *FilePromise) Resolve() {
	filename := f.filename.String()
	var sumFile []byte
	sumFile, f.err = sha256sumOfFile(filename)
	if f.err != nil {
		f.status = BROKEN
		return
	}
	contents := f.contents.Bytes()
	sumContents := sha256sum(contents)
	if !bytes.Equal(sumFile, sumContents) {
		if f.err = writeFile(filename, contents); f.err != nil {
			f.status = BROKEN
			slog.Error("file", "filename", f.filename, "status", f.status, "error", f.err)
			return
		}
		f.status = REPAIRED
		slog.Info("file", "filename", f.filename, "status", f.status)
		for _, p := range f.chain {
			p.Resolve()
		}
		return
	}
	f.status = KEPT
	slog.Debug("file", "filename", f.filename, "status", f.status)
}

// ----- Internal --------------------------------------------------------------
func resolveFiles() (status Status) {
	status = KEPT
	for _, f := range files {
		if f.status == PROMISED {
			f.Resolve()
			switch f.status {
			case BROKEN:
				return BROKEN
			case REPAIRED:
				status = REPAIRED
			}
		}
	}
	return
}

func sha256sumOfFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func writeFile(filename string, contents []byte) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(contents)
	return err
}
