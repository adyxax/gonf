package gonf

import (
	"bytes"
	"crypto/sha256"
	"io"
	"log/slog"
	"net/url"
	"os"
	"text/template"
)

// ----- Globals ---------------------------------------------------------------
var files []*FilePromise

// ----- Init ------------------------------------------------------------------
func init() {
	files = make([]*FilePromise, 0)
}

// ----- Public ----------------------------------------------------------------
func File(filename string, contents []byte) *FilePromise {
	return &FilePromise{
		chain:             nil,
		contents:          contents,
		err:               nil,
		filename:          filename,
		status:            PROMISED,
		templateFunctions: nil,
		useTemplate:       false,
	}
}

type FilePromise struct {
	chain             []Promise
	contents          []byte
	err               error
	filename          string
	status            Status
	templateFunctions map[string]any
	useTemplate       bool
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
	if f.useTemplate {
		tpl := template.New(f.filename)
		tpl.Option("missingkey=error")
		tpl.Funcs(builtinTemplateFunctions)
		tpl.Funcs(f.templateFunctions)
		if ttpl, err := tpl.Parse(string(f.contents)); err != nil {
			f.status = BROKEN
			slog.Error("file", "filename", f.filename, "status", f.status, "error", f.err)
			return
		} else {
			var buff bytes.Buffer
			if err := ttpl.Execute(&buff, 0 /* TODO */); err != nil {
				f.status = BROKEN
				slog.Error("file", "filename", f.filename, "status", f.status, "error", f.err)
				return
			}
			f.contents = buff.Bytes()
		}
	}
	var sumFile []byte
	sumFile, f.err = sha256sumOfFile(f.filename)
	if f.err != nil {
		f.status = BROKEN
		return
	}
	sumContents := sha256sum(f.contents)
	if !bytes.Equal(sumFile, sumContents) {
		if f.err = writeFile(f.filename, f.contents); f.err != nil {
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

func Template(filename string, contents []byte) *FilePromise {
	f := File(filename, contents)
	f.useTemplate = true
	return f
}

func TemplateWith(filename string, contents []byte, templateFunctions map[string]any) *FilePromise {
	f := Template(filename, contents)
	f.templateFunctions = templateFunctions
	return f
}

// ----- Internal --------------------------------------------------------------
var builtinTemplateFunctions = map[string]any{
	"encodeURIQueryParameter": url.QueryEscape,
	"var":                     getVariable,
}

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
