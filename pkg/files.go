package gonf

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

var files []*FilePromise

func init() {
	files = make([]*FilePromise, 0)
}

type FilePromise struct {
	chain          []Promise
	contents       Value
	dirPermissions *Permissions
	err            error
	filename       Value
	permissions    *Permissions
	status         Status
}

func File(filename any) *FilePromise {
	return &FilePromise{
		chain:          nil,
		contents:       nil,
		dirPermissions: nil,
		err:            nil,
		filename:       interfaceToTemplateValue(filename),
		permissions:    nil,
		status:         PROMISED,
	}
}

func (f *FilePromise) Contents(contents any) *FilePromise {
	f.contents = interfaceToValue(contents)
	return f
}

func (f *FilePromise) DirectoriesPermissions(p *Permissions) *FilePromise {
	f.dirPermissions = p
	return f
}

func (f *FilePromise) Permissions(p *Permissions) *FilePromise {
	f.permissions = p
	return f
}

func (f *FilePromise) Template(contents any) *FilePromise {
	f.contents = interfaceToTemplateValue(contents)
	return f
}

func (f *FilePromise) IfRepaired(p ...Promise) Promise {
	f.chain = append(f.chain, p...)
	return f
}

func (f *FilePromise) Promise() Promise {
	files = append(files, f)
	return f
}

func (f *FilePromise) Resolve() {
	filename := f.filename.String()
	if f.dirPermissions != nil {
		if f.status, f.err = makeDirectoriesHierarchy(filepath.Dir(filename), f.dirPermissions); f.err != nil {
			return
		}
	}
	if f.contents != nil {
		var sumFile []byte
		sumFile, f.err = sha256sumOfFile(filename)
		if f.err != nil {
			if !errors.Is(f.err, fs.ErrNotExist) {
				slog.Error("file", "filename", f.filename, "status", f.status, "error", f.err)
				f.status = BROKEN
				return
			}
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
		}
	}
	if f.permissions != nil {
		var status Status
		status, f.err = f.permissions.resolve(filename)
		if f.status == PROMISED || status == BROKEN {
			f.status = status
		}
		if f.err != nil {
			slog.Error("file", "filename", f.filename, "status", f.status, "error", f.err)
			return
		}
	}
	if f.status == REPAIRED {
		slog.Info("file", "filename", f.filename, "status", f.status)
		for _, p := range f.chain {
			p.Resolve()
		}
	} else {
		f.status = KEPT
		slog.Debug("file", "filename", f.filename, "status", f.status)
	}
}

func (f FilePromise) Status() Status {
	return f.status
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

func sha256sumOfFile(filename string) (hash []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := f.Close(); err == nil {
			err = e
		}
	}()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func writeFile(filename string, contents []byte) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if e := f.Close(); err == nil {
			err = e
		}
	}()
	_, err = f.Write(contents)
	return err
}
