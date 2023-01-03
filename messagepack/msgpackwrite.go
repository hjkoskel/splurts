/*
Writing functions for messagepack.
Generic and ad hoc solutions

Trying to avoid dependencies if only needed thing is to write messagepack
*/
package messagepack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func WriteBool(buf *bytes.Buffer, value bool) error {
	if value {
		return buf.WriteByte(0xc3)
	}
	return buf.WriteByte(0xc2)
}

func WriteNumber(buf *bytes.Buffer, f float64, maxErr float64) error {
	i := int64(f)
	if float64(i) == f {
		return WriteInt(buf, i)
	}

	f32 := float32(f)
	if math.Abs(float64(f32)-f) < maxErr {
		return writeFloat32(buf, f32)
	}
	return writeFloat64(buf, f)
}

func WriteInt(buf *bytes.Buffer, i int64) error {

	//Use unsigned if positive? better compression?  Messagepack is a little stupid?
	if 0 <= i {
		return writeUInt(buf, uint64(i))
	}

	if -32 <= i {
		return binary.Write(buf, binary.BigEndian, int8(i))
	}
	if -127 <= i {
		e := buf.WriteByte(0xd0)
		if e != nil {
			return e
		}
		return binary.Write(buf, binary.BigEndian, int8(i))
	}
	if -32767 <= i {
		e := buf.WriteByte(0xd1)
		if e != nil {
			return e
		}
		return binary.Write(buf, binary.BigEndian, int16(i))
	}

	if -2147483647 <= i {
		e := buf.WriteByte(0xd2)
		if e != nil {
			return e
		}
		return binary.Write(buf, binary.BigEndian, int32(i))
	}
	e := buf.WriteByte(0xd3)
	if e != nil {
		return e
	}
	return binary.Write(buf, binary.BigEndian, i)
}

func WriteArray(buf *bytes.Buffer, n uint32) error {
	if n < 0 {
		return fmt.Errorf("Invalid number")
	}
	if 0xFFFFFFF < n {
		return fmt.Errorf("too many")
	}
	if n <= 0x0F {
		return buf.WriteByte(0x90 | byte(n))

	}
	if n < 0xFFFF {
		e := buf.WriteByte(0xdc)
		if e != nil {
			return e
		}
		return binary.Write(buf, binary.BigEndian, uint16(n))
	}

	e := buf.WriteByte(0xdd)
	if e != nil {
		return e
	}
	return binary.Write(buf, binary.BigEndian, n)
}

func WriteFixmap(buf *bytes.Buffer, n uint32) error {
	if n < 0 {
		return fmt.Errorf("Invalid number")
	}
	if 0xFFFFFFF < n {
		return fmt.Errorf("too many")
	}
	if n <= 0x0F {
		return buf.WriteByte(0x80 | byte(n))

	}
	if n < 0xFFFF {
		e := buf.WriteByte(0xde)
		if e != nil {
			return e
		}
		return binary.Write(buf, binary.BigEndian, uint16(n))
	}

	e := buf.WriteByte(0xdf)
	if e != nil {
		return e
	}
	return binary.Write(buf, binary.BigEndian, uint32(n))
}

func WriteString(buf *bytes.Buffer, s string) error { //TODO only 8 bit chars... or?
	data := []byte(s) //handle UTF-8 names?
	n := len(data)
	if n <= 0x1F {
		e := buf.WriteByte(0xA0 | byte(n))
		if e != nil {
			return e
		}
		_, e = buf.WriteString(s)
		return e
	}
	if n <= 0xFF {
		_, e := buf.Write([]byte{0xd9, byte(n)})
		if e != nil {
			return e
		}
		_, e = buf.WriteString(s)
		return e
	}

	if n <= 0xFFFF {
		e := buf.WriteByte(0xda)
		if e != nil {
			return e
		}
		e = binary.Write(buf, binary.BigEndian, uint16(n))
		if e != nil {
			return e
		}
		_, e = buf.WriteString(s)
		return e
	}
	e := buf.WriteByte(0xdb)
	if e != nil {
		return e
	}
	e = binary.Write(buf, binary.BigEndian, uint32(n))
	if e != nil {
		return e
	}
	_, e = buf.WriteString(s)
	return e
}

//Low level
func writeFloat64(buf *bytes.Buffer, f float64) error {
	e := buf.WriteByte(0xcb)
	if e != nil {
		return e
	}
	return binary.Write(buf, binary.BigEndian, f)
}

func writeUInt(buf *bytes.Buffer, u uint64) error {
	if u <= 0x7F {
		return buf.WriteByte(byte(u))
	}
	if u <= 0xFF {
		e := buf.WriteByte(0xcc)
		if e != nil {
			return e
		}
		return buf.WriteByte(byte(u))
	}

	if u <= 0xFFFF {
		e := buf.WriteByte(0xcd)
		if e != nil {
			return e
		}
		return binary.Write(buf, binary.BigEndian, uint16(u))
	}

	if u <= 0xFFFFFFFF {
		e := buf.WriteByte(0xce)
		if e != nil {
			return e
		}
		return binary.Write(buf, binary.BigEndian, uint32(u))
	}
	e := buf.WriteByte(0xcf)
	if e != nil {
		return e
	}
	return binary.Write(buf, binary.BigEndian, u)
}

func writeFloat32(buf *bytes.Buffer, f float32) error {
	e := buf.WriteByte(0xca)
	if e != nil {
		return e
	}
	return binary.Write(buf, binary.BigEndian, f)
}
