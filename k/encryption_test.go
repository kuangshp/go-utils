package k

import (
	"bytes"
	"testing"
)

func TestAesCBC_LegacyRoundTrip(t *testing.T) {
	key := []byte("1234567890abcdef")
	plain := []byte("hello-go-utils")
	crypted, err := AesEcrypt(plain, key)
	if err != nil {
		t.Fatalf("AesEcrypt: %v", err)
	}
	got, err := AesDeCrypt(crypted, key)
	if err != nil {
		t.Fatalf("AesDeCrypt: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("got %q want %q", got, plain)
	}
}

func TestAesCBC_InvalidKey(t *testing.T) {
	if _, err := AesEcrypt([]byte("x"), []byte("short")); err == nil {
		t.Fatal("短密钥应失败")
	}
}

func TestAesGCM_RoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef")
	plain := []byte("secure-payload")
	sealed, err := AesEncryptGCM(plain, key)
	if err != nil {
		t.Fatalf("AesEncryptGCM: %v", err)
	}
	got, err := AesDecryptGCM(sealed, key)
	if err != nil {
		t.Fatalf("AesDecryptGCM: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("got %q want %q", got, plain)
	}
	if _, err := AesDecryptGCM(sealed[:2], key); err == nil {
		t.Fatal("过短密文应失败")
	}
}
