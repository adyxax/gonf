package gonf

import (
	"crypto/sha256"
)

var builtinTemplateFunctions = map[string]any{
	//"encodeURIQueryParameter": url.QueryEscape,
	"var": getVariable,
}

func sha256sum(contents []byte) []byte {
	h := sha256.New()
	h.Write(contents)
	return h.Sum(nil)
}
