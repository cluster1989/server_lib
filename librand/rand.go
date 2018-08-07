package librand

import (
	"math/rand"
	"time"
)

func UniqRand(l int, n int) []int {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	set := make(map[int]struct{})

	nums := make([]int, 0, l)
	for {
		num := rnd.Intn(n)
		if _, ok := set[num]; !ok {
			set[num] = struct{}{}
			nums = append(nums, num)
		}
		if len(nums) == l {
			return nums
		}
	}

}

func NormalRand(l int, n int) []int {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	nums := make([]int, 0, l)
	for {
		num := rnd.Intn(n)
		nums = append(nums, num)
		if len(nums) == l {
			return nums
		}
	}

}

// NormalFromToRand 包含from，不包含to
func NormalFromToRand(l int, from int, to int) []int {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	nums := make([]int, 0, l)
	for {
		num := rnd.Intn(to-from) + from
		nums = append(nums, num)
		if len(nums) == l {
			return nums
		}
	}
}
