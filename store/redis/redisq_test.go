package redis

import (
	"flag"
	"log"
	"os"
	"strconv"
	"testing"

	"cigol/pkg/logger/mocklogger"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var conn *Connection

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(0)
	}

	conf := viper.New()
	conf.Set("profile", "development")
	conf.Set("development.store.connectionStatus.receiveEventInSec", 10)
	conf.Set("development.store.connectionStatus.allowedEventMissCount", 3)
	conf.Set("development.store.retryMaxAttempt", 5)

	lgr := mocklogger.CreateLogger()
	conn = &Connection{
		address:           "127.0.0.1",
		port:              "6379",
		connectionChannel: make(chan bool),
	}

	logger = lgr

	err := conn.connect()
	if err != nil {
		log.Println("unable to connect: ", err)
		os.Exit(0)
	}

	os.Exit(m.Run())
}

func TestDoublePush(t *testing.T) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	ctlq := "ctlq" + uuid.String()
	q := "q" + uuid.String()
	data := []byte("random-data")

	err = conn.DoublePush(ctlq, q, data)
	assert.NoError(t, err)

	peekData, err := conn.PeekQ(q)
	assert.NoError(t, err)
	assert.Equal(t, peekData, data)

	popData, err := conn.PopQ(q)
	assert.NoError(t, err)
	assert.Equal(t, popData, data)

	qbyte := []byte(q)
	peekq, err := conn.PeekQ(ctlq)
	assert.NoError(t, err)
	assert.Equal(t, peekq, qbyte)

	popq, err := conn.PopQ(ctlq)
	assert.NoError(t, err)
	assert.Equal(t, popq, qbyte)
}

func TestPopAndMove(t *testing.T) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	ctlq := "ctlq" + uuid.String()
	q := "q" + uuid.String()
	data := []byte("random-data")

	err = conn.SimplePush(ctlq, data)
	assert.NoError(t, err)

	v, err := conn.PopAndMoveQ(ctlq, q, 2)
	assert.NoError(t, err)
	assert.Equal(t, v, data)

	v, err = conn.PopQ(q)
	assert.NoError(t, err)
	assert.Equal(t, v, data)

	v, err = conn.PeekQ(ctlq)
	assert.Error(t, err)

	v, err = conn.PeekQ(q)
	assert.Error(t, err)
}

func TestRemoveItem(t *testing.T) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	q := "q" + uuid.String()
	data := []byte("random-data")

	cnt := 5
	for i := 0; i < cnt; i++ {
		err = conn.SimplePush(q, data)
		assert.NoError(t, err)
	}

	c, err := conn.RemoveItem(q, data)
	assert.NoError(t, err)
	assert.Equal(t, c, cnt)

	c, err = conn.RemoveItem(q, data)
	assert.NoError(t, err)
	assert.Equal(t, c, 0)
}

func TestLockUnlock(t *testing.T) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	key := "Key" + uuid.String()
	lockid := "TestLockUnlock-LOCKID"

	err = conn.LockMsg(key, lockid, 90)
	assert.NoError(t, err)

	randomLockid := "random-string"
	err = conn.UnlockMsg(key, randomLockid)
	assert.NoError(t, err)

	err = conn.UnlockMsg(key, lockid)
	assert.NoError(t, err)
}

func TestMultiLock(t *testing.T) {
	type table struct {
		id int
		ch chan error
	}

	const numOfWorkers = 10
	tbl := [numOfWorkers]table{}

	uuid, err := uuid.NewUUID()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	key := "Key" + uuid.String()

	for i := 0; i < numOfWorkers; i++ {
		tbl[i] = table{
			id: i,
			ch: make(chan error),
		}
		go getLock(t, key, tbl[i].ch, tbl[i].id, 30)
	}

	k := 0
	acql := -1
	for ; k < numOfWorkers; k++ {
		if acql < 0 {
			err := <-tbl[k].ch
			t.Log("rcvd error:", err, "channel:", k)
			if err == nil {
				acql = k
			}
			continue
		}

		err := <-tbl[k].ch
		t.Log("rcvd error:", err, "channel:", k, "acqL:", acql)
		assert.Error(t, err)
	}

	assert.GreaterOrEqual(t, acql, 0)
}

func getLock(t *testing.T, key string, errch chan error, id, ex int) {
	lockid := "LOCKID" + strconv.Itoa(id)
	err := conn.LockMsg(key, lockid, 10)
	t.Log("channel:", id, "key:", key, "error:", err)

	errch <- err
}
