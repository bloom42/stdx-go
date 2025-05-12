package bchacha20blake3

import (
	"crypto/cipher"
	"crypto/subtle"
	"encoding/binary"
	"errors"

	"github.com/bloom42/stdx-go/crypto/blake3"
	"github.com/bloom42/stdx-go/crypto/chacha20"
	// "golang.org/x/crypto/chacha20"
)

const (
	KeySize   = 32
	NonceSize = 32
	TagSize   = 32

	// encryptionKeyContext    = "ChaCha20-BLAKE3 2023-12-31 encryption key ChaCha20"
	encryptionKeyContext    = "ChaCha20-BLAKE3 encryption key ChaCha20"
	athenticationKeyContext = "ChaCha20-BLAKE3 authentication key BLAKE3"
)

var (
	ErrOpen = errors.New("chacha20blake3: error decrypting ciphertext")
)

type ChaCha20Blake3 struct {
	key [KeySize]byte
}

// ensure that ChaCha20Blake3 implements `cipher.AEAD` interface at build time
var _ cipher.AEAD = (*ChaCha20Blake3)(nil)

func New(key []byte) (*ChaCha20Blake3, error) {
	ret := new(ChaCha20Blake3)
	copy(ret.key[:], key)
	return ret, nil
}

func (*ChaCha20Blake3) NonceSize() int {
	return NonceSize
}

func (*ChaCha20Blake3) Overhead() int {
	return TagSize
}

func (x *ChaCha20Blake3) Seal(dst, nonce, plaintext, additionalData []byte) []byte {
	var encryptionKey [32]byte
	var authenticationKey [32]byte

	deriveKey(encryptionKey[:], x.key[:], encryptionKeyContext, nil)
	deriveKey(authenticationKey[:], x.key[:], athenticationKeyContext, nonce)

	ret, out := sliceForAppend(dst, len(plaintext)+TagSize)
	ciphertext, tag := out[:len(plaintext)], out[len(plaintext):]

	chacha20Cipher, _ := chacha20.New(encryptionKey[:], nonce[24:32])
	// chacha20Cipher, _ := chacha20.NewUnauthenticatedCipher(encryptionKey[:], nonce[20:32])
	chacha20Cipher.XORKeyStream(ciphertext, plaintext)

	// _ = tag
	macHasher := blake3.New(32, authenticationKey[:])
	macHasher.Write(additionalData)
	// macHasher.Write(nonce)
	macHasher.Write(ciphertext)
	writeUint64LittleEndian(macHasher, uint64(len(additionalData)))
	// writeUint64(macHasher, uint64(len(nonce)))
	writeUint64LittleEndian(macHasher, uint64(len(ciphertext)))
	macHasher.Sum(tag[:0])

	zeroize(encryptionKey[:])
	zeroize(authenticationKey[:])

	return ret
}

func (x *ChaCha20Blake3) Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	var encryptionKey [32]byte
	var authenticationKey [32]byte

	deriveKey(encryptionKey[:], x.key[:], encryptionKeyContext, nil)
	deriveKey(authenticationKey[:], x.key[:], athenticationKeyContext, nonce)

	tag := ciphertext[len(ciphertext)-TagSize:]
	ciphertext = ciphertext[:len(ciphertext)-TagSize]

	chacha20Cipher, _ := chacha20.New(encryptionKey[:], nonce[24:32])
	// chacha20Cipher, _ := chacha20.NewUnauthenticatedCipher(encryptionKey[:], nonce[20:32])

	var computedTag [TagSize]byte
	macHasher := blake3.New(32, authenticationKey[:])
	macHasher.Write(additionalData)
	// macHasher.Write(nonce)
	macHasher.Write(ciphertext)
	writeUint64LittleEndian(macHasher, uint64(len(additionalData)))
	// writeUint64(macHasher, uint64(len(nonce)))
	writeUint64LittleEndian(macHasher, uint64(len(ciphertext)))
	macHasher.Sum(computedTag[:0])

	ret, plaintext := sliceForAppend(dst, len(ciphertext))

	if subtle.ConstantTimeCompare(computedTag[:], tag) != 1 {
		return nil, ErrOpen
	}

	chacha20Cipher.XORKeyStream(plaintext, ciphertext)

	zeroize(encryptionKey[:])
	zeroize(authenticationKey[:])

	return ret, nil
}

func deriveKey(out, parentKey []byte, context string, nonce []byte) {
	// it seems that it's faster to use slices as arguments instead of arrays, as slices are passed as pointers

	// we use a fixed-size array even if nonce is null to avoid heap allocations
	var keyMaterial [KeySize + NonceSize]byte

	copy(keyMaterial[:], nonce)
	copy(keyMaterial[len(nonce):], parentKey[:])

	blake3.DeriveKey(out[:], context, keyMaterial[:len(nonce)+len(parentKey)])

	// hasher := blake3.NewDeriveKey(context)
	// hasher.Write(nonce)
	// hasher.Write(parentKey)
	// writeUint64(hasher, uint64(len(nonce)))
	// writeUint64(hasher, uint64(len(parentKey)))
	// hasher.Sum(out)
}

// sliceForAppend takes a slice and a requested number of bytes. It returns a
// slice with the contents of the given slice followed by that many bytes and a
// second slice that aliases into it and contains only the extra bytes. If the
// original slice has sufficient capacity then no allocation is performed.
func sliceForAppend(in []byte, n int) (head, tail []byte) {
	if total := len(in) + n; cap(in) >= total {
		head = in[:total]
	} else {
		head = make([]byte, total)
		copy(head, in)
	}
	tail = head[len(in):]
	return
}

func writeUint64LittleEndian(p *blake3.Hasher, n uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], n)
	p.Write(buf[:])
}

func zeroize(input []byte) {
	for i := range input {
		input[i] = 0
	}
}
