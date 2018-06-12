package base64_encrypt

import (
	"encoding/base64"
)

func Base64StdEncode(buff []byte) string {
	return base64.StdEncoding.EncodeToString(buff)
}

func Base64StdDecode(val string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(val)
}
