package libzookeeper

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"
)

// 切分zk地址和他的根目录
func ParseZKAddrString(zkAddr string) (nodes []string, chroot string) {
	strArr := strings.SplitN(zkAddr, "/", 2)
	if len(strArr) == 2 {
		chroot = fmt.Sprintf("/%s", strArr[1])
	}
	nodes = strings.Split(strArr[0], ",")
	return
}

// 组合zk的地址
func ContractZKAddrs(zkAddrs []string) string {
	return strings.Join(zkAddrs, ",")
}

func GenerateUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// 生成consumer的实例的id
func GenerateConsumerInstanceID() (string, error) {
	uuid, err := GenerateUUID()
	if err != nil {
		return "", err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s:%s", hostname, uuid), nil
}
