package k

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// PKCS7Padding 定义填充模式
func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

// PKC7UnPadding 定义反填充(删除填充的字符串)
func PKC7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	unpadding := int(origData[length-1])
	if unpadding <= 0 || unpadding > length {
		return nil, errors.New("填充长度非法")
	}
	return origData[:(length - unpadding)], nil
}

// AesEcrypt 使用 AES-CBC 加密。
//
// 安全红线（遗留 API）：IV 取自 key 前缀，可预测，禁止用于新协议或新密钥材料。
// 新代码请使用 AesEncryptGCM。
func AesEcrypt(origData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(key) < blockSize {
		return nil, errors.New("密钥长度不足")
	}
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// AesDeCrypt AES-CBC 解密，对应 AesEcrypt（IV=key 前缀的遗留格式）。
func AesDeCrypt(cypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(key) < blockSize {
		return nil, errors.New("密钥长度不足")
	}
	if len(cypted) == 0 || len(cypted)%blockSize != 0 {
		return nil, errors.New("密文长度非法")
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(cypted))
	blockMode.CryptBlocks(origData, cypted)
	return PKC7UnPadding(origData)
}

// AesEncryptGCM 使用 AES-GCM 加密，输出格式为 nonce||ciphertext（含认证标签）。
// key 长度须为 16/24/32 字节。推荐用于新协议。
func AesEncryptGCM(plain []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("生成 nonce 失败: %w", err)
	}
	return gcm.Seal(nonce, nonce, plain, nil), nil
}

// AesDecryptGCM 解密 AesEncryptGCM 产出的 nonce||ciphertext。
func AesDecryptGCM(sealed []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(sealed) < nonceSize {
		return nil, errors.New("密文过短")
	}
	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
