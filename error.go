package azblob

import "github.com/aos-dev/go-storage/v3/services"

var (
	// ErrEncryptionKeyInvalid will be returned while encryption key is invalid.
	// Encryption key must be a 32-byte AES-256 key.
	ErrEncryptionKeyInvalid = services.NewErrorCode("invalid encryption key")
)
