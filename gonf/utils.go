package gonf

import "crypto/sha256"

func sha256sum(contents []byte) []byte {
	h := sha256.New()
	h.Write(contents)
	return h.Sum(nil)
}
