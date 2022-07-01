package redis

import (
	"github.com/gomodule/redigo/redis"
)

//Set key
func (c *Connection) Set(key, value string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("SET", key, value)
	return err
}

//Get key
func (c *Connection) Get(key string) (string, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	s, err := redis.String(conn.Do("GET", key))
	return s, err
}

//GetStruct - get struct
func (c *Connection) GetStruct(key string) (string, error) {

	conn, err := c.getConnFromPool()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	data, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}

	return data, err
}

//SetStruct - set struct
func (c *Connection) SetStruct(key string, value string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	// SETEX save data with expire time
	_, err = conn.Do("SETEX", key, c.expireTime, value)
	return err
}

//DeleteKey ...
func (c *Connection) DeleteKey(key string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("DEL", key)
	return err
}

//KeyExists ...
func (c *Connection) KeyExists(key string) (int, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	val, err := redis.Int(conn.Do("EXISTS", key))
	return val, err
}

//SetExpireTime ...
func (c *Connection) SetExpireTime(key string, timeout int) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("EXPIRE", key, timeout)
	return err
}

//GetKeys list of keys via pattern matching
func (c *Connection) GetKeys(pattern string) ([]string, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	pattern += "*"
	if keys, err := redis.Strings(conn.Do("KEYS", pattern)); err != nil {
		return nil, err
	} else {
		return keys, nil
	}
}

//GetHash get val(value) using key(secondaryKey) present inside a hash (primaryKey)
func (c *Connection) GetStructFromHash(primaryKey, secondaryKey string) (string, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	data, err := redis.String(conn.Do("HGET", primaryKey, secondaryKey))
	if err != nil {
		return "", err
	}

	return data, err
}

//SetHash set key(secondaryKey) val(value) inside a hash (primaryKey)
func (c *Connection) SetStructInHash(primaryKey, secondaryKey string, value string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err = conn.Do("HSET", primaryKey, secondaryKey, value); err != nil {
		return err
	}
	_, err = conn.Do("EXPIRE", primaryKey, c.expireTime)
	return err
}

//GetKeysFromHash get list of keys inside a hash
func (c *Connection) GetKeysFromHash(uuid string) ([]string, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if keys, err := redis.Strings(conn.Do("HKEYS", uuid)); err != nil {
		return nil, err
	} else {
		return keys, nil
	}
}

//DeleteStructFromHash delete one or more keys from set
func (c *Connection) DeleteStructFromHash(primaryKey, secondaryKey string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("HDEL", primaryKey, secondaryKey)
	return err
}

//KeyExistsInHash ...
func (c *Connection) KeyExistsInHash(primaryKey, secondaryKey string) (int, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	val, err := redis.Int(conn.Do("HEXISTS", primaryKey, secondaryKey))
	return val, err
}

// Atomic Operation on a value
func (c *Connection) AtomicIncrement(key string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Do("INCR", key)
	return err
}

func (c *Connection) SetMultiStructInHash(primaryKey string, keyVal map[string]string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	cmdArgs := []interface{}{primaryKey}
	for key, val := range keyVal {
		cmdArgs = append(cmdArgs, key, val)
	}

	if _, err := conn.Do("HSET", cmdArgs...); err != nil {
		return err
	}
	_, err = conn.Do("EXPIRE", primaryKey, c.expireTime)
	return err
}

func (c *Connection) DelMultiKeyFromHash(primaryKey string, delKeys []interface{}) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	delKeys = append([]interface{}{primaryKey}, delKeys...)
	if _, err := conn.Do("HDEL", delKeys...); err != nil {
		return err
	}
	return nil
}

//GetHashKeyCount ...
func (c *Connection) GetHashKeyCount(key string) (int, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	val, err := redis.Int(conn.Do("HLEN", key))
	return val, err
}

// QueuePush Push multiple items in the queue
func (c *Connection) QueuePush(key string, data ...string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	cmdData := []interface{}{key}
	for _, v := range data {
		cmdData = append(cmdData, v)
	}

	_, err = conn.Do("RPUSH", cmdData...)
	return noErrNil(err)
}

// QueuePop Pop item which is at the top of the queue
func (c *Connection) QueuePop(key string) (string, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	data, err := redis.String(conn.Do("LPOP", key))
	if err != nil {
		return "", err
	}

	return data, err
}

// QueuePeak Peak the item which is at the top of the queue
func (c *Connection) QueuePeak(key string) (string, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	data, err := redis.String(conn.Do("LINDEX", key, 0))
	if err != nil {
		return "", err
	}

	return data, err
}

// QueuePeakIndex Read an item from the specific index in the queue
func (c *Connection) QueuePeakIndex(key string, index int32) (string, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	data, err := redis.String(conn.Do("LINDEX", key, index))
	if err != nil {
		return "", err
	}

	return data, err
}

func (c *Connection) AddSortedSet(key string, score int, data string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("ZADD", key, score, data)
	return noErrNil(err)
}

func (c *Connection) RemoveSortedSet(key string, data string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("ZREM", key, data)
	return noErrNil(err)
}

func (c *Connection) GetRankSortedSet(key string, data string) (int, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	rank, err := redis.Int(conn.Do("ZRANK", key, data))
	if err != nil {
		return 0, err
	}
	return rank, nil
}

func (c *Connection) GetAllItemsSortedSet(key string) ([]string, error) {
	var items []string
	conn, err := c.getConnFromPool()
	if err != nil {
		return items, err
	}
	defer conn.Close()

	items, err = redis.Strings(conn.Do("ZRANGE", key, 0, -1))
	if err != nil {
		return items, err
	}
	return items, nil
}
