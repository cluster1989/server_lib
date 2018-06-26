package libredis

import (
	"fmt"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
)

var redisCfg = Options{
	MaxIdle:            8,
	MaxActive:          64,
	IdleTimeout:        300,
	RedisAddress:       "127.0.0.1:6379",
	DialConnectTimeout: 30,
	DialReadTimeout:    30,
	DialWriteTimeout:   30,
	Auth:               "",
	DbNum:              3,
}

func Test_setVAL(t *testing.T) {
	cache := NewCache(&redisCfg)
	time.Sleep(3)
	if err := cache.Setex("name777", "nick", time.Duration(30)*time.Second); err != nil {
		t.Error(err)
	}
	if !cache.IsExists("name777") {
		t.Error("cache set error")
	}
	val := cache.GetValue("name")
	printRedisVal(val)

	if err := cache.Set("name886", "broto"); err != nil {
		t.Error(err)
	}
	if !cache.IsExists("name886") {
		t.Error("cache set error")
	}
	val = cache.GetValue("name886")
	printRedisVal(val)
}

func Test_Hsetget(t *testing.T) {
	cache := NewCache(&redisCfg)
	if err := cache.HSet("name", "nick", "wuqifei"); err != nil {
		t.Error(err)
	}
	val := cache.HGet("name", "luck")
	printRedisVal(val)

}

func printRedisVal(val interface{}) {
	var err error
	str, _ := redis.String(val, err)
	fmt.Print("redis val:", str, "\n")
}

func Test_Multi(t *testing.T) {
	cache := NewCache(&redisCfg)
	cache.Set("name881", "broto1")
	cache.Set("name882", "broto2")
	cache.Set("name883", "broto3")
	cache.Set("name884", "broto4")
	val := cache.MultiGet([]string{"name881", "name882", "name883", "name884"})

	for _, v := range val {
		var err error
		str, _ := redis.String(v, err)
		fmt.Print("redis val:", str, "\n")
	}
}

func Test_Trans(t *testing.T) {

	cache := NewCache(&redisCfg)
	_, err := cache.Transaction(func() (r []*RedisTransactionRet, errCode int) {
		ret := []*RedisTransactionRet{}

		ret = append(ret, &RedisTransactionRet{
			Table: "test1",
			Key:   "age",
			Value: "asd",
		})
		ret = append(ret, &RedisTransactionRet{
			Table: "test1",
			Key:   "name",
			Value: "broto",
		})
		ret = append(ret, &RedisTransactionRet{
			Table: "test1",
			Key:   "num",
			Value: "fas",
		})
		ret = append(ret, &RedisTransactionRet{
			Table: "test1",
			Key:   "naks",
			Value: "123",
		})
		r = ret
		errCode = 0
		return
	}, "test1:age", "test1:name", "test1:num")
	if err != nil {
		t.Error(err)
	}

	val := cache.HGet("test1", "age")
	printRedisVal(val)

}

func Test_HMOption(t *testing.T) {
	cache := NewCache(&redisCfg)
	if err := cache.HMSet("tes1tkey1", "abc", "123", "bas", "234"); err != nil {
		t.Error(err)
	}

	if _, err := cache.HGETALL("tes1tkey1"); err != nil {
		t.Error(err)
	}

}

func Test_SET(t *testing.T) {

	cache := NewCache(&redisCfg)
	err := cache.SAdd("test", "test1", "test2")
	if err != nil {
		t.Error(err)
	}
	val, err := cache.SMembers("test")
	if err != nil {
		t.Error(err)
	}

	t.Log(val)

}

func Test_ZSET(t *testing.T) {
	cache := NewCache(&redisCfg)
	err := cache.ZADD("test123", 8, "test212")
	if err != nil {
		t.Error(err)
	}
	val, err := cache.ZCARD("test123")
	if err != nil {
		t.Error(err)
	}

	vals, err := cache.ZRangeByScoreCARD("test123", 2, 8)

	if err != nil {

		t.Error(err)
	}

	vals, err = cache.ZRANGE("test123", 0, 2)
	t.Log(val)
	t.Log(vals)
}
