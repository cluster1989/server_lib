package hmacsha1_encrypt

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

func Hmacsha1(key, data []byte) []byte {

	//hmac ,use sha1
	mac := hmac.New(sha1.New, key)
	mac.Write(data)
	b := mac.Sum(nil)
	return b
}

func Hmacsha12String(key, data []byte) string {

	b := Hmacsha1(key, data)
	return hex.EncodeToString(b)
}
