/*
Functions for creating delta coding in messagepack integer vector.
Int does not introduce rounding error

Supports two levels

-Start and then deltas
-Startvalue, initial delta and then deltas of delta

Also RLE packing after repeating values

*/
package messagepack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

//Keep messagepack format in memory... search etc... then write just append that to messagepack
type DeltaRLEVec []byte

func ReadDeltaRLEVec(buf io.Reader) (DeltaRLEVec, error) {
	n, errN := ReadArr(buf) //Number of elements. can be numbers or actual
	if errN != nil {
		return nil, errN
	}

	result := new(bytes.Buffer)

	wErr := WriteArray(result, n)
	if wErr != nil {
		return nil, wErr
	}

	for element := 0; element < int(n); element++ {
		var first byte
		errFirst := binary.Read(buf, binary.BigEndian, &first)

		if errFirst != nil {
			return nil, errFirst
		}

		if IsArr(first) {
			//fmt.Printf("first is arr\n")
			is2, is2err := ReadArrWithFirst(buf, first)
			if is2err != nil {
				return nil, is2err
			}
			if is2 != 2 {
				return nil, fmt.Errorf("rle data have array length of %v, only len 2 is allowed", is2)
			}

			wErr := WriteArray(result, 2)
			if wErr != nil {
				return nil, wErr
			}

			intval, errintval := ReadInt(buf)
			if errintval != nil {
				return nil, errintval
			}

			wErr = WriteInt(result, intval)
			if wErr != nil {
				return nil, wErr
			}
			countval, errcountval := ReadInt(buf)
			if errcountval != nil {
				return nil, errcountval
			}

			wErr = WriteInt(result, countval)
			if wErr != nil {
				return nil, wErr
			}

		} else {
			intval, errintval := ReadIntWithFirst(buf, first)
			if errintval != nil {
				return nil, errintval
			}
			wErr := WriteInt(result, intval)
			if wErr != nil {
				return nil, wErr
			}
		}
	}
	return result.Bytes(), nil
}

func (p *DeltaRLEVec) ToArr(deltas int) ([]int64, error) {
	if 2 < deltas {
		return nil, fmt.Errorf("number of deltas %v not supported", deltas)
	}

	codebuf := bytes.NewBuffer(*p)
	unpacked, unpackerr := RLEMessagepackToArr(codebuf)
	if unpackerr != nil {
		return nil, unpackerr
	}

	for d := 0; d <= deltas; d++ {
		unpacked = UnDeltaVec(unpacked)
	}

	return unpacked, nil
}

func (p *DeltaRLEVec) WriteToBuf(buf *bytes.Buffer) error {
	_, e := buf.Write(*p)
	return e
}

func DeltaVec(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}
	result := make([]int64, len(values))
	previous := values[0]
	for i, v := range values {
		result[i] = v - previous
		previous = v
	}
	result[0] = values[0] //First is 0 anyways, use that for setting that as start value
	return result
}

func UnDeltaVec(v []int64) []int64 {
	if len(v) < 2 {
		return nil
	}
	n := len(v)
	result := make([]int64, n)
	result[0] = v[0]
	for i := 1; i < n; i++ {
		result[i] = result[i-1] + v[i]
	}
	return result
}

func writeRLE(buf *bytes.Buffer, value int64, count int64, rleLimit int64) (int64, error) {
	if rleLimit <= count && 0 < rleLimit {
		e := WriteArray(buf, 2)
		if e != nil {
			return 0, e
		}
		e = WriteInt(buf, value)
		if e != nil {
			return 0, e
		}
		return 1, WriteInt(buf, count)
	}
	for i := 0; i < int(count); i++ { //TODO uint32 int64 at writeInt
		e := WriteInt(buf, value)
		if e != nil {
			return 0, e
		}
	}

	return count, nil
}

func ArrToMessagepack(buf *bytes.Buffer, arr []int64) error {
	e := WriteArray(buf, uint32(len(arr)))
	if e != nil {
		return e
	}

	for _, value := range arr {
		e := WriteInt(buf, value)
		if e != nil {
			return e
		}
	}
	return nil
}

func ArrToRLEMessagepack(buf *bytes.Buffer, arr []int64, rleLimit int64) error {
	if len(arr) == 0 {
		return nil
	}
	workbuf := new(bytes.Buffer)

	itemcount := int64(0)

	repeatedCount := int64(0)
	previous := arr[0]
	for _, v := range arr {
		if previous == v {
			repeatedCount++
		} else {
			itemsWritten, e := writeRLE(workbuf, previous, repeatedCount, rleLimit)
			if e != nil {
				return e
			}
			repeatedCount = 1
			itemcount += itemsWritten
		}
		previous = v
	}

	itemsWritten, e := writeRLE(workbuf, previous, repeatedCount, rleLimit)
	if e != nil {
		return e
	}
	itemcount += itemsWritten

	e = WriteArray(buf, uint32(itemcount))
	if e != nil {
		return e
	}
	_, e = buf.Write(workbuf.Bytes())
	return e
}

func RLEMessagepackToArr(buf *bytes.Buffer) ([]int64, error) {
	first, firstErr := buf.ReadByte()
	if firstErr != nil {
		return nil, firstErr
	}
	itemcount, errItemCount := ReadArrWithFirst(buf, first)
	if errItemCount != nil {
		return nil, errItemCount
	}

	result := []int64{}

	//Only numbers and two item arrays
	for itemcounter := uint32(0); itemcounter < itemcount; itemcounter++ {
		first, firstErr := buf.ReadByte()
		if firstErr != nil {
			return nil, firstErr
		}
		if IsArr(first) {
			two, errTwo := ReadArrWithFirst(buf, first)
			if errTwo != nil {
				return nil, errTwo
			}
			if two != 2 {
				return nil, fmt.Errorf("item %v array length is %v, not 2", itemcounter, two)
			}
			value, errValue := ReadInt(buf)
			if errValue != nil {
				return nil, errValue
			}
			repeats, errRepeats := ReadInt(buf)
			if errRepeats != nil {
				return nil, errRepeats
			}
			for i := 0; i < int(repeats); i++ {
				result = append(result, value)
			}
		} else {

			value, valueErr := ReadIntWithFirst(buf, first)
			if valueErr != nil {
				return nil, valueErr
			}
			result = append(result, value)
		}
	}
	return result, nil
}
