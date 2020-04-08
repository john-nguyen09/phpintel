package storage

import (
	"log"
	"testing"
)

func TestIterator(t *testing.T) {
	mem := NewMemory()
	mem.Put([]byte("test1"), nil)
	mem.Put([]byte("test2"), nil)
	mem.Put([]byte("akjshdfkajsdf"), nil)

	mem.PrefixStream([]byte("test"), func(it Iterator) {
		log.Println(string(it.Key()))
	})
}
