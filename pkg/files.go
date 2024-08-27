package gonf

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
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

type FileType int

const (
	FILE = iota
	DIRECTORY
)

type FilePromise struct {
	chain          []Promise
	contents       Value
	dirPermissions *Permissions
	err            error
	filename       Value
	fileType       FileType
	permissions    *Permissions
	status         Status
}

func Directory(filename any) *FilePromise {
	return &FilePromise{
		chain:          nil,
		contents:       nil,
		dirPermissions: nil,
		err:            nil,
		filename:       interfaceToTemplateValue(filename),
		fileType:       DIRECTORY,
		permissions:    nil,
		status:         PROMISED,
	}
}

func File(filename any) *FilePromise {
	return &FilePromise{
		chain:          nil,
		contents:       nil,
		dirPermissions: nil,
		err:            nil,
		filename:       interfaceToTemplateValue(filename),
		fileType:       FILE,
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

func (f *FilePromise) Promise() *FilePromise {
	files = append(files, f)
	return f
}

func (f *FilePromise) Resolve() {
	filename := f.filename.String()
	if f.status, f.err = makeDirectoriesHierarchy(filepath.Dir(filename), f.dirPermissions); f.err != nil {
		return
	}
	switch f.fileType {
	case DIRECTORY:
		if f.status, f.err = makeDirectoriesHierarchy(filepath.Clean(filename), f.dirPermissions); f.err != nil {
			return
		}
	case FILE:
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
	default:
		panic(fmt.Errorf("unknown File type enum value %d", f.fileType))
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
