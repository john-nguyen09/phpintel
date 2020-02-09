package storage

import "encoding/binary"

import "github.com/john-nguyen09/phpintel/internal/lsp/protocol"

func PutUInt64(dst []byte, v uint64) []byte {
	return append(dst, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func ReadUInt64(src []byte) uint64 {
	return binary.BigEndian.Uint64(src)
}

type coder struct {
	buf    []byte
	offset int
}

type Encoder coder

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) WriteInt(v int) {
	e.buf = PutUInt64(e.buf, uint64(v))
}

func (e *Encoder) WriteBytes(b []byte) {
	e.WriteInt(len(b))
	e.buf = append(e.buf, b...)
}

func (e *Encoder) WriteString(v string) {
	e.WriteBytes([]byte(v))
}

func (e *Encoder) WriteBool(b bool) {
	by := byte(0)
	if b {
		by = byte(1)
	}
	e.buf = append(e.buf, by)
}

func (e *Encoder) WritePosition(v protocol.Position) {
	e.WriteInt(v.Line)
	e.WriteInt(v.Character)
}

func (e *Encoder) WriteLocation(v protocol.Location) {
	e.WriteString(v.URI)
	e.WritePosition(v.Range.Start)
	e.WritePosition(v.Range.End)
}

func (e *Encoder) Bytes() []byte {
	return e.buf
}

type Decoder coder

func NewDecoder(b []byte) *Decoder {
	return &Decoder{b, 0}
}

func (d *Decoder) ReadInt() int {
	u := ReadUInt64(d.buf)
	d.buf = d.buf[8:]
	return int(u)
}

func (d *Decoder) ReadBytes() []byte {
	len := d.ReadInt()
	b := append(d.buf[:0:0], d.buf[:len]...)
	d.buf = d.buf[len:]
	return b
}

func (d *Decoder) ReadString() string {
	return string(d.ReadBytes())
}

func (d *Decoder) ReadBool() bool {
	b := d.buf[0]
	d.buf = d.buf[1:]
	return b != 0
}

func (d *Decoder) ReadPosition() protocol.Position {
	return protocol.Position{
		Line:      d.ReadInt(),
		Character: d.ReadInt(),
	}
}

func (d *Decoder) ReadLocation() protocol.Location {
	return protocol.Location{
		URI: d.ReadString(),
		Range: protocol.Range{
			Start: d.ReadPosition(),
			End:   d.ReadPosition(),
		},
	}
}
