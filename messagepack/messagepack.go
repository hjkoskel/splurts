package messagepack

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hjkoskel/splurts"
)

/*
Convert series of values to messagepack.

Contains basic information for conversion
And then array values coded in integer format?
Float is such waste

Code piecewise by piecewise. Avoid memory peak consumption

*/

const (
	MPNAME_CODING         = "coding"
	MPNAME_CODING_MIN     = "min"
	MPNAME_CODING_MAX     = "max"
	MPNAME_CODING_CLAMPED = "cla"

	MPNAME_ENUMS = "enums"
	MPNAME_STEPS = "steps"

	MPNAME_POINTCOUNT = "n"
	MPNAME_STEPSIZE   = "s"

	MPCODING_SIMPLEARR = "arr"
	MPCODING_DELTA1ARR = "delta1"
	MPCODING_DELTA2ARR = "delta2"

	MPDIRECTIVE       = "messagepack"
	MPDIRECTIVE_NAME  = "name"
	MPDIRECTIVE_DELTA = "delta" //0=no delta coding, 1 once, 2 twice
	MPDIRECTIVE_RLE   = "rle"   //0=not in use 1=more than 1 then use array, at least 3 prefered
)

//Requires just string and bytes
func toMessagepackHead(b *bytes.Buffer, p splurts.PiecewiseCoding, name string) error {

	minstep := p.MinStep()
	e := WriteString(b, name) //Tell it name for map
	if e != nil {
		return e
	}

	if 0 < len(p.Enums) {
		e = WriteFixmap(b, 2) //TODO HOW MANY
		if e != nil {
			return e
		}
		e = WriteString(b, MPNAME_ENUMS)
		if e != nil {
			return e
		}
		e = WriteArray(b, uint32(len(p.Enums)))
		if e != nil {
			return e
		}
		for _, s := range p.Enums {
			e = WriteString(b, s)
			if e != nil {
				return e
			}
		}
		return nil

	}
	e = WriteFixmap(b, 3) //TODO HOW MANY
	if e != nil {
		return e
	}

	//give range
	e = WriteString(b, MPNAME_CODING) //1
	if e != nil {
		return e
	}
	e = WriteFixmap(b, 3)
	if e != nil {
		return e
	}
	e = WriteString(b, MPNAME_CODING_MIN) //	1
	if e != nil {
		return e
	}
	e = WriteNumber(b, p.Min, minstep)
	if e != nil {
		return e
	}
	e = WriteString(b, MPNAME_CODING_MAX) //	2
	if e != nil {
		return e
	}
	e = WriteNumber(b, p.Max(), minstep)
	if e != nil {
		return e
	}
	e = WriteString(b, MPNAME_CODING_CLAMPED) //	3
	if e != nil {
		return e
	}
	e = WriteBool(b, p.Clamped)
	if e != nil {
		return e
	}

	e = WriteString(b, MPNAME_STEPS) //5
	if e != nil {
		return e
	}
	e = WriteArray(b, uint32(len(p.Steps)))
	if e != nil {
		return e
	}
	for _, step := range p.Steps {
		e = WriteFixmap(b, 2)
		if e != nil {
			return e
		}
		e = WriteString(b, MPNAME_POINTCOUNT)
		if e != nil {
			return e
		}
		e = writeUInt(b, step.Count)
		if e != nil {
			return e
		}
		e = WriteString(b, MPNAME_STEPSIZE)
		if e != nil {
			return e
		}
		e = WriteNumber(b, step.Size, minstep/10) //requires extra decimal?  32 vs 64
		if e != nil {
			return e
		}

	}
	return nil
}

