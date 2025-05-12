package schacha20blake3_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/bloom42/stdx-go/crypto/schacha20blake3"
)

func TestBasic(t *testing.T) {
	var key [schacha20blake3.KeySize]byte
	var nonce [schacha20blake3.NonceSize]byte

	originalPlaintext := []byte("Hello World")
	additionalData := []byte("!")

	rand.Read(key[:])
	rand.Read(nonce[:])

	cipher, _ := schacha20blake3.New(key[:])
	ciphertext := cipher.Seal(nil, nonce[:], originalPlaintext, additionalData)

	decryptedPlaintext, err := cipher.Open(nil, nonce[:], ciphertext, additionalData)
	if err != nil {
		t.Errorf("decrypting message: %s", err)
		return
	}

	if !bytes.Equal(decryptedPlaintext, originalPlaintext) {
		t.Errorf("original message (%s) != decrypted message (%s)", string(originalPlaintext), string(decryptedPlaintext))
		return
	}
}

func TestAdditionalData(t *testing.T) {
	var key [schacha20blake3.KeySize]byte
	var nonce [schacha20blake3.NonceSize]byte

	originalPlaintext := []byte("Hello World")
	additionalData := []byte("!")

	rand.Read(key[:])
	rand.Read(nonce[:])

	cipher, _ := schacha20blake3.New(key[:])
	ciphertext := cipher.Seal(nil, nonce[:], originalPlaintext, additionalData)

	_, err := cipher.Open(nil, nonce[:], ciphertext, []byte{})
	if !errors.Is(err, schacha20blake3.ErrOpen) {
		t.Errorf("expected error (%s) | got (%s)", schacha20blake3.ErrOpen, err)
		return
	}
}
