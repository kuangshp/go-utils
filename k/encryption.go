package k

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// 高级加密标准
// 16,24,32位字符串加密的话，分别对用AES-128,AES-192,AES-256加密方法
// 定义一个key不能泄漏(这里采用16位的)
//var pwdKey = []byte("DIS**#KKKDJJSKDI")

// PKCS7Padding 定义填充模式
func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

// PKC7UnPadding 定义反填充(删除填充的字符串)
func PKC7UnPadding(origData []byte) ([]byte, error) {
	//获取数据长度
	length := len(origData)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	} else {
		//获取填充字符串长度
		unpadding := int(origData[length-1])
		//截取切片，删除填充字节，并且返回明文
		return origData[:(length - unpadding)], nil
	}
}

// AesEcrypt 加密操作
func AesEcrypt(origData []byte, key []byte) ([]byte, error) {
	// 创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 获取块的大小
	blockSize := block.BlockSize()
	// 对数据进行填充，让数据长度满足需求
	origData = PKCS7Padding(origData, blockSize)
	// 采用AES加密方法中的CBC加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 执行加密操作
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// AesDeCrypt 解密算法
func AesDeCrypt(cypted []byte, key []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块大小
	blockSize := block.BlockSize()
	//创建加密客户端实例
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(cypted))
	//这个函数也可以用来解密
	blockMode.CryptBlocks(origData, cypted)
	//去除填充字符串
	origData, err = PKC7UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, err
}

// EnPwdCode 加密后的byte转换为bs64
//func EnPwdCode(pwd []byte) (string, error) {
//	result, err := AesEcrypt(pwd, pwdKey)
//	if err != nil {
//		return "", nil
//	}
//	return base64.StdEncoding.EncodeToString(result), err
//}
//
//func DePwdCode(pwd string) ([]byte, error) {
//	pwdByte, err := base64.StdEncoding.DecodeString(pwd)
//	if err != nil {
//		return nil, err
//	}
//	// 执行解密
//	return AesDeCrypt(pwdByte, pwdKey)
//}
