package rsa_encrypt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var (
	// 公钥出错
	RSAPublicKeyErr = errors.New("rsa public error")
)

// 根据size打包组装数据
func pactetData(data []byte, size int) [][]byte {
	src := make([]byte, len(data))
	copy(src, data)
	val := make([][]byte, 0)
	if len(src) <= size {
		return append(val, src)
	}
	for len(src) > 0 {
		p := src[:size]
		val = append(val, p)
		src = src[size:]
		if len(src) < size {
			val = append(val, src)
			break
		}
	}
	return val
}

// rsa 加密
func RSAEncrypt(text, key []byte) ([]byte, error) {
	block, _ := pem.Decode(key)

	if block == nil {
		return nil, RSAPublicKeyErr

	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	pub := pubInterface.(*rsa.PublicKey)

	data := pactetData(text, pub.N.BitLen()/8-11)

	cryptData := make([]byte, 0, 0)

	for _, val := range data {
		c, err := rsa.EncryptPKCS1v15(rand.Reader, pub, val)
		if err != nil {
			return nil, err
		}
		cryptData = append(cryptData, c...)
	}
	return cryptData, nil
}

// rsa 解密
func RSADecrypt(cryptData, key []byte) ([]byte, error) {

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, RSAPublicKeyErr
	}
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	data := pactetData(cryptData, pri.PublicKey.N.BitLen()/8)
	decryptData := make([]byte, 0, 0)
	for _, val := range data {
		c, err := rsa.DecryptPKCS1v15(rand.Reader, pri, val)
		if err != nil {
			return nil, err
		}
		decryptData = append(decryptData, c...)
	}
	return decryptData, nil
}

// 签名
func SignPKCS1v15(src, key []byte, hash crypto.Hash) ([]byte, error) {
	var h = hash.New()
	h.Write(src)
	var hashed = h.Sum(nil)

	var err error
	var block *pem.Block
	block, _ = pem.Decode(key)
	if block == nil {
		return nil, errors.New("private key error")
	}

	var pri *rsa.PrivateKey
	pri, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.SignPKCS1v15(rand.Reader, pri, hash, hashed)
}

// 验证签名
func VerifyPKCS1v15(src, sig, key []byte, hash crypto.Hash) error {
	var h = hash.New()
	h.Write(src)
	var hashed = h.Sum(nil)

	var err error
	var block *pem.Block
	block, _ = pem.Decode(key)
	if block == nil {
		return errors.New("public key error")
	}

	var pubInterface interface{}
	pubInterface, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	var pub = pubInterface.(*rsa.PublicKey)

	return rsa.VerifyPKCS1v15(pub, hash, hashed, sig)
}
