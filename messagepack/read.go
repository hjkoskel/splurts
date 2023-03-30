/*
Raw messagepack functions
*/
package messagepack

import (
	"encoding/binary"
	"fmt"
	"io"
)

/*
func IsString(first byte) bool {
	return first&0xF0 == 0xA0 || first == 0x9d || first == 0xda
}
*/

func ReadString(buf io.Reader) (string, error) {
	//Assuming it is string...fail if it is not

	var first byte

	firstErr := binary.Read(buf, binary.BigEndian, &first)
	if firstErr != nil {
		return "", firstErr
	}

	if first == 0xc0 { //NIL
		return "", nil
	}

	/*if (first < 0xa0 || 0xbf < first) && first != 0x9d {
		return "", fmt.Errorf("is not string 0x%X", first)
	}*/
	var n int64
	n = -1
	if first&0xE0 == 0xA0 {
		n = int64(uint32(first & 0xF))
	}
	if first == 0xd9 {
		var u8 uint8
		readErr := binary.Read(buf, binary.BigEndian, &u8)
		if readErr != nil {
			return "", readErr
		}
		n = int64(u8)
	}
	if first == 0xda {
		var u16 uint16
		readErr := binary.Read(buf, binary.BigEndian, &u16)
		if readErr != nil {
			return "", readErr
		}
		n = int64(u16)
	}
	if first == 0xdb {
		var u32 uint32
		readErr := binary.Read(buf, binary.BigEndian, &u32)
		if readErr != nil {
			return "", readErr
		}
		n = int64(u32)
	}

	if n < 0 {
		return "", fmt.Errorf("is not string 0x%X", first)
	}

	s := make([]byte, n)
	_, readErr := io.ReadFull(buf, s)
	return string(s), readErr
}

func IsFixmap(first byte) bool {
	return first&0xF0 == 0x80 || first == 0xde || first == 0xdf
}

func ReadFixmap(buf io.Reader) (uint32, error) {
	var first byte
	errFirst := binary.Read(buf, binary.BigEndian, &first)
	if errFirst != nil {
		return 0, errFirst
	}

	if first == 0xc0 { //NIL
		return 0, nil
	}

	if !IsFixmap(first) {
		return 0, fmt.Errorf("is not fixmap")
	}

	if first&0xF0 == 0x80 {
		return uint32(first & 0x0F), nil
	}
	if first == 0xde {
		var u16 uint16
		readErr := binary.Read(buf, binary.BigEndian, &u16)
		return uint32(u16), readErr
	}
	var u32 uint32
	readErr := binary.Read(buf, binary.BigEndian, &u32)
	return u32, readErr
}

func IsArr(first byte) bool {
	return first&0xF0 == 0x90 || first == 0xdc || first == 0xdd
}

func ReadArrWithFirst(buf io.Reader, first byte) (uint32, error) { //Just read how many and how long after
	if first == 0xc0 { //NIL
		return 0, nil
	}
	if !IsArr(first) {
		return 0, fmt.Errorf("invalid arr start %X", first)
	}
	if first&0xF0 == 0x90 {
		return uint32(first & 0x0F), nil
	}
	if first == 0xdc {
		var u16 uint16
		readErr := binary.Read(buf, binary.BigEndian, &u16)
		return uint32(u16), readErr
	}
	var u32 uint32
	readErr := binary.Read(buf, binary.BigEndian, &u32)
	return uint32(u32), readErr
}

func ReadArr(buf io.Reader) (uint32, error) { //Just read how many and how long after
	var first byte
	errFirst := binary.Read(buf, binary.BigEndian, &first)
	if errFirst != nil {
		return 0, errFirst
	}
	return ReadArrWithFirst(buf, first)
}

func ReadBool(buf io.Reader) (bool, error) {
	var first byte
	errFirst := binary.Read(buf, binary.BigEndian, &first)
	if errFirst != nil {
		return false, errFirst
	}
	if first == 0xc3 {
		return true, nil
	}
	if first == 0xc2 {
		return false, nil
	}
	return false, fmt.Errorf("value 0x%X is not boolean", first)
}

