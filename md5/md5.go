package md5

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"unsafe"
)

type MD5Info struct {
	Md5  string
	File string
}

func (this *MD5Info) Write(buff []byte) (int, error) {
	info := *(*string)(unsafe.Pointer(&buff))

	infos := strings.Split(info, " ")
	if len(infos) < 2 {
		return 0, errors.New("get md5 error")
	}

	this.Md5 = infos[0]
	this.File = infos[1]

	return len(buff), nil
}

func MD5(buff []byte) (string, []byte, error) {
	h := md5.New()

	_, err := h.Write(buff)
	if err != nil {
		return "", nil, err
	}

	m := h.Sum(nil)
	return hex.EncodeToString(m), m, nil
}

func IoMD5(read io.Reader) (string, []byte, error) {
	h := md5.New()

	buff := make([]byte, h.BlockSize())
	for {
		_, err := read.Read(buff)
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", nil, err
		}

		_, err = h.Write(buff)
		if err != nil {
			return "", nil, err
		}
	}

	m := h.Sum(nil)
	return hex.EncodeToString(m), m, nil
}

//Do not recommend, a mistake method
func FileMD5(filename string) (string, []byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	m, h, err := IoMD5(file)
	if err != nil {
		return "", nil, err
	}

	return m, h, nil
}

//Recommend
func MD5CMD(file string) (string, []byte, error) {
	cmd := exec.Command("md5sum", file)

	info := &MD5Info{}
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = info

	err := cmd.Run()
	if err != nil {
		return "", nil, err
	}

	return info.Md5, *(*[]byte)(unsafe.Pointer(&info.Md5)), nil
}
