package filter

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	cuckoo "github.com/seiflotfy/cuckoofilter"
)

// Filter is a wrapper around cuckoo filter
type Filter struct {
	mutex  sync.RWMutex
	head   *cuckoo.Filter
	buffer [][]byte
}

// NewFilter creates the Filter
func NewFilter() *Filter {
	return &Filter{
		head:   nil,
		buffer: [][]byte{},
	}
}

// Insert inserts data into a buffer but not yet available to be used
func (f *Filter) Insert(data []byte) *Filter {
	f.buffer = append(f.buffer, data)
	return f
}

// Commit commits the buffer into a cuckoo filter
func (f *Filter) Commit() error {
	keys := f.dataWithoutDup()
	f.buffer = [][]byte{}
	filter := cuckoo.NewFilter(uint(len(keys)))
	for _, key := range keys {
		filter.Insert(key)
	}
	f.mutex.Lock()
	f.head = filter
	f.mutex.Unlock()
	return nil
}

func (f *Filter) Lookup(data []byte) (bool, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	if f.head == nil {
		return false, fmt.Errorf("filter is not yet commited")
	}
	return f.head.Lookup(data), nil
}

// Encode encodes the filter into byte slice
func (f *Filter) Encode(e *storage.Encoder) {
	buffer := f.head.Encode()
	e.WriteBytes(buffer)
}

// FilterDecode decodes a filter from a byte slice
func FilterDecode(d *storage.Decoder) *Filter {
	head, err := cuckoo.Decode(d.ReadBytes())
	if err != nil {
		panic(err)
	}
	f := NewFilter()
	f.head = head
	return f
}

func (f *Filter) dataWithoutDup() [][]byte {
	in := f.buffer
	if len(in) == 0 {
		return in
	}
	sortByteArrays(in)
	j := 0
	for i := 1; i < len(in); i++ {
		if bytes.Equal(in[j], in[i]) {
			continue
		}
		j++
		in[j] = in[i]
	}
	return in[:j+1]
}

func sortByteArrays(src [][]byte) {
	sort.Slice(src, func(i, j int) bool {
		return bytes.Compare(src[i], src[j]) < 0
	})
}
