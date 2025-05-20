package chacha20blake3_test

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/bloom42/stdx-go/crypto/chacha12blake3"
	"github.com/bloom42/stdx-go/crypto/chacha20blake3"
	"github.com/bloom42/stdx-go/crypto/xchacha20sha256"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	BENCHMARKS = []int64{
		64,
		1024,
		16 * 1024,
		64 * 1024,
		1 * 1024 * 1024,
		10 * 1024 * 1024,
		// 100_000_000,
	}
)

// go test -benchmem -bench=. github.com/bloom42/stdx-go/crypto/chacha20blake3
func BenchmarkEncryptAEAD(b *testing.B) {
	additionalData := randBytes(b, 100)

	chaCha20Key := randBytes(b, chacha20.KeySize)
	xChaCha20Nonce := randBytes(b, chacha20.NonceSizeX)
	chaCha20Blake3Nonce := randBytes(b, chacha20blake3.NonceSize)

	for _, size := range BENCHMARKS {
		benchmarkEncrypt(b, size, "XChaCha20-Poly1305", newXChaCha20Poly1305Cipher(b, chaCha20Key), xChaCha20Nonce, additionalData)
		benchmarkEncrypt(b, size, "ChaCha12-BLAKE3", newChaCha12Blake3Cipher(b, chaCha20Key), chaCha20Blake3Nonce, additionalData)
		benchmarkEncrypt(b, size, "ChaCha20-BLAKE3", newChaCha20Blake3Cipher(b, chaCha20Key), chaCha20Blake3Nonce, additionalData)
		benchmarkEncrypt(b, size, "XChaCha20-SHA256", newXChaCha20Sha256Cipher(b, chaCha20Key), xChaCha20Nonce, additionalData)
	}
}

func BenchmarkDecryptAEAD(b *testing.B) {
	additionalData := randBytes(b, 100)

	chaCha20Key := randBytes(b, chacha20.KeySize)
	xChaCha20Nonce := randBytes(b, chacha20.NonceSizeX)
	chaCha20Blake3Nonce := randBytes(b, chacha20blake3.NonceSize)

	for _, size := range BENCHMARKS {
		benchmarkDecrypt(b, size, "XChaCha20-Poly1305", newXChaCha20Poly1305Cipher(b, chaCha20Key), xChaCha20Nonce, additionalData)
		benchmarkDecrypt(b, size, "ChaCha12-BLAKE3", newChaCha12Blake3Cipher(b, chaCha20Key), chaCha20Blake3Nonce, additionalData)
		benchmarkDecrypt(b, size, "ChaCha20-BLAKE3", newChaCha20Blake3Cipher(b, chaCha20Key), chaCha20Blake3Nonce, additionalData)
		benchmarkDecrypt(b, size, "XChaCha20-SHA256", newXChaCha20Sha256Cipher(b, chaCha20Key), xChaCha20Nonce, additionalData)
	}
}

func benchmarkEncrypt[C cipher.AEAD](b *testing.B, size int64, algorithm string, cipher C, nonce, additionalData []byte) {
	b.Run(fmt.Sprintf("%s-%s", bytesCount(size), algorithm), func(b *testing.B) {
		plaintext := randBytes(b, size)
		dst := make([]byte, 0, len(plaintext)+512)
		b.ReportAllocs()
		b.SetBytes(size)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cipher.Seal(dst, nonce, plaintext, additionalData)
		}
	})
}

func benchmarkDecrypt[C cipher.AEAD](b *testing.B, size int64, algorithm string, cipher C, nonce, additionalData []byte) {
	b.Run(fmt.Sprintf("%s-%s", bytesCount(size), algorithm), func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(size)
		plaintext := randBytes(b, size)
		cipherText := make([]byte, len(plaintext)+512)
		cipherText = cipher.Seal(cipherText, nonce, plaintext, additionalData)
		dst := make([]byte, len(cipherText))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cipher.Open(dst, nonce, cipherText, additionalData)
		}
	})
}

func newChaCha12Blake3Cipher(b *testing.B, key []byte) *chacha12blake3.ChaCha12Blake3 {
	cipher, err := chacha12blake3.New(key)
	if err != nil {
		b.Error(err)
	}

	return cipher
}

func newChaCha20Blake3Cipher(b *testing.B, key []byte) *chacha20blake3.ChaCha20Blake3 {
	cipher, err := chacha20blake3.New(key)
	if err != nil {
		b.Error(err)
	}

	return cipher
}

func newXChaCha20Sha256Cipher(b *testing.B, key []byte) *xchacha20sha256.XChaCha20Sha256 {
	cipher, err := xchacha20sha256.New(key)
	if err != nil {
		b.Error(err)
	}

	return cipher
}

func newXChaCha20Poly1305Cipher(b *testing.B, key []byte) cipher.AEAD {
	cipher, err := chacha20poly1305.NewX(key)
	if err != nil {
		b.Error(err)
	}

	return cipher
}

func randBytes(b *testing.B, n int64) []byte {
	buff := make([]byte, n)

	_, err := rand.Read(buff)
	if err != nil {
		b.Error(err)
	}

	return buff
}

func bytesCount(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
