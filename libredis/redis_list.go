package libredis

import "github.com/garyburd/redigo/redis"

// redis  队列
func (r *RedisPool) LPUSH(groupName string, args ...interface{}) error {

	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, args...)

	if _, err := r.DoRedis("LPUSH", rArgs...); err != nil {
		return err
	}
	return nil
}

func (r *RedisPool) LRange(groupName string, page, num int) ([]string, error) {
	rArgs := make(redis.Args, 0)
	rArgs = append(rArgs, groupName)
	rArgs = append(rArgs, page*num)
	rArgs = append(rArgs, (page+1)*num-1)

	reply, err := r.DoRedis("LRANGE", rArgs...)
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
