package libsnowflake

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// 雪花算法的id部分, 总共64位
const (
	// BitLenTime 时间序列部分长度
	BitLenTime = 39
	// BitLenSequence 自增部分长度
	BitLenSequence = 8
	// BitLenMachineID 机器码部分长度
	BitLenMachineID = 63 - BitLenTime - BitLenSequence
)

// Setting 雪花算法的设置
type Setting struct {
	// 开始时间不能高于当前时间,可以为0,或者不设置
	StartTime time.Time
	// 机器码,如果返回error,或者机器码为空,都不会执行snowflake
	MachineID func() (uint16, error)
	// 检查机器码,如果是false ,则不会创建snowflake
	CheckMachineID func(uint16) bool
}

// SnowFlake 雪花算法的主类
type SnowFlake struct {
	mutex       *sync.Mutex
	startTime   int64
	elapsedTime int64
	sequence    uint16
	machineID   uint16
}

// New 创建snowflake对象
func New(st Setting) *SnowFlake {
	sf := new(SnowFlake)
	sf.mutex = new(sync.Mutex)
	// 变成最高序列
	sf.sequence = uint16(1<<BitLenSequence - 1)
	if st.StartTime.After(time.Now()) {
		return nil
	}

	if st.StartTime.IsZero() {
		// 取当前的utc时间
		sf.startTime = toSnowFlakeTime(time.Now().UTC())
	} else {
		sf.startTime = toSnowFlakeTime(st.StartTime)
	}

	var err error
	if st.MachineID == nil {
		sf.machineID, err = lower16BitPrivateIP()
	} else {
		sf.machineID, err = st.MachineID()
	}
	if err != nil || (st.CheckMachineID != nil && !st.CheckMachineID(sf.machineID)) {
		return nil
	}
	return sf
}

// 取id
func (sf *SnowFlake) NextID() (uint64, error) {

	const maskSequence = uint16(1<<BitLenSequence - 1)
	sf.mutex.Lock()
	defer sf.mutex.Unlock()
	current := currentElapsedTime(sf.startTime)
	if sf.elapsedTime < current {
		// 重置时间
		sf.elapsedTime = current
		sf.sequence = 0
	} else {
		sf.sequence = (sf.sequence + 1) & maskSequence
		// 被重置之后
		if sf.sequence == 0 {
			sf.elapsedTime++
			overtime := sf.elapsedTime - current
			time.Sleep(sleepTime(overtime))
		}
	}
	return sf.toID()
}

func (sf *SnowFlake) toID() (uint64, error) {
	if sf.elapsedTime >= 1<<BitLenTime {
		return 0, errors.New("over the time limit")
	}

	return uint64(sf.elapsedTime)<<(BitLenSequence+BitLenMachineID) |
		uint64(sf.sequence)<<BitLenMachineID |
		uint64(sf.machineID), nil
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond - time.Duration(time.Now().UTC().UnixNano()%snowflakeTimeUnit)*time.Nanosecond
}

const snowflakeTimeUnit = 1e7

// 转化为snowflake时间,保持37位
func toSnowFlakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / snowflakeTimeUnit
}

// 当前已经运行的时间
func currentElapsedTime(startTime int64) int64 {
	return toSnowFlakeTime(time.Now().UTC()) - startTime
}

// 取本机的ip
func lower16BitPrivateIP() (uint16, error) {
	ip, err := getPrivateIPv4()
	if err != nil {
		return 0, err
	}
	// 取192.168.0.1,第3位和第4位
	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

func getPrivateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		// 去除127.0.0.1的地址
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		// 查询私有的ip地址
		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}
