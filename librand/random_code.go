package librand

import (
	"math/rand"
	"strings"
	"time"
)

var (
	// NUM码
	RandomNumItem []string

	// 小写
	RandomLowerCaseItem []string

	// 大写
	RandomUpperCaseItem []string
)

func init() {
	RandomUpperCaseItem = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	RandomNumItem = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	RandomLowerCaseItem = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
}

// 创建l个大写随机码
func CreateUpperRandomCode(l int) string {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	length := len(RandomUpperCaseItem)
	arr := make([]string, 0)
	for {
		index := rnd.Intn(length)
		arr = append(arr, RandomUpperCaseItem[index])

		if len(arr) == l {
			return strings.Join(arr, "")
		}
	}
}

// 创建l个小写随机码
func CreateLowerRandomCode(l int) string {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	length := len(RandomLowerCaseItem)
	arr := make([]string, 0)
	for {
		index := rnd.Intn(length)
		arr = append(arr, RandomLowerCaseItem[index])

		if len(arr) == l {
			return strings.Join(arr, "")
		}
	}
}

// 创建l个大写以及数字随机码
func CreateUpperNumRandomCode(l int) string {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	lengthUpper := len(RandomUpperCaseItem)
	lengthNum := len(RandomNumItem)
	arr := make([]string, 0)
	for {
		index := rnd.Intn(lengthNum + lengthUpper)
		if index < lengthUpper {

			arr = append(arr, RandomUpperCaseItem[index])
		} else {
			arr = append(arr, RandomNumItem[index-lengthUpper])
		}

		if len(arr) == l {
			return strings.Join(arr, "")
		}
	}
}

//  创建l个小写以及数字随机码
func CreateLowerNumRandomCode(l int) string {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	lengthLower := len(RandomLowerCaseItem)
	lengthNum := len(RandomNumItem)
	arr := make([]string, 0)
	for {
		index := rnd.Intn(lengthNum + lengthLower)
		if index < lengthLower {

			arr = append(arr, RandomLowerCaseItem[index])
		} else {
			arr = append(arr, RandomNumItem[index-lengthLower])
		}

		if len(arr) == l {
			return strings.Join(arr, "")
		}
	}
}

// 创建l个小写大写以及数字随机码
func CreateASCIIRandomCode(l int) string {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	lengthUpper := len(RandomUpperCaseItem)
	lengthLower := len(RandomLowerCaseItem)
	lengthNum := len(RandomNumItem)
	arr := make([]string, 0)
	for {
		index := rnd.Intn(lengthNum + lengthUpper + lengthLower)
		if index < lengthUpper {

			arr = append(arr, RandomUpperCaseItem[index])
		} else if index < lengthUpper+lengthLower {
			arr = append(arr, RandomLowerCaseItem[index-lengthUpper])
		} else {
			arr = append(arr, RandomNumItem[index-lengthUpper-lengthLower])
		}

		if len(arr) == l {
			return strings.Join(arr, "")
		}
	}
}

// 创建l个数字随机码
func CreateNumRandomCode(l int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	lengthNum := len(RandomNumItem)
	arr := make([]string, 0)
	for {
		index := rnd.Intn(lengthNum)
		arr = append(arr, RandomNumItem[index])
		if len(arr) == l {
			return strings.Join(arr, "")
		}
	}
}
