package secure

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)
	c, err := NewNIKCipher(key, key)
	if err != nil {
		t.Fatal(err)
	}
	nik := "3273010101900001"
	enc, err := c.Encrypt(nik)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(enc, []byte(nik)) {
		t.Fatal("NIK terlihat di ciphertext")
	}
	got, err := c.Decrypt(enc)
	if err != nil || got != nik {
		t.Fatalf("decrypt = %q, %v", got, err)
	}
	if !bytes.Equal(c.Hash(nik), c.Hash(nik)) {
		t.Fatal("hash tidak deterministik")
	}
}

func TestMaskAndValid(t *testing.T) {
	if MaskNIK("3273010101900001") != "3273********0001" {
		t.Fatal("mask salah")
	}
	if ValidNIK("12345") || !ValidNIK("3273010101900001") || ValidNIK("327301010190000a") {
		t.Fatal("validasi salah")
	}
}
