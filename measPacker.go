/*
Packs up float64 struct into binary blob. Does not obey byte boundaries but pads results to full bytes
*/

package splurts

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

//PiecewiseFloats describe order and how values on struct are scaled
type PiecewiseFloats []PiecewiseCoding //Map does not provide fixed order must be array

func (p *PiecewiseFloats) getCoding(name string) (PiecewiseCoding, error) {
	for _, result := range *p {
		if result.Name == name {
			return result, nil
		}
	}
	return PiecewiseCoding{}, fmt.Errorf("name %v not found", name)
}

func (p PiecewiseFloats) String() string {
	result := ""
	for i, a := range p {
		if 0 < len(a.Enums) {
			result += fmt.Sprintf("%v: %v (enums:%#v)\n", i, a, a.Enums)
		} else {
			result += fmt.Sprintf("%v: %v\n", i, a)
		}
	}
	return strings.TrimSpace(result)
}

//NumberOfBits actual bits. Result byte array is padded with 0 bits at end
func (p *PiecewiseFloats) NumberOfBits() int {
	result := 0
	for _, piece := range *p {
		result += piece.NumberOfBits()
	}
	return result
}

//Decodes hex string 4bit or 8bit
func (p *PiecewiseFloats) DecodeHex(hexString string, allowNaN bool) (map[string]float64, error) {
	s := hexString
	if len(s)%2 != 0 {
		s = s + "0"
	}

	binarr := []byte{}
	for n := 0; n*2 < len(s); n++ {
		piece := s[n*2 : n*2+2]
		v, parseErr := strconv.ParseInt(piece, 16, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("Invalid hex string")
		}
		binarr = append(binarr, byte(v))
	}
	return p.Decode(binarr, allowNaN)
}

//Decode codes byte array to float value map.   If values are missing and allowNaN is true. Skip or replace with NaN on map
func (p *PiecewiseFloats) Decode(binarr []byte, allowNaN bool) (map[string]float64, error) {
	if int(math.Ceil(float64(p.NumberOfBits())/8.0)) != len(binarr) {
		return nil, fmt.Errorf("Struct have %v bits means %v bytes. BUT binary array have %v bytes",
			p.NumberOfBits(),
			int(math.Ceil(float64(p.NumberOfBits())/8.0)),
			len(binarr))
	}

	result := make(map[string]float64)
	binstr := ""
	for _, b := range binarr {
		binstr += fmt.Sprintf("%08b", b)
	}

	for _, a := range *p {
		//Pick bits for variable
		bits := a.NumberOfBits()
		piece := binstr[0:bits]
		binstr = binstr[bits:]

		pieceval, errParse := strconv.ParseInt(piece, 2, 64)
		if errParse != nil { //Non-unit testable
			return result, fmt.Errorf("internal parse error can not happen err=%v  (piece=%v, bits=%v)", errParse, piece, a.NumberOfBits())
		}

		if 0 < len(a.Enums) {
			if len(a.Enums) < int(pieceval) {
				return result, fmt.Errorf("variable %s have value %v, but it have %v enums + empty", a.Name, pieceval, len(a.Enums))
			}
			result[a.Name] = float64(pieceval)
		}

		v := a.ScaleToFloat(uint64(pieceval))
		if !math.IsNaN(v) || (allowNaN && math.IsNaN(v)) {
			result[a.Name] = v
		}
	}
	return result, nil
}

//Formatted to 7bit
type SevenBitArr []byte

//Decode7bitBytes decode, skipping MSB from
func (p *PiecewiseFloats) Decode7bitBytes(binarr SevenBitArr, allowNaN bool) (map[string]float64, error) {
	if len(binarr) == 0 {
		return nil, fmt.Errorf("No data")
	}
	//Naive solution, optimize later
	var sb strings.Builder
	//Remove zero bits from array
	for i, b := range binarr {
		if 127 < b {
			return nil, fmt.Errorf("Non 7-bit byte vector at index %v on %#X", i, binarr)
		}
		sb.WriteString(fmt.Sprintf("%.7b", b))
	}

	//Trim to expected length and split
	s := sb.String()
	bits := p.NumberOfBits()
	if len(s) < bits {
		return nil, fmt.Errorf("Not enough bits")
	}
	arr := splitFixedSizePieces(s[:bits], 8)
	//Pad if needed.. add AFTER zeros
	last := len(arr) - 1
	for len(arr[last]) < 8 {
		arr[last] += "0"
	}

	result := make([]byte, len(arr))
	for i, v := range arr {
		intv, _ := strconv.ParseInt(v, 2, 16)
		result[i] = byte(intv)
	}
	return p.Decode(result, allowNaN)
}

