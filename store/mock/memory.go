package mock

//Store interface definition of meta-data storage
/*type Store interface {
	get() (string, error)
	set(string, string) error
	getStruct() (interface{}, error)
	setStruct(string, interface{}) error
}
*/

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	setFail = "fail-set"
	getFail = "fail-get"
	delFail = "fail-del"
	genFail = "fail-sto"
)

var (
	errKeyNotFound = errors.New("key not found")
	errSetFailed   = errors.New("failed to set key")
	errGetFailed   = errors.New("failed to get key")
	errDelFailed   = errors.New("failed to delete key")
)

//MemStore - struct for in memory store
type MemStore struct {
	dict sync.Map
}

// NewStore - to create the redis connection
func NewStore(address, port string) *MemStore {
	return &MemStore{
		dict: sync.Map{},
	}

}

//Set ...
func (m *MemStore) Set(key, value string) error {
	if fail := strings.Contains(key, setFail); fail {
		return errSetFailed
	}

	m.dict.Store(key, value)
	return nil
}

//Get ...
func (m *MemStore) Get(key string) (string, error) {
	if fail := strings.Contains(key, getFail); fail {
		return "", errGetFailed
	}

	val, ok := m.dict.Load(key)
	if !ok {
		return "", errKeyNotFound
	}

	s := fmt.Sprintf("%s", val)
	return s, nil
}

//GetStruct ...
func (m *MemStore) GetStruct(key string) (string, error) {
	if fail := strings.Contains(key, getFail); fail {
		return "", errGetFailed
	}

	reply, ok := m.dict.Load(key)
	if !ok {
		return "", errKeyNotFound
	}

	var val string
	switch reply := reply.(type) {
	case []byte:
		val = string(reply)
	case string:
		val = reply
	}
	return val, nil
}

//SetStruct ...
func (m *MemStore) SetStruct(key string, value string) error {
	if fail := strings.Contains(key, setFail); fail {
		return errSetFailed
	}

	// SET object
	m.dict.Store(key, value)

	return nil
}

//DeleteKey ...
func (m *MemStore) DeleteKey(key string) error {
	if fail := strings.Contains(key, delFail); fail {
		return errDelFailed
	}

	_, ok := m.dict.Load(key)
	if !ok {
		return errKeyNotFound
	}

	m.dict.Delete(key)
	return nil
}

//KeyExists ...
func (m *MemStore) KeyExists(key string) (int, error) {
	if fail := strings.Contains(key, genFail); fail {
		return 0, errKeyNotFound
	}

	_, ok := m.dict.Load(key)
	if ok {
		return 1, nil
	}

	return 0, nil
}

//SetExpireTime ...
func (m *MemStore) SetExpireTime(key string, timeout int) error {
	if fail := strings.Contains(key, genFail); fail {
		return errSetFailed
	}

	_, ok := m.dict.Load(key)
	if !ok {
		return errKeyNotFound
	}

	return nil
}

//GetKeys list of keys via pattern matching
func (m *MemStore) GetKeys(pattern string) ([]string, error) {
	if fail := strings.Contains(pattern, genFail); fail {
		return nil, errGetFailed
	}

	var keys []string

	m.dict.Range(func(key, val interface{}) bool {
		if strVal, ok := val.(string); !ok {
			return true
		} else if strings.Contains(strVal, pattern) {
			keys = append(keys, strVal)
		}
		return true
	})

	return keys, nil
}

//GetStructFromHash - get val(value) using key(secondaryKey) present inside a hash (primaryKey)
func (m *MemStore) GetStructFromHash(primaryKey, secondaryKey string) (string, error) {
	if fail := strings.Contains(primaryKey+secondaryKey, getFail); fail {
		return "", errGetFailed
	}

	reply, ok := m.dict.Load(primaryKey)
	if !ok {
		return "", errKeyNotFound
	}

	hset, ok := reply.(map[string]interface{})
	if !ok {
		return "", errors.New("parser error in GetStructFromHash")
	}

	reply, ok = hset[secondaryKey]
	if !ok {
		return "", errKeyNotFound
	}

	var val string
	switch reply := reply.(type) {
	case []byte:
		val = string(reply)
	case string:
		val = reply
	}
	return val, nil
}

//SetStructInHash - set key(secondaryKey) val(value) inside a hash (primaryKey)
func (m *MemStore) SetStructInHash(primaryKey, secondaryKey string, value string) error {
	if fail := strings.Contains(primaryKey+secondaryKey, setFail); fail {
		return errSetFailed
	}

	recMap := make(map[string]interface{})
	if val, ok := m.dict.Load(primaryKey); ok {
		recMap, ok = val.(map[string]interface{})
		if !ok {
			return errors.New("parsing failed at SetStructInHash")
		}
	}

	// SET object
	recMap[secondaryKey] = value
	m.dict.Store(primaryKey, recMap)

	return nil
}

