package rsa_encrypt

import (
	"encoding/pem"
	"errors"
)

var (
	// 公钥出错
	RSAPublicKeyErr = errors.New("rsa public error")
)

func pactetData(data []byte, size int) [][]byte {
	src := make([]byte, len(data))
	copy(src, data)
	val := make([][]byte, 0)
	if len(src) <= size {
		return append(val, src)
	}
	if len(src) > 0 {
		// p :=
	}
	return val
}

// rsa 加密
func RSAEncrypt(text, key []byte) ([]byte, error) {
	block, _ := pem.Decode(key)

	if block == nil {
		return nil, RSAPublicKeyErr

	}

	// pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)

	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}
