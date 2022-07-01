package redis

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

func (c *Connection) DoublePush(ctlq, q string, data []byte) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Send("MULTI")
	if noErrNil(err) != nil {
		return err
	}

	//1st push event data
	err = conn.Send("LPUSH", q, data)
	if noErrNil(err) != nil {
		conn.Send("DISCARD")
		return err
	}

	//then control data
	queuebyte := []byte(q)
	err = conn.Send("LPUSH", ctlq, queuebyte)
	if noErrNil(err) != nil {
		conn.Send("DISCARD")
		return err
	}

	err = conn.Send("EXEC")
	return noErrNil(err)
}

func (c *Connection) SimplePush(q string, data []byte) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("LPUSH", q, data)
	return noErrNil(err)
}

func (c *Connection) BPopQ(ctrlq string, timeout int) ([]byte, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	values, err := redis.Values(conn.Do("BRPOP", ctrlq, timeout))
	if err != nil {
		return nil, err
	}

	if values != nil && len(values) > 0 {
		switch reply := values[1].(type) {
		case []byte:
			return reply, nil
		case string:
			return []byte(reply), nil
		case nil:
			return nil, errors.New("bpop returned nil value")
		}
	}

	return nil, errors.New("bpop returned nil")
}

func (c *Connection) PeekQ(q string) ([]byte, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	values, err := redis.Values(conn.Do("LRANGE", q, -1, -1))
	if err != nil {
		return nil, err
	}

	if values != nil && len(values) > 0 {
		switch reply := values[0].(type) {
		case []byte:
			return reply, nil
		case string:
			return []byte(reply), nil
		case nil:
			return nil, errors.New("peek returned nil value")
		}
	}

	return nil, errors.New("peek returned nil")
}

func (c *Connection) PopQ(q string) ([]byte, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("RPOP", q))
	if err != nil {
		return nil, err
	}

	return data, err
}

func (c *Connection) PopAndMoveQ(srcQ, destQ string, timeout int) ([]byte, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("BRPOPLPUSH", srcQ, destQ, timeout))
	if err != nil {
		return nil, err
	}

	return data, err
}

func (c *Connection) RemoveItem(q string, item []byte) (int, error) {
	conn, err := c.getConnFromPool()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	cnt, err := redis.Int(conn.Do("LREM", q, 0, item))
	return cnt, err
}

func (c *Connection) LockMsg(q, lockId string, ex int) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	/*
	 * Note:
	 * This works for single Redis master. For multiple
	 * Redis masters, use Redlock algorithm described
	 * at https://redis.io/topics/distlock.
	 */
	reply, err := redis.String(conn.Do("SET", q, lockId, "NX", "EX", ex))
	return negReplyErr(reply, err)
}

func (c *Connection) UnlockMsg(q, lockId string) error {
	conn, err := c.getConnFromPool()
	if err != nil {
		return err
	}
	defer conn.Close()

	var unlockScript = redis.NewScript(1, `
		if redis.call("get",KEYS[1]) == ARGV[1]
		then
			return redis.call("del",KEYS[1])
		else
			return 0
		end
	`)

	_, err = unlockScript.Do(conn, q, lockId)
	return noErrNil(err)
}

func noErrNil(err error) error {
	if err == redis.ErrNil {
		return nil
	} else {
		return err
	}
}

func negReplyErr(reply string, err error) error {
	if err != nil {
		return err
	} else {
		switch reply {
		case "OK":
			return nil
		default:
			return errors.New("Failed to acquire lock: " + reply)
		}
	}
}
