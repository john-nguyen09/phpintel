package indexer

type Serialiser struct {
	index int
	buf   []byte
}

type Serialisable interface {
	Serialise() []byte
}

const uIntSize = 32 << (^uint(0) >> 32 & 1)
const bufferSize int = 1024 // Grow size by bufferSize bytes

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

func (s *Serialiser) needs(numberOfBytes int) {
	if cap(s.buf) < (s.index + numberOfBytes) {
		extraBuf := make([]byte, bufferSize, bufferSize)
		s.buf = append(s.buf, extraBuf...)
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
	s.needs(4)
	s.buf[s.index] = byte(number)
	s.buf[s.index+1] = byte(number >> 8)
	s.buf[s.index+2] = byte(number >> 16)
	s.buf[s.index+3] = byte(number >> 24)
	s.index += 4
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
	s.needs(8)
	s.buf[s.index] = byte(number)
	s.buf[s.index+1] = byte(number >> 8)
	s.buf[s.index+2] = byte(number >> 16)
	s.buf[s.index+3] = byte(number >> 24)
	s.buf[s.index+4] = byte(number >> 32)
	s.buf[s.index+5] = byte(number >> 40)
	s.buf[s.index+6] = byte(number >> 48)
	s.buf[s.index+7] = byte(number >> 56)
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
	count := len(theBytes)
	s.WriteInt(count)
	s.needs(count)
	copy(s.buf[s.index:], theBytes)
	s.index += count
}

func (s *Serialiser) ReadString() string {
	count := s.ReadInt()
	theBytes := append(s.buf[:0:0], s.buf[s.index:s.index+count]...)
	s.index += count
	return string(theBytes)
}

func (s *Serialiser) WriteBool(theBool bool) {
	s.needs(1)
	var theByte byte = 0
	if theBool {
		theByte = 1
	} else {
		theByte = 0
	}
	s.buf[s.index] = theByte
	s.index++
}

func (s *Serialiser) ReadBool() bool {
	theByte := s.buf[s.index]
	s.index++
	if theByte == 1 {
		return true
	} else {
		return false
	}
}

func (s *Serialiser) GetBytes() []byte {
	return s.buf[0:s.index]
}
