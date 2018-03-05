package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
)

// ShaFingerprint return SHA1 fingerprint
func ShaFingerprint(content interface{}) string {
	hasher := sha1.New()
	if _, err := hasher.Write([]byte(fmt.Sprintf(`%v`, content))); err != nil {
		log.Printf(`Error while generating hash for %s: %v`, content, err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}
