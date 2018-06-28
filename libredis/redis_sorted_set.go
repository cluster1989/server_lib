package libredis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/wuqifei/server_lib/libio"
)

// redis zadd
func (r *RedisPool) ZADD(groupName string, rank int, val string) error {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, rank)
	rArgs = append(rArgs, val)

	if _, err := r.DoRedis("ZADD", rArgs...); err != nil {
		return err
	}
	return nil
}

// redis zadd
func (r *RedisPool) ZREMByMember(groupName string, member ...interface{}) error {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, member...)

	if _, err := r.DoRedis("ZREM", rArgs...); err != nil {
		return err
	}
	return nil
}

// redis zcard
func (r *RedisPool) ZCARD(groupName string) (int64, error) {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	reply, err := r.DoRedis("ZCARD", rArgs...)
	if err != nil {
		return 0, err
	}

	return reply.(int64), nil
}

// redis zcard
func (r *RedisPool) ZRangeByScoreCARD(groupName string, min, max int) ([]string, error) {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, min)
	rArgs = append(rArgs, max)
	reply, err := r.DoRedis("ZRANGEBYSCORE", rArgs...)
	if err != nil {
		return nil, err
	}

	arr := reply.([]interface{})
	m := make([]string, 0)
	for i := 0; i < len(arr); i++ {
		key := arr[i].([]uint8)
		keyByte := []byte(key)
		m = append(m, string(keyByte))
	}
	return m, nil
}

func (r *RedisPool) ZRANGE(groupName string, start, end int) ([]string, error) {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, start)
	rArgs = append(rArgs, end)
	reply, err := r.DoRedis("ZRANGE", rArgs...)
	if err != nil {
		return nil, err
	}

	arr := reply.([]interface{})
	m := make([]string, 0)
	for i := 0; i < len(arr); i++ {
		key := arr[i].([]uint8)
		keyByte := []byte(key)
		m = append(m, string(keyByte))
	}
	return m, nil
}

func (r *RedisPool) ZREVRANGE(groupName string, start, end int) ([]string, error) {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, start)
	rArgs = append(rArgs, end)
	reply, err := r.DoRedis("ZREVRANGE", rArgs...)
	if err != nil {
		return nil, err
	}

	arr := reply.([]interface{})
	m := make([]string, 0)
	for i := 0; i < len(arr); i++ {
		key := arr[i].([]uint8)
		keyByte := []byte(key)
		m = append(m, string(keyByte))
	}
	return m, nil
} //ZREVRANGE

func (r *RedisPool) ZSCORE(groupName, member string) (int, error) {
	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, member)
	reply, err := r.DoRedis("ZSCORE", rArgs...)
	if err != nil {
		return 0, err
	}

	if reply == nil {
		return 0, nil
	}
	val := reply.([]uint8)
	valStr := string([]byte(val))

	convert := libio.NewConvert(valStr)
	return convert.Int()
}
