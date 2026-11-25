// Package secure berisi enkripsi dan penyamaran NIK sesuai prinsip UU PDP.
package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"regexp"
)

var nikPattern = regexp.MustCompile(`^\d{16}$`)

// ValidNIK memeriksa format NIK (16 digit angka).
func ValidNIK(nik string) bool { return nikPattern.MatchString(nik) }

// MaskNIK menampilkan 4 digit awal dan 4 digit akhir saja, contoh 3273********0001.
func MaskNIK(nik string) string {
	if len(nik) != 16 {
		return "****************"
	}
	return nik[:4] + "********" + nik[12:]
}

type NIKCipher struct {
	aead    cipher.AEAD
	hashKey []byte
}

func NewNIKCipher(encKey, hashKey []byte) (*NIKCipher, error) {
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &NIKCipher{aead: aead, hashKey: hashKey}, nil
}

// Hash menghasilkan indeks buta (blind index) untuk mencari NIK tanpa menyimpan NIK asli.
func (c *NIKCipher) Hash(nik string) []byte {
	m := hmac.New(sha256.New, c.hashKey)
	m.Write([]byte(nik))
	return m.Sum(nil)
}

func (c *NIKCipher) Encrypt(nik string) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return c.aead.Seal(nonce, nonce, []byte(nik), nil), nil
}

func (c *NIKCipher) Decrypt(data []byte) (string, error) {
	ns := c.aead.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext terlalu pendek")
	}
	pt, err := c.aead.Open(nil, data[:ns], data[ns:], nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}