//GetKeysFromHash get list of keys inside a hash
func (m *MemStore) GetKeysFromHash(uuid string) ([]string, error) {
	if fail := strings.Contains(uuid, getFail); fail {
		return nil, errGetFailed
	}

	val, ok := m.dict.Load(uuid)
	if !ok {
		return nil, nil
	}

	hashMap, ok := val.(map[string]interface{})
	if !ok {
		return nil, errors.New("parsing failed at GetKeysFromHash")
	}

	var keys []string
	for k, _ := range hashMap {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *MemStore) DeleteStructFromHash(primaryKey, secondaryKey string) error {
	if fail := strings.Contains(primaryKey+secondaryKey, setFail); fail {
		return errDelFailed
	}

	recMap := make(map[string]interface{})
	if val, ok := m.dict.Load(primaryKey); ok {
		recMap, ok = val.(map[string]interface{})
		if !ok {
			return errors.New("parsing failed at DeleteStructFromHash")
		}
	}

	// SET object
	delete(recMap, secondaryKey)
	m.dict.Store(primaryKey, recMap)

	return nil
}

func (m *MemStore) KeyExistsInHash(primaryKey, secondaryKey string) (int, error) {
	if fail := strings.Contains(primaryKey, getFail); fail {
		return -1, errGetFailed
	}

	val, ok := m.dict.Load(primaryKey)
	if !ok {
		return 0, nil
	}

	hashMap, ok := val.(map[string]interface{})
	if !ok {
		return 0, errors.New("parsing failed at KeyExistsInHash")
	}

	_, ok = hashMap[secondaryKey]
	if ok {
		return 1, nil
	}

	return 0, nil
}

func (m *MemStore) AtomicIncrement(key string) error {
	return nil
}

func (m *MemStore) SetMultiStructInHash(primaryKey string, keyVal map[string]string) error {
	if fail := strings.Contains(primaryKey, setFail); fail {
		return errSetFailed
	}

	multiMap := make(map[string]interface{})
	if val, ok := m.dict.Load(primaryKey); ok {
		multiMap, ok = val.(map[string]interface{})
		if !ok {
			return errors.New("parsing failed at SetMultiStructInHash")
		}
	}

	for key, val := range keyVal {
		multiMap[key] = val
	}

	m.dict.Store(primaryKey, multiMap)

	return nil
}

func (m *MemStore) DelMultiKeyFromHash(primaryKey string, delKeys []interface{}) error {
	if fail := strings.Contains(primaryKey, setFail); fail {
		return errDelFailed
	}

	multiMap := make(map[string]interface{})
	if val, ok := m.dict.Load(primaryKey); ok {
		multiMap, ok = val.(map[string]interface{})
		if !ok {
			return errors.New("parsing failed at DeleteStructFromHash")
		}
	}

	for _, val := range delKeys {
		delete(multiMap, val.(string))
	}
	m.dict.Store(primaryKey, multiMap)

	return nil
}

//GetHashKeyCount ...
func (m *MemStore) GetHashKeyCount(key string) (int, error) {
	if fail := strings.Contains(key, setFail); fail {
		return 0, errGetFailed
	}

	multiMap := make(map[string]interface{})
	if val, ok := m.dict.Load(key); ok {
		multiMap, ok = val.(map[string]interface{})
		if !ok {
			return 0, errors.New("parsing failed at DeleteStructFromHash")
		}
	}

	return len(multiMap), nil
}

// QueuePush Push multiple items in the queue
func (m *MemStore) QueuePush(key string, data ...string) error {
	if fail := strings.Contains(key, setFail); fail {
		return errGetFailed
	}

	dataArr := []string{}
	if val, ok := m.dict.Load(key); ok {
		dataArr, ok = val.([]string)
		if !ok {
			dataArr = []string{}
		}
	}
	dataArr = append(dataArr, data...)
	m.dict.Store(key, dataArr)
	return nil
}

// QueuePop Pop item which is at the top of the queue
func (m *MemStore) QueuePop(key string) (string, error){
	if fail := strings.Contains(key, getFail); fail {
		return "", errGetFailed
	}
	dataArr := []string{}
	if val, ok := m.dict.Load(key); ok {
		dataArr, ok = val.([]string)
		if !ok {
			return "" , fmt.Errorf("no key found")
		}
	}
	if len(dataArr) == 0{
		return "", fmt.Errorf("queue is empty")
	}
	dataToReturn := dataArr[0]
	dataArr = dataArr[1:]
	m.dict.Store(key, dataArr)
	return dataToReturn, nil
}

// QueuePeak Peak the item which is at the top of the queue
func (m *MemStore) QueuePeak(key string) (string, error){
	if fail := strings.Contains(key, getFail); fail {
		return "", errGetFailed
	}
	dataArr := []string{}
	if val, ok := m.dict.Load(key); ok {
		dataArr, ok = val.([]string)
		if !ok {
			return "" , fmt.Errorf("no key found")
		}
	}
	if len(dataArr) == 0{
		return "", fmt.Errorf("queue is empty")
	}
	return dataArr[0], nil

}

// QueuePeakIndex Read an item from the specific index in the queue
func (m *MemStore) QueuePeakIndex(key string, index int32) (string, error){
	if fail := strings.Contains(key, getFail); fail {
		return "", errGetFailed
	}

	dataArr := []string{}
	if val, ok := m.dict.Load(key); ok {
		dataArr, ok = val.([]string)
		if !ok {
			return "" , fmt.Errorf("no key found")
		}
	}
	if len(dataArr) == 0{
		return "", fmt.Errorf("queue is empty")
	}
	if int(index) < len(dataArr){
		return dataArr[index], nil
	}else{
		return "", fmt.Errorf("index out of range")
	}
}
