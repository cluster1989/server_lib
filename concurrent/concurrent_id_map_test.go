package concurrent

import "testing"
import "fmt"
import "time"

type T1 struct {
}

func (t *T1) Close() error {
	fmt.Printf("t1 close ----\n")
	return nil
}

func TestConcurrentMap(t *testing.T) {
	ma := NewCocurrentIDGroup()
	obj0 := &T1{}
	ma.Set(50, obj0)
	obj1 := &T1{}
	ma.Set(10, obj1)
	ma.Del(10)
	time.Sleep(time.Duration(2) * time.Second)
	obj2 := &T1{}
	ma.Set(30, obj2)
	obj3 := &T1{}
	ma.Set(20, obj3)
	obj4 := &T1{}
	ma.Set(40, obj4)
	ma.Dispose()
}
