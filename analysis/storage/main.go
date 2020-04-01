package storage

import "fmt"

type Iterator interface {
	valid() bool
	next()
	close()
	Key() []byte
	Value() []byte
	Stop()
}

type Batch interface {
	Delete([]byte)
	Put([]byte, []byte)
}

type DB interface {
	Close()
	Clear()
	Delete([]byte) error
	Get([]byte) ([]byte, error)
	PrefixStream([]byte, func(Iterator))
	Put([]byte, []byte) error
	WriteBatch(func(Batch) error) error
}

// Combined is a combination of storage modes
type Combined struct {
	dbs []DB
}

// DBMode is the mode of the DB
type DBMode int

const (
	// ModeMemory indicates memory
	ModeMemory = iota
	// ModeDisk indicates disk
	ModeDisk DBMode = iota
)

// NewCombined returns a combined instance of disk and memory
func NewCombined(path string) (*Combined, error) {
	disk, err := newDisk(path)
	if err != nil {
		return nil, err
	}
	return &Combined{
		// This index must match DBMode
		dbs: []DB{
			newMemory(),
			disk,
		},
	}, nil
}

func (c *Combined) Close() {
	c.GetMode(ModeMemory).Close()
}

func (c *Combined) GetMode(mode DBMode) DB {
	return c.dbs[mode]
}

func (c *Combined) Clear(mode DBMode) {
	c.GetMode(mode).Clear()
}

func (c *Combined) Delete(mode DBMode, key []byte) error {
	return c.GetMode(mode).Delete(key)
}

func (c *Combined) Get(mode DBMode, key []byte) ([]byte, error) {
	return c.GetMode(mode).Get(key)
}

func (c *Combined) GetFromAll(key []byte) ([]byte, error) {
	for _, db := range c.dbs {
		if value, err := db.Get(key); err == nil {
			return value, nil
		}
	}
	return nil, fmt.Errorf("Key not found")
}

func (c *Combined) PrefixStream(mode DBMode, prefix []byte, onData func(Iterator)) {
	c.GetMode(mode).PrefixStream(prefix, onData)
}

func (c *Combined) PrefixStreamAll(prefix []byte, onData func(Iterator)) {
	for _, db := range c.dbs {
		db.PrefixStream(prefix, onData)
	}
}

func (c *Combined) Put(mode DBMode, key []byte, value []byte) error {
	return c.GetMode(mode).Put(key, value)
}

func (c *Combined) WriteBatch(mode DBMode, f func(Batch) error) error {
	return c.GetMode(mode).WriteBatch(f)
}