func WriteSplurtsVariableArr(b *bytes.Buffer, p splurts.PiecewiseCoding, rleLimit int, deltas int, overrideName string, arr []float64) error {
	if len(arr) == 0 {
		return fmt.Errorf("no data")
	}

	if len(overrideName) == 0 {
		overrideName = p.Name
	}

	e := toMessagepackHead(b, p, overrideName)
	if e != nil {
		return e
	}

	codingName, hazCoding := map[int]string{
		0: MPCODING_SIMPLEARR,
		1: MPCODING_DELTA1ARR,
		2: MPCODING_DELTA2ARR,
	}[deltas]

	if !hazCoding {
		return fmt.Errorf("deltaCoding %v not supported", deltas)
	}
	e = WriteString(b, codingName) //TODO better codings?  delta? dualdelta?
	if e != nil {
		return e
	}

	arrInt64 := make([]int64, len(arr)) //TODO issue.... not possible to use all 64bit.. with delta coding :(
	//TODO previousVal := p.ScaleToUint(arr[0]) //Initialization
	for i, f := range arr {
		//TODO add later previousVal = p.ScaleToUintByPreviousUint(f, previousVal)
		//TODO ADD later arrInt64[i] = int64(previousVal)
		arrInt64[i] = int64(p.ScaleToUint(f))
	}

	if 2 < deltas {
		return fmt.Errorf("number of deltas %v not supported", deltas)
	}

	for d := 0; d < deltas; d++ {
		arrInt64 = DeltaVec(arrInt64)
	}

	return ArrToRLEMessagepack(b, arrInt64, int64(rleLimit))
}

func getMessagepackTagString(input interface{}) (map[string]string, error) {
	result := make(map[string]string)

	rt := reflect.TypeOf(input)

	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		vo := reflect.ValueOf(input)
		if vo.Len() == 0 {
			return nil, fmt.Errorf("No slice items\n")
		}
		item := vo.Index(0)
		return getMessagepackTagString(item.Interface())
	case reflect.Struct:
		t := reflect.TypeOf(input)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get(MPDIRECTIVE)
			if tag != "" {
				result[field.Name] = tag
			}
		}
	default:
		return nil, fmt.Errorf("unknown kind of %v", rt.Kind())
	}
	return result, nil
}

func splitTags(s string) map[string]string {
	arr := strings.Split(s, ",")

	result := make(map[string]string)

	for _, a := range arr {
		nameVal := strings.Split(a, "=")
		if len(nameVal) == 1 {
			result[MPDIRECTIVE_NAME] = nameVal[0]
		} else {
			result[nameVal[0]] = nameVal[1]
		}
	}
	return result
}

func SplurtsArrToMessagepack(pw splurts.PiecewiseFloats, input interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	//TODO consume less memory, get variable by variable...optimize later when unit tests are done
	valmap, errmap := pw.GetValuesToFloatMapArr(input)
	if errmap != nil {
		return nil, errmap
	}

	tagsmap, errnamemap := getMessagepackTagString(input)
	if errnamemap != nil {
		return nil, errnamemap
	}

	variablecount := 0
	for _, p := range pw {
		if p.Omit || p.ConstDefined { //skip omits and consts
			continue
		}

		arr, haz := valmap[p.Name]
		if !haz {
			return nil, fmt.Errorf("name %s not found", p.Name)
		}

		overrideName := ""
		deltaCoding := int64(0)
		rleLimit := int64(0)
		tag, haztag := tagsmap[p.Name]
		if haztag {
			tagmap := splitTags(tag)
			overrideName, _ = tagmap[MPDIRECTIVE_NAME]
			deltaString, hazDelta := tagmap[MPDIRECTIVE_DELTA]
			rleString, hazRLE := tagmap[MPDIRECTIVE_RLE]
			if hazDelta {
				var errParseDelta error //TODO TOO MUCH REPEAT, generalize
				deltaCoding, errParseDelta = strconv.ParseInt(deltaString, 10, 64)
				if errParseDelta != nil {
					return nil, errParseDelta
				}
			}
			if hazRLE {
				var errParseRLE error
				rleLimit, errParseRLE = strconv.ParseInt(rleString, 10, 64)
				if errParseRLE != nil {
					return nil, errParseRLE
				}
			}

		}

		err := WriteSplurtsVariableArr(b, p, int(rleLimit), int(deltaCoding), overrideName, arr)
		if err != nil {
			return nil, err
		}
		variablecount++
	}

	result := new(bytes.Buffer)
	e := WriteFixmap(result, uint32(variablecount)) //TODO write after?
	if e != nil {
		return nil, e
	}
	_, e = result.Write(b.Bytes())
	if e != nil {
		return nil, e
	}

	return result.Bytes(), nil
}
