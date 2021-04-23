package azblob

import "errors"

var (
	// ErrInvalidEncryptionKey will be returned while encryption key is invalid.
	// Encryption key must be a 32-byte AES-256 key.
	ErrInvalidEncryptionKey = errors.New("invalid encryption key")
)
