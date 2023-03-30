/*
Writing functions for messagepack.
Generic and ad hoc solutions

Trying to avoid dependencies if only needed thing is to write messagepack
*/
package messagepack

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// WriteBool writes bool value to buffer
func WriteBool(w io.Writer, value bool) error {
	if value {
		_, e := w.Write([]byte{0xc3})
		return e
	}
	_, e := w.Write([]byte{0xc2})
	return e
}

func WriteNumber(w io.Writer, f float64, maxErr float64) error {
	i := int64(f)
	if float64(i) == f {
		return WriteInt(w, i)
	}
	f32 := float32(f)
	if math.Abs(float64(f32)-f) < maxErr {
		return writeFloat32(w, f32)
	}
	return writeFloat64(w, f)
}

func WriteInt(w io.Writer, i int64) error {
	//Use unsigned if positive? better compression?  Messagepack is a little stupid?
	if 0 <= i {
		return writeUInt(w, uint64(i))
	}

	if -32 <= i {
		return binary.Write(w, binary.BigEndian, int8(i))
	}
	if -127 <= i {
		_, e := w.Write([]byte{0xd0})
		if e != nil {
			return e
		}
		return binary.Write(w, binary.BigEndian, int8(i))
	}
	if -32767 <= i {
		_, e := w.Write([]byte{0xd1})
		if e != nil {
			return e
		}
		return binary.Write(w, binary.BigEndian, int16(i))
	}

	if -2147483647 <= i {
		_, e := w.Write([]byte{0xd2})
		if e != nil {
			return e
		}
		return binary.Write(w, binary.BigEndian, int32(i))
	}
	_, e := w.Write([]byte{0xd3})
	if e != nil {
		return e
	}
	return binary.Write(w, binary.BigEndian, i)
}

func WriteArray(w io.Writer, n uint32) error {
	if 0xFFFFFFF < n {
		return fmt.Errorf("too many")
	}
	if n <= 0x0F {
		_, e := w.Write([]byte{0x90 | byte(n)})
		return e
	}
	if n < 0xFFFF {
		_, e := w.Write([]byte{0xdc})
		if e != nil {
			return e
		}
		return binary.Write(w, binary.BigEndian, uint16(n))
	}

	_, e := w.Write([]byte{0xdd})
	if e != nil {
		return e
	}
	return binary.Write(w, binary.BigEndian, n)
}

func WriteFixmap(w io.Writer, n uint32) error {
	if 0xFFFFFFF < n {
		return fmt.Errorf("too many")
	}
	if n <= 0x0F {
		_, e := w.Write([]byte{0x80 | byte(n)})
		return e
	}
	if n < 0xFFFF {
		_, e := w.Write([]byte{0xde})
		if e != nil {
			return e
		}
		return binary.Write(w, binary.BigEndian, uint16(n))
	}

	_, e := w.Write([]byte{0xdf})
	if e != nil {
		return e
	}
	return binary.Write(w, binary.BigEndian, uint32(n))
}

func WriteStringMapString(w io.Writer, m map[string]string) error {
	for name, value := range m {
		err := WriteString(w, name)
		if err != nil {
			return err
		}
		err = WriteString(w, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteNil(w io.Writer) error {
	_, e := w.Write([]byte{0xc0})
	return e
}

func WriteString(w io.Writer, s string) error { //TODO only 8 bit chars... or?
	data := []byte(s) //handle UTF-8 names?
	n := len(data)
	if n <= 0x1F {
		_, e := w.Write([]byte{0xA0 | byte(n)})
		if e != nil {
			return e
		}
		_, e = w.Write([]byte(s))
		return e
	}
	if n <= 0xFF {
		_, e := w.Write([]byte{0xd9, byte(n)})
		if e != nil {
			return e
		}
		_, e = w.Write([]byte(s))
		return e
	}

	if n <= 0xFFFF {
		_, e := w.Write([]byte{0xda})
		if e != nil {
			return e
		}
		e = binary.Write(w, binary.BigEndian, uint16(n))
		if e != nil {
			return e
		}
		_, e = w.Write([]byte(s))
		return e
	}
	_, e := w.Write([]byte{0xdb})
	if e != nil {
		return e
	}
	e = binary.Write(w, binary.BigEndian, uint32(n))
	if e != nil {
		return e
	}
	_, e = w.Write([]byte(s))
	return e
}

// Low level
func writeFloat64(w io.Writer, f float64) error {
	_, e := w.Write([]byte{0xcb})
	if e != nil {
		return e
	}
	return binary.Write(w, binary.BigEndian, f)
}

func writeUInt(w io.Writer, u uint64) error {
	if u <= 0x7F {
		_, e := w.Write([]byte{byte(u)})
		return e
	}
	if u <= 0xFF {
		_, e := w.Write([]byte{0xcc})
		if e != nil {
			return e
		}
		_, e = w.Write([]byte{byte(u)})
		return e
	}

	if u <= 0xFFFF {
		_, e := w.Write([]byte{0xcd})
		if e != nil {
			return e
		}
		return binary.Write(w, binary.BigEndian, uint16(u))
	}

	if u <= 0xFFFFFFFF {
		_, e := w.Write([]byte{0xce})
		if e != nil {
			return e
		}
		return binary.Write(w, binary.BigEndian, uint32(u))
	}
	_, e := w.Write([]byte{0xcf})
	if e != nil {
		return e
	}
	return binary.Write(w, binary.BigEndian, u)
}

func writeFloat32(w io.Writer, f float32) error {
	_, e := w.Write([]byte{0xca})
	if e != nil {
		return e
	}
	return binary.Write(w, binary.BigEndian, f)
}

func WriteBinArray(w io.Writer, data []byte) error {
	n := len(data)
	if n <= 0xFF {
		_, e := w.Write([]byte{0xc4, byte(n)})
		if e != nil {
			return e
		}
	} else {
		if n <= 0xFFFF {
			_, e := w.Write([]byte{0xc5})
			if e != nil {
				return e
			}
			e = binary.Write(w, binary.BigEndian, uint16(n))
			if e != nil {
				return e
			}
		} else {
			_, e := w.Write([]byte{0xc6})
			if e != nil {
				return e
			}
			e = binary.Write(w, binary.BigEndian, uint32(n))
			if e != nil {
				return e
			}
		}
	}
	_, e := w.Write(data)
	return e
}