func ReadNumber(buf io.Reader) (float64, error) {
	var first byte
	errFirst := binary.Read(buf, binary.BigEndian, &first)
	if errFirst != nil {
		return 0, errFirst
	}
	return ReadNumberWithFirst(buf, first)
}

func ReadNumberWithFirst(buf io.Reader, first byte) (float64, error) {
	if first == 0xcb { //float64
		var f64 float64
		e := binary.Read(buf, binary.BigEndian, &f64)
		if e != nil {
			return 0, e
		}
		return f64, nil
	}
	if first == 0xca {
		var f32 float32
		e := binary.Read(buf, binary.BigEndian, &f32)
		if e != nil {
			return 0, e
		}
		return float64(f32), nil
	}
	i, iErr := ReadIntWithFirst(buf, first)
	if iErr != nil {
		return 0, iErr
	}
	return float64(i), nil
}

func ReadBinArray(buf io.Reader) ([]byte, error) {
	var first byte
	errFirst := binary.Read(buf, binary.BigEndian, &first)
	if errFirst != nil {
		return nil, errFirst
	}
	return ReadBinArrayWithFirst(buf, first)
}

func ReadBinArrayWithFirst(buf io.Reader, first byte) ([]byte, error) {
	var result []byte
	switch first {
	case 0xc4:
		var u8 uint8
		if err := binary.Read(buf, binary.BigEndian, &u8); err != nil {
			return nil, err
		}
		result = make([]byte, u8)
	case 0xc5:
		var u16 uint16
		if err := binary.Read(buf, binary.BigEndian, &u16); err != nil {
			return nil, err
		}
		result = make([]byte, u16)
	case 0xc6:
		var u32 uint32
		if err := binary.Read(buf, binary.BigEndian, &u32); err != nil {
			return nil, err
		}
		result = make([]byte, u32)
	default:
		return nil, fmt.Errorf("invalid first byte 0x%X for binarr", first)
	}
	_, err := buf.Read(result)
	return result, err
}

func ReadInt(buf io.Reader) (int64, error) {
	var first byte
	errFirst := binary.Read(buf, binary.BigEndian, &first)
	if errFirst != nil {
		return 0, errFirst
	}
	return ReadIntWithFirst(buf, first)
}

// Read signed int... gives byte that was already read out
func ReadIntWithFirst(buf io.Reader, first byte) (int64, error) {
	if first&0x80 == 0 { //pos integer 8bit
		return int64(first), nil
	}
	if first&0xE0 == 0xE0 {
		return int64(int8(first)), nil
	}
	switch first {
	case 0xcc:
		var u8 uint8
		readErr := binary.Read(buf, binary.BigEndian, &u8)
		return int64(u8), readErr
	case 0xcd:
		var u16 uint16
		readErr := binary.Read(buf, binary.BigEndian, &u16)
		return int64(u16), readErr
	case 0xce:
		var u32 uint32
		readErr := binary.Read(buf, binary.BigEndian, &u32)
		return int64(u32), readErr
	case 0xcf:
		var u64 uint64
		readErr := binary.Read(buf, binary.BigEndian, &u64)
		return int64(u64), readErr //TODO error if goes over signed 64. TODO make custom function
		//return 0, 0, fmt.Errorf("use unsigned read for uint64")
	case 0xd0:
		var i8 int8
		readErr := binary.Read(buf, binary.BigEndian, &i8)
		return int64(i8), readErr
	case 0xd1:
		var i16 int16
		readErr := binary.Read(buf, binary.BigEndian, &i16)
		return int64(i16), readErr
	case 0xd2:
		var i32 int32
		readErr := binary.Read(buf, binary.BigEndian, &i32)
		return int64(i32), readErr
	case 0xd3:
		var i64 int64
		readErr := binary.Read(buf, binary.BigEndian, &i64)
		return i64, readErr
	default:
		return 0, fmt.Errorf("unknown first byte on ReadInt 0x%X", first)
	}

}
