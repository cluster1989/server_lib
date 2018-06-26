package libredis

import (
	"github.com/garyburd/redigo/redis"
)

// redis hmset
func (r *RedisPool) SAdd(groupName string, args ...interface{}) error {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, args...)

	if _, err := r.DoRedis("SADD", rArgs...); err != nil {
		return err
	}
	return nil
}

// redis hmset
func (r *RedisPool) SRem(groupName string, args ...interface{}) error {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, args...)

	if _, err := r.DoRedis("SREM", rArgs...); err != nil {
		return err
	}
	return nil
}

// redis hmset
func (r *RedisPool) SMembers(groupName string) ([]string, error) {

	reply, err := r.DoRedis("SMEMBERS", groupName)
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
