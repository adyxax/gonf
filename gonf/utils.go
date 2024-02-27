package gonf

import (
	"crypto/sha256"
)

var builtinTemplateFunctions = map[string]any{
	//"encodeURIQueryParameter": url.QueryEscape,
	"var": getVariable,
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

func sha256sum(contents []byte) []byte {
	h := sha256.New()
	h.Write(contents)
	return h.Sum(nil)
}
