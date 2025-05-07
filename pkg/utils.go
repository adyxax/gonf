package gonf

import (
	"crypto/sha256"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

var builtinTemplateFunctions = map[string]any{
	//"encodeURIQueryParameter": url.QueryEscape,
	"fact": isFact,
	"var":  getVariable,
}

func FilterSlice[T any](slice *[]T, predicate func(T) bool) {
	i := 0
	for _, element := range *slice {
		if predicate(element) { // if the element matches the predicate function
			(*slice)[i] = element // then we keep it in the slice
			i++
		} // otherwise the element will get overwritten
	}
	*slice = (*slice)[:i] // or truncated out of the slice
}

// We cannot just use os.MakedirAll because we need to set the user:group on every intermediate directories created
func makeDirectoriesHierarchy(dir string, perms *Permissions) (Status, error) {
	if _, err := os.Lstat(dir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if _, err := makeDirectoriesHierarchy(filepath.Dir(dir), perms); err != nil {
				return BROKEN, err
			}
			m, err := perms.mode.Int()
			if err != nil {
				return BROKEN, err
			}
			if err := os.Mkdir(dir, fs.FileMode(m)); err != nil {
				return BROKEN, err
			}
			if _, err := perms.resolve(dir); err != nil {
				return BROKEN, err
			}
			return REPAIRED, nil
		} else {
			return BROKEN, err
		}
	} else {
		return KEPT, nil
	}
}

func sha256sum(contents []byte) []byte {
	h := sha256.New()
	h.Write(contents)
	return h.Sum(nil)
}
