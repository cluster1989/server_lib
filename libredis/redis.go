package libredis

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Options struct {
	DialReadTimeout    time.Duration
	DialWriteTimeout   time.Duration
	DialConnectTimeout time.Duration
	IdleTimeout        time.Duration
	RedisAddress       string
	MaxIdle            int //zuida kongxian
	MaxActive          int
	Auth               string
	DbNum              int
}

type RedisTransactionRet struct {
	Table string
	Key   interface{}
	Value interface{}
}

type RedisPool struct {
	options *Options
	redis.Pool
}

func NewConf() *Options {
	o := &Options{
		MaxIdle:            8,
		MaxActive:          64,
		IdleTimeout:        300,
		DialConnectTimeout: 30,
		DialReadTimeout:    30,
		DialWriteTimeout:   30,
	}
	return o
}

func NewCache(options *Options) *RedisPool {
	r := new(RedisPool)
	r.options = options
	r.initRedis()
	conn := r.Get() //显初始化一个conn，在redispool中
	defer conn.Close()
	return r
}

func (r *RedisPool) initRedis() {
	r.Pool = redis.Pool{
		MaxIdle:     r.options.MaxIdle,   // 最大空闲连接数
		MaxActive:   r.options.MaxActive, // 一个pool所能分配的最大的连接数目
		IdleTimeout: r.options.IdleTimeout,
		Dial:        r.Dial,
	}
}

func (r *RedisPool) Dial() (c redis.Conn, err error) {
	c, err = redis.Dial("tcp", r.options.RedisAddress)
	if err != nil {
		return
	}
	if len(r.options.Auth) != 0 {
		if _, err = c.Do("AUTH", r.options.Auth); err != nil {
			c.Close()
			return
		}
	}
	if r.options.DbNum > 0 {
		if _, err = c.Do("SELECT", r.options.DbNum); err != nil {
			c.Close()
			return
		}
	}
	return
}

// 执行redis命令
func (r *RedisPool) DoRedis(commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := r.Get()
	defer conn.Close()
	return conn.Do(commandName, args...)
}

// 封装事务
func (r *RedisPool) Transaction(callback func() (r []*RedisTransactionRet, errCode int), key ...string) (code int, err error) {
	conn := r.Pool.Get()
	defer conn.Close()

	conn.Send("WATCH", key)
	ret, errCode := callback()
	if ret == nil {
		conn.Send("UNWATCH")
		return errCode, errors.New("callback return error")
	}
	conn.Send("MULTI")
	for _, v := range ret {
		if conn.Send("HSET", v.Table, v.Key, v.Value) != nil {
			conn.Send("DISCARD")
			break
		}
	}
	_, err = conn.Do("EXEC")
	return 0, err
}

// 通过key 获取value
func (r *RedisPool) GetValue(key string) interface{} {
	if v, err := r.DoRedis("GET", key); err == nil {
		return v
	}
	return nil
}

// 通过hget方式获取value
func (r *RedisPool) HGet(key string, field interface{}) interface{} {
	if v, err := r.DoRedis("HGET", key, field); err == nil {
		return v
	}
	return nil
}

// 获取多个value
func (r *RedisPool) MultiGet(keys []string) []interface{} {
	var rv []interface{}
	size := len(keys)
	var err error
	conn := r.Get()
	defer conn.Close()
	for _, key := range keys {
		if err = conn.Send("GET", key); err != nil {
			goto ERROR
		}
	}

	if err = conn.Flush(); err != nil {
		goto ERROR
	}
	for i := 0; i < size; i++ {
		if v, err := conn.Receive(); err == nil {
			rv = append(rv, v.([]byte))
		} else {
			goto ERROR
		}
	}
	return rv
ERROR:
	rv = rv[0:0] //清空
	for i := 0; i < size; i++ {
		rv = append(rv, nil)
	}
	return rv
}

// 通过hget方式，获得多值
func (r *RedisPool) MultiHGet(key string, fields []interface{}) []interface{} {
	size := len(fields)
	var rv []interface{}
	conn := r.Get()
	defer conn.Close()
	var err error
	for _, filed := range fields {
		err = conn.Send("HGET", key, filed)
		if err != nil {
			goto ERROR
		}
	}
	if err = conn.Flush(); err != nil {
		goto ERROR
	}
	for i := 0; i < size; i++ {
		if v, err := conn.Receive(); err == nil {
			rv = append(rv, v.([]byte))
		} else {
			rv = append(rv, err)
		}
	}
	return rv
ERROR:
	rv = rv[0:0]
	for i := 0; i < size; i++ {
		rv = append(rv, nil)
	}
	return rv
}

// HSET redis
func (r *RedisPool) HSet(key string, field interface{}, val interface{}) error {
	if _, err := r.DoRedis("HSET", key, field, val); err != nil {
		return err
	}
	return nil
}

// setex redis
func (r *RedisPool) Setex(key string, val interface{}, timeout time.Duration) error {
	if _, err := r.DoRedis("SETEX", key, int64(timeout/time.Second), val); err != nil {
		return err
	}
	return nil
}

// set redis
func (r *RedisPool) Set(key string, val interface{}) error {
	if _, err := r.DoRedis("SET", key, val); err != nil {
		return err
	}
	return nil
}

// 根据key 删除value
func (r *RedisPool) Del(key string) error {
	if _, err := r.DoRedis("DEL", key); err != nil {
		return err
	}
	return nil
}

// 根据key 删除value
func (r *RedisPool) HDel(key string, field interface{}) error {
	if _, err := r.DoRedis("HDEL", key, field); err != nil {
		return err
	}
	return nil
}

func (r *RedisPool) IsExists(key string) bool {
	v, err := redis.Bool(r.DoRedis("EXISTS", key))
	if err != nil {
		return false
	}
	return v
}

// 给redis的key，延长或者减少过期时间
func (r *RedisPool) ExpireKey(key string, timeout time.Duration) bool {
	_, err := r.DoRedis("Expire", key, int64(timeout/time.Second))
	if err != nil {
		return false
	}
	return true
}

// 指定自增key
func (r *RedisPool) Incr(key string) error {
	_, err := redis.Bool(r.DoRedis("INCRBY", key, 1))
	return err
}

// 指定自减key
func (r *RedisPool) Decr(key string) error {
	_, err := redis.Bool(r.DoRedis("INCRBY", key, -1))
	return err
}

// redis hmset
func (r *RedisPool) HMSet(key string, args ...interface{}) error {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, key)
	rArgs = append(rArgs, args...)

	if _, err := r.DoRedis("HMSET", rArgs...); err != nil {
		return err
	}
	return nil
}

// redis hmget
func (r *RedisPool) HGETALL(key string) (map[string]interface{}, error) {
	reply, err := r.DoRedis("HGETALL", key)
	if err != nil {
		return nil, err
	}
	arr := reply.([]interface{})
	m := make(map[string]interface{})
	for i := 0; i < len(arr); i = i + 2 {
		key := arr[i].([]uint8)
		val := arr[i+1].([]uint8)
		m[string(key)] = val

	}
	return m, nil
}
