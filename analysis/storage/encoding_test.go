package storage

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

const MaxUint = ^uint(0)
const MinUint = 0
const MaxUint64 = ^uint64(0)
const MinUint64 = 0
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1
const MaxInt64 = int64(MaxUint64 >> 1)
const MinInt64 = -MaxInt - 1

func TestPutUInt64(t *testing.T) {
	cases := []uint64{
		100,
		1000,
		10000,
		238745687,
		2347852,
		12394871293,
		MaxUint64,
	}
	for _, i := range cases {
		decoded := ReadUInt64(PutUInt64([]byte{}, i))
		assert.Equal(t, i, decoded, "Failed to encode & decode %v, got %v", i, decoded)
	}
}

func TestPutInt64(t *testing.T) {
	log.Printf("Here")
	fmt.Println(PutUInt64([]byte{}, uint64(1)))
	cases := []int{
		1,
		100,
		1000,
		10000,
		238745687,
		2347852,
		12394871293,
		MaxInt,
	}
	for _, i := range cases {
		decoded := int(ReadUInt64(PutUInt64([]byte{}, uint64(i))))
		assert.Equal(t, i, decoded, "Failed to encode & decode %v, got %v", i, decoded)
	}

	value := 747239
	encoded1 := PutUInt64([]byte{}, uint64(value))
	decoded1 := int(ReadUInt64(encoded1))
	assert.Equal(t, value, decoded1, "Failed to encode & decode %v, got %v", value, decoded1)
	originalValue := value
	value = 19998
	// Make sure updating the value does not update the encoded1
	decoded1 = int(ReadUInt64(encoded1))
	assert.Equal(t, originalValue, decoded1, "Failed to encode & decode %v, got %v", value, decoded1)
	// And can encode & decode the new value
	decoded2 := int(ReadUInt64(PutUInt64([]byte{}, uint64(value))))
	assert.Equal(t, value, decoded2, "Failed to encode & decode %v, got %v", value, decoded2)
}

func BenchmarkEncodingInt(b *testing.B) {
	e := NewEncoder()

	for i := 0; i < b.N; i++ {
		e.WriteInt(1000)
	}
}

func BenchmarkEncodingString(b *testing.B) {
	e := NewEncoder()

	for i := 0; i < b.N; i++ {
		e.WriteString("Hello World")
	}
}
