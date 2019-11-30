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

func (p PiecewiseFloats) String() string {
	result := ""
	for i, a := range p {
		result += fmt.Sprintf("%v: %v\n", i, a)
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
			return result, fmt.Errorf("internal parse error can not happen err=%v", errParse)
		}

		v := a.ScaleToFloat(uint64(pieceval))
		if !math.IsNaN(v) || (allowNaN && math.IsNaN(v)) {
			result[a.Name] = v
		}
	}
	return result, nil
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

//Encode map of float values to byte struct. Low level function. Call Splurts
func (p *PiecewiseFloats) Encode(values map[string]float64) ([]byte, error) {
	bitString := ""
	for _, a := range *p {
		f, haz := values[a.Name]
		if haz {
			code := a.BitCode(f)
			bitString += code
		} else {
			maxCodeBin := fmt.Sprintf("%b", a.MaxCode())
			bitString += maxCodeBin //All bits up
		}
	}
	neededPad := 8 - (len(bitString) % 8)
	if 0 < neededPad {
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
