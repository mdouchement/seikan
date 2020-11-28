package noise

// A HandshakePattern is an handshake pattern.
type HandshakePattern uint8

// Supported patterns.
const (
	PatternIK HandshakePattern = 0x01
)

// An HashFunction is a cryptographic hash function.
type HashFunction uint8

// Supported hash functions.
const (
	HashBlake2b HashFunction = 0x01
	HashBlake2s HashFunction = 0x02
)

// An CipherFunction is an AEAD symmetric cipher.
type CipherFunction uint8

// Supported hash functions.
const (
	CipherChaCha20Poly1305 CipherFunction = 0x01
	CipherAES256GCM        CipherFunction = 0x02
)

// Streaming constants.
const (
	SizePrefixLength = 2
	ChunkSize        = 0xFFFF
)
