package worker

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"

	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/iot/pkg/dyson"
)

var initialVector = []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
var aesKey = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}

func decipherLocalCredentials(localCredential string) (*dyson.Credentials, error) {
	aesBlock, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(aesBlock, initialVector)

	rawPassword, err := base64.RawStdEncoding.DecodeString(localCredential)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	mode.CryptBlocks(rawPassword, rawPassword)
	rawPassword = rawPassword[:len(rawPassword)-8]

	var credentials dyson.Credentials
	if err := json.Unmarshal(rawPassword, &credentials); err != nil {
		return nil, errors.WithStack(err)
	}

	return &credentials, nil
}
