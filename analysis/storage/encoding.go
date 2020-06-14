package storage

import (
	"encoding/binary"
	"math"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// PutUInt64 appends an uint64 to the byte slice
func PutUInt64(dst []byte, v uint64) []byte {
	return append(dst, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// PutUInt32 appends an uint32 to the byte slice
func PutUInt32(dst []byte, v uint32) []byte {
	return append(dst, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// PutFloat64 appends a float64 to the byte slice
func PutFloat64(dst []byte, v float64) []byte {
	return PutUInt64(dst, math.Float64bits(v))
}

// PutFloat32 appens a float32 to the byte slice
func PutFloat32(dst []byte, v float32) []byte {
	return PutUInt32(dst, math.Float32bits(v))
}

// ReadUInt64 reads an uint64 from the byte slice
func ReadUInt64(src []byte) uint64 {
	return binary.BigEndian.Uint64(src)
}

// ReadUInt32 reads an uint32 from the byte slice
func ReadUInt32(src []byte) uint32 {
	return binary.BigEndian.Uint32(src)
}

// ReadFloat64 reads a float64 from the byte slice
func ReadFloat64(src []byte) float64 {
	u := ReadUInt64(src)
	return math.Float64frombits(u)
}

// ReadFloat32 reads a float32 from the byte slice
func ReadFloat32(src []byte) float32 {
	u := ReadUInt32(src)
	return math.Float32frombits(u)
}

type coder struct {
	buf    []byte
	offset int
}

// Encoder is an encoder to encode primitives to byte slice
type Encoder coder

// NewEncoder creates an encoder
func NewEncoder() *Encoder {
	return &Encoder{}
}

// WriteUInt64 writes an uint64
func (e *Encoder) WriteUInt64(v uint64) {
	e.buf = PutUInt64(e.buf, v)
}

// WriteInt writes an int
func (e *Encoder) WriteInt(v int) {
	e.WriteUInt64(uint64(v))
}

// WriteUInt32 writes an uint32
func (e *Encoder) WriteUInt32(v uint32) {
	e.buf = PutUInt32(e.buf, v)
}

// WriteFloat64 writes a float64
func (e *Encoder) WriteFloat64(v float64) {
	e.buf = PutFloat64(e.buf, v)
}

// WriteFloat32 writes a float32
func (e *Encoder) WriteFloat32(v float32) {
	e.buf = PutFloat32(e.buf, v)
}

// WriteBytes writes bytes
func (e *Encoder) WriteBytes(b []byte) {
	e.WriteInt(len(b))
	e.buf = append(e.buf, b...)
}

// WriteString writes string
func (e *Encoder) WriteString(v string) {
	e.WriteBytes([]byte(v))
}

// WriteBool writes bool
func (e *Encoder) WriteBool(b bool) {
	by := byte(0)
	if b {
		by = byte(1)
	}
	e.buf = append(e.buf, by)
}

// WritePosition writes a LSP position
func (e *Encoder) WritePosition(v protocol.Position) {
	e.WriteInt(v.Line)
	e.WriteInt(v.Character)
}

// WriteLocation writes a LSP location
func (e *Encoder) WriteLocation(v protocol.Location) {
	e.WriteString(v.URI)
	e.WritePosition(v.Range.Start)
	e.WritePosition(v.Range.End)
}

// Bytes returns the underlying byte slice
func (e *Encoder) Bytes() []byte {
	return e.buf
}

// Decoder is a coder to decode byte slice into primitives
type Decoder coder

// NewDecoder creates a decoder
func NewDecoder(b []byte) *Decoder {
	return &Decoder{b, 0}
}

// ReadInt reads an int
func (d *Decoder) ReadInt() int {
	return int(d.ReadUInt64())
}

// ReadUInt64 reads an uint64
func (d *Decoder) ReadUInt64() uint64 {
	u := ReadUInt64(d.buf)
	d.buf = d.buf[8:]
	return u
}

// ReadUInt32 reads an uint32
func (d *Decoder) ReadUInt32() uint32 {
	u := ReadUInt32(d.buf)
	d.buf = d.buf[4:]
	return u
}

// ReadFloat64 reads a float64
func (d *Decoder) ReadFloat64() float64 {
	f := ReadFloat64(d.buf)
	d.buf = d.buf[8:]
	return f
}

// ReadFloat32 reads a float32
func (d *Decoder) ReadFloat32() float32 {
	f := ReadFloat32(d.buf)
	d.buf = d.buf[4:]
	return f
}

// ReadBytes reads bytes
func (d *Decoder) ReadBytes() []byte {
	len := d.ReadInt()
	b := append(d.buf[:0:0], d.buf[:len]...)
	d.buf = d.buf[len:]
	return b
}

// ReadString reads a string
func (d *Decoder) ReadString() string {
	return string(d.ReadBytes())
}

// ReadBool reads a bool
func (d *Decoder) ReadBool() bool {
	b := d.buf[0]
	d.buf = d.buf[1:]
	return b != 0
}

// ReadPosition reads a LSP position
func (d *Decoder) ReadPosition() protocol.Position {
	return protocol.Position{
		Line:      d.ReadInt(),
		Character: d.ReadInt(),
	}
}

// ReadLocation reads a LSP location
func (d *Decoder) ReadLocation() protocol.Location {
	return protocol.Location{
		URI: d.ReadString(),
		Range: protocol.Range{
			Start: d.ReadPosition(),
			End:   d.ReadPosition(),
		},
	}
}
