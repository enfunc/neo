package neo

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/insecurecleartextkeyset"
	"github.com/google/tink/go/keyset"
)

// Encrypter encrypts the given string.
type Encrypter func(s string) (string, error)

// NewEncrypter creates an encryption function from the JSON encryption key at the given filepath.
func NewEncrypter(jsonPath string) (Encrypter, error) {
	f, err := os.Open(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("neo: failed to open the encryption key file: %w", err)
	}
	defer f.Close()
	kh, err := insecurecleartextkeyset.Read(keyset.NewJSONReader(f))
	if err != nil {
		return nil, fmt.Errorf("neo: failed to read the encryption key file: %w", err)
	}
	a, err := aead.New(kh)
	if err != nil {
		return nil, fmt.Errorf("neo: failed to create an AEAD primitive: %w", err)
	}
	var empty []byte
	return func(s string) (string, error) {
		enc, err := a.Encrypt([]byte(s), empty)
		if err != nil {
			return "", fmt.Errorf("neo: failed to encrypt the given string: %w", err)
		}
		b64 := base64.StdEncoding.EncodeToString(enc)
		return b64, nil
	}, nil
}
