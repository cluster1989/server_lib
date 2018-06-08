package hmacsha1_encrypt

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

func Hmacsha1(key, data []byte) []byte {

	//hmac ,use sha1
	mac := hmac.New(sha1.New, key)
	mac.Write(data)
	fmt.Printf("%x\n", mac.Sum(nil))
	b := mac.Sum(nil)
	return b
}

func Hmacsha12String(key, data []byte) string {

	b := Hmacsha1(key, data)
	return hex.EncodeToString(b)
}
