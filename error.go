package azblob

import "errors"

var (
	// ErrInvalidEncryptionCustomerKey will be returned while encryption customer key is invalid.
	// Encryption key must be a 32-byte AES-256 key.
	ErrInvalidEncryptionCustomerKey = errors.New("invalid encryption customer key")
)
