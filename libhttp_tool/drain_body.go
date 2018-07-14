package libhttp_tool

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// 获取http的body，并且再填充
func DrainAndResetHTTPBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, io.EOF
	}

	buf, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}
	err = req.Body.Close()
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

	return buf, nil
}