//IsInvalid check with this before further proceccing
func (p *PiecewiseFloats) IsInvalid() error {
	for i, a := range *p {
		if a.Name == "" {
			return fmt.Errorf("Name on index %v is not defined", i)
		}
		errValid := a.IsInvalid()
		if errValid != nil {
			return fmt.Errorf("Name is %v not valid: %v (%#v)", a.Name, errValid.Error(), a)
		}
	}
	return nil
}

func (p *PiecewiseFloats) EncodeToBitString(values map[string]float64) string {
	bitString := ""
	for _, a := range *p {
		f, haz := values[a.Name]
		if haz {
			s, _ := a.BitCode(f)
			bitString += s
		} else {
			bitString += fmt.Sprintf("%b", a.MaxCode())
		}
	}
	return bitString
}

//pad to 4bit
func (p *PiecewiseFloats) EncodeToHexNybble(values map[string]float64) (string, error) {
	bitString := p.EncodeToBitString(values)
	neededPad := 4 - (len(bitString) % 4)
	if 0 < neededPad {
		padformat := "%0" + fmt.Sprintf("%v", neededPad) + "b"
		bitString = bitString + fmt.Sprintf(padformat, 0)
	}

	length := len(bitString)
	result := ""
	for n := 0; n*4 < length; n++ {
		piece := bitString[n*4 : n*4+4]
		v, parseErr := strconv.ParseInt(piece, 2, 64)
		if parseErr != nil {
			return result, fmt.Errorf("Code internal failure %v", parseErr) //can not happen
		}
		result += fmt.Sprintf("%X", v)
	}
	return result, nil
}

//pad so it will fit to 8bit
func (p *PiecewiseFloats) EncodeToHex(values map[string]float64) (string, error) {
	result, err := p.EncodeToHexNybble(values)
	if len(result)%2 != 0 {
		result += "0"
	}
	return result, err
}

func bitStringToByteArr(bitString string) ([]byte, error) {
	neededPad := 8 - (len(bitString) % 8)
	if (0 != neededPad) && (8 != neededPad) {
		padformat := "%0" + fmt.Sprintf("%v", neededPad) + "b"
		bitString = bitString + fmt.Sprintf(padformat, 0)
	}
	length := len(bitString)
	result := []byte{}

	for n := 0; n*8 < length; n++ {
		piece := bitString[n*8 : n*8+8]
		v, parseErr := strconv.ParseInt(piece, 2, 64)
		if parseErr != nil {
			return result, fmt.Errorf("Code internal failure %v", parseErr)
		}
		result = append(result, byte(v))
	}
	//Split array in 8 long pieces
	return result, nil
}

//Encode map of float values to byte struct. Low level function. Call Splurts
func (p *PiecewiseFloats) Encode(values map[string]float64) ([]byte, error) {
	return bitStringToByteArr(p.EncodeToBitString(values))
}

func splitFixedSizePieces(s string, step int) []string {
	arr := make([]string, int(math.Ceil(float64(len(s))/float64(step))))
	for i := range arr {
		arr[i] = s[i*step : int(math.Min(float64(len(s)), float64((i+1)*step)))]
	}

	return arr
}

//Encode7bitBytes  used in FPGA projects when MSB bit reserved for data/command flag
func (p *PiecewiseFloats) Encode7bitBytes(values map[string]float64) (SevenBitArr, error) {
	s := p.EncodeToBitString(values)
	pieces := splitFixedSizePieces(s, 7)
	if len(pieces) == 0 {
		return nil, nil
	}
	neededPad := 7 - (len(pieces[len(pieces)-1]) % 7)

	if (neededPad != 0) && (neededPad != 7) {
		padformat := "%0" + fmt.Sprintf("%v", neededPad) + "b"
		pieces[len(pieces)-1] = pieces[len(pieces)-1] + fmt.Sprintf(padformat, 0)
	}
	result := "0" + strings.Join(pieces, "0")
	return bitStringToByteArr(result) //add one front zero per 7bit byte -> 8bit byte
}
