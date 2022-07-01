package store

//Store interface definition of meta-data storage
type Store interface {
	Get(string) (string, error)
	Set(string, string) error
	GetStruct(string) (string, error)
	SetStruct(string, string) error
	DeleteKey(string) error
	KeyExists(string) (int, error)
	SetExpireTime(string, int) error
	GetKeys(string) ([]string, error)
	GetHashKeyCount(string) (int, error)
	GetStructFromHash(string, string) (string, error)
	SetStructInHash(string, string, string) error
	GetKeysFromHash(string) ([]string, error)
	DeleteStructFromHash(string, string) error
	KeyExistsInHash(string, string) (int, error)
	AtomicIncrement(key string) error
	SetMultiStructInHash(string, map[string]string) error
	DelMultiKeyFromHash(string, []interface{}) error
	QueuePush(string, ...string) error
	QueuePop(string) (string, error)
	QueuePeak(string) (string, error)
	QueuePeakIndex(string, int32) (string, error)
	AddSortedSet(key string, score int, data string) error
	RemoveSortedSet(key string, data string) error
	GetRankSortedSet(key string, data string) (int, error)
	GetAllItemsSortedSet(key string) ([]string, error)
}

//Simple queue interface definition
type Queue interface {
	DoublePush(ctrlQ, dataQ string, data []byte) error
	SimplePush(qname string, data []byte) error
	BPopQ(qname string, timeout int) ([]byte, error)
	PopQ(qname string) ([]byte, error)
	PeekQ(qname string) ([]byte, error)
	PopAndMoveQ(srcQ, destQ string, timeout int) ([]byte, error)
	RemoveItem(qname string, data []byte) (int, error)
	LockMsg(key, lockId string, expires int) error
	UnlockMsg(key, lockId string) error
}
