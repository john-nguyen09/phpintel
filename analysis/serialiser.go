package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

type Serialiser struct {
	index int
	buf   []byte
}

type serialisable interface {
	GetCollection() string
	GetKey() string
	Serialise(serialiser *Serialiser)
}

const uIntSize = 32 << (^uint(0) >> 32 & 1)

// int is same size as uint
const intSize = uIntSize

func NewSerialiser() *Serialiser {
	return &Serialiser{
		index: 0,
		buf:   []byte{},
	}
}

func SerialiserFromByteSlice(theBytes []byte) *Serialiser {
	return &Serialiser{
		index: 0,
		buf:   theBytes,
	}
}

func (s *Serialiser) WriteInt(number int) {
	if uIntSize == 32 {
		s.WriteInt32(int32(number))
	} else if uIntSize == 64 {
		s.WriteIn64(int64(number))
	}
}

func (s *Serialiser) ReadInt() int {
	if uIntSize == 32 {
		return int(s.ReadInt32())
	} else if uIntSize == 64 {
		return int(s.ReadInt64())
	}
	return 0
}

func (s *Serialiser) WriteInt32(number int32) {
	buf := make([]byte, 4)
	buf[0] = byte(number)
	buf[1] = byte(number >> 8)
	buf[2] = byte(number >> 16)
	buf[3] = byte(number >> 24)
	s.buf = append(s.buf, buf...)
}

func (s *Serialiser) ReadInt32() int32 {
	var number int32 = 0
	number |= int32(s.buf[s.index])
	number |= int32(s.buf[s.index+1]) << 8
	number |= int32(s.buf[s.index+2]) << 16
	number |= int32(s.buf[s.index+3]) << 24
	s.index += 4
	return number
}

func (s *Serialiser) WriteIn64(number int64) {
	buf := make([]byte, 8)
	buf[0] = byte(number)
	buf[1] = byte(number >> 8)
	buf[2] = byte(number >> 16)
	buf[3] = byte(number >> 24)
	buf[4] = byte(number >> 32)
	buf[5] = byte(number >> 40)
	buf[6] = byte(number >> 48)
	buf[7] = byte(number >> 56)
	s.buf = append(s.buf, buf...)
	s.index += 8
}

func (s *Serialiser) ReadInt64() int64 {
	var number int64 = 0
	number |= int64(s.buf[s.index])
	number |= int64(s.buf[s.index+1]) << 8
	number |= int64(s.buf[s.index+2]) << 16
	number |= int64(s.buf[s.index+3]) << 24
	number |= int64(s.buf[s.index+4]) << 32
	number |= int64(s.buf[s.index+5]) << 40
	number |= int64(s.buf[s.index+6]) << 48
	number |= int64(s.buf[s.index+7]) << 56
	s.index += 8
	return number
}

func (s *Serialiser) WriteString(theString string) {
	theBytes := []byte(theString)
	s.WriteBytes(theBytes)
}

func (s *Serialiser) ReadString() string {
	theBytes := s.ReadBytes()
	return string(theBytes)
}

func (s *Serialiser) WriteBool(theBool bool) {
	var theByte byte = 0
	if theBool {
		theByte = 1
	} else {
		theByte = 0
	}
	s.buf = append(s.buf, theByte)
}

func (s *Serialiser) ReadBool() bool {
	theByte := s.buf[s.index]
	s.index++
	if theByte == 1 {
		return true
	}
	return false
}

func (s *Serialiser) WriteLocation(location protocol.Location) {
	s.WriteString(string(location.URI))
	s.WritePosition(location.Range.Start)
	s.WritePosition(location.Range.End)
}

func (s *Serialiser) WritePosition(position protocol.Position) {
	s.WriteInt(position.Line)
	s.WriteInt(position.Character)
}

func (s *Serialiser) WriteBytes(theBytes []byte) {
	count := len(theBytes)
	s.WriteInt(count)
	s.buf = append(s.buf, theBytes...)
}

func (s *Serialiser) ReadBytes() []byte {
	count := s.ReadInt()
	theBytes := append(s.buf[:0:0], s.buf[s.index:s.index+count]...)
	s.index += count
	return theBytes
}

func (s *Serialiser) ReadLocation() protocol.Location {
	return protocol.Location{
		URI: protocol.DocumentURI(s.ReadString()),
		Range: protocol.Range{
			Start: s.ReadPosition(),
			End:   s.ReadPosition(),
		},
	}
}

func (s *Serialiser) ReadPosition() protocol.Position {
	return protocol.Position{
		Line:      s.ReadInt(),
		Character: s.ReadInt(),
	}
}

func (s *Serialiser) GetBytes() []byte {
	return s.buf
}
