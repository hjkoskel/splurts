package splurts

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

//Keywords in struct
const (
	SPLURTS          = "splurts"
	DIRECTIVECLAMPED = "clamped"
	DIRECTIVEMIN     = "min"
	DIRECTIVEMAX     = "max"
	DIRECTIVESTEP    = "step"
	DIRECTIVESTEPS   = "steps"
	DIRECTIVEBITS    = "bits" //Use instead of step or steps
	DIRECTIVEENUM    = "enum" //Used for string datatypes, array of strings of names
)

type DirectiveSettings struct { //For parsed
	Clamped bool
	Min     float64
	Max     float64
	Bits    int
	Step    float64
	Steps   []PiecewiseCodingStep

	MinDefined bool
	MaxDefined bool

	Enums []string
}

// StepCount, if can not calc then 0
func (p *DirectiveSettings) StepCount() uint64 {
	count := uint64(0)
	for _, a := range p.Steps {
		count += a.Count
	}
	return count
}

/*
func MaxBitsForType(typename string) (int, error) {
	bitsByType, hazType := map[string]int{
		"float64": 64, "int64": 64, "uint64": 64, "uint": 64, "int": 64,
		"float32": 32, "int32": 32, "uint32": 32,
		"int16": 16, "uint16": 16,
		"int8": 8, "uint8": 8,
		"bool": 1,
	}[typename]
	if !hazType {
		return 0, fmt.Errorf("Unknown type %v", typename)
	}
	return bitsByType, nil
}*/

func parseDirectives(tag string) (DirectiveSettings, error) {
	if len(tag) == 0 {
		return DirectiveSettings{}, fmt.Errorf("No directives") //bool does not need directives. It is just bit
	}
	maintokens := strings.Split(tag, ",")
	result := DirectiveSettings{}

	//Default step is 1
	result.Step = 1.0
	for tokindex, tok := range maintokens {
		eqsplit := strings.Split(tok, "=")
		if (len(eqsplit) != 2) && (len(eqsplit) != 1) {
			return result, fmt.Errorf("invalid tag %v", tag)
		}

		if eqsplit[0] == DIRECTIVEENUM {
			result.Enums = maintokens[tokindex : len(maintokens)-1]
			result.Enums[0] = strings.Replace(result.Enums[0], DIRECTIVEENUM, "", 1)
			result.Enums[0] = strings.Replace(result.Enums[0], "=", "", 1)
			result.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Size: 1, Count: uint64(len(result.Enums))}}
			result.Clamped = true
			return result, nil
		}

		if len(eqsplit) == 1 {
			switch eqsplit[0] { //Too many
			case DIRECTIVECLAMPED:
				result.Clamped = true
			default:
				return result, fmt.Errorf("invalid tag %v, unknown token %v", tag, tok)
			}
		}
		if len(eqsplit) == 2 {
			switch eqsplit[0] {
			case DIRECTIVEMIN:
				f, ferr := strconv.ParseFloat(eqsplit[1], 64)
				if ferr != nil {
					return result, fmt.Errorf("invalid tag %v, invalid token %v parse error %v", tag, tok, ferr.Error())
				}
				result.Min = f
				result.MinDefined = true
			case DIRECTIVEMAX:
				f, ferr := strconv.ParseFloat(eqsplit[1], 64)
				if ferr != nil {
					return result, fmt.Errorf("invalid tag %v, invalid token %v parse error %v", tag, tok, ferr.Error())
				}
				result.Max = f
				result.MaxDefined = true
			case DIRECTIVESTEP:

				f, ferr := strconv.ParseFloat(eqsplit[1], 64)
				if ferr != nil {
					return result, fmt.Errorf("invalid tag %v, invalid token %v parse error %v", tag, tok, ferr.Error())
				}
				result.Step = f

			case DIRECTIVEBITS:
				bts, parseError := strconv.ParseInt(eqsplit[1], 10, 8)
				if parseError != nil {
					return result, fmt.Errorf("invalid tag %v, invalid token %v", tag, tok)
				}
				//Invalidity check
				if bts < 1 {
					return result, fmt.Errorf("invalid tag %v, invalid token %v, negative bits used", tag, tok)
				}
				result.Bits = int(bts)
			case DIRECTIVESTEPS:
				sSteps := strings.Split(eqsplit[1], "|")
				for nStep, sStep := range sSteps {
					sizecountArr := strings.Split(sStep, " ")
					if len(sizecountArr) != 2 {
						return result, fmt.Errorf("invalid tag %v, invalid token %v Only size|count pairs NOT %#v", tag, tok, sStep)
					}
					stepSize, parseErrSize := strconv.ParseFloat(sizecountArr[0], 64)
					stepCount, parseErrCount := strconv.ParseInt(sizecountArr[1], 10, 64)
					if parseErrSize != nil || parseErrCount != nil {
						return result, fmt.Errorf("invalid tag %v, invalid token %v  error on token step %v", tag, tok, nStep)
					}
					result.Steps = append(result.Steps, PiecewiseCodingStep{Size: stepSize, Count: uint64(stepCount)})
				}
			}
		}
	}

	return result, nil
}

func createPiecewiseCodingFromStruct(name string, typename string, tag string) (PiecewiseCoding, error) {
	if typename == "bool" { //Bool does not need any other directives
		return PiecewiseCoding{Name: name, Min: 0, Steps: []PiecewiseCodingStep{PiecewiseCodingStep{Size: 1, Count: 2}}, Clamped: true}, nil
	}

	dir, dirErr := parseDirectives(tag)
	if dirErr != nil {
		return PiecewiseCoding{}, fmt.Errorf("%v fail %v", name, dirErr.Error())
	}

	result := PiecewiseCoding{Name: name, Min: dir.Min, Steps: dir.Steps, Clamped: dir.Clamped, Enums: dir.Enums}
	if 0 < len(result.Steps) {
		return result, nil //OK
	}

	if typename == "string" && 0 < len(dir.Enums) {
		return result, nil
	}

	steps := int(0)
	if 1 < dir.Bits { //ok, calc from bits
		if result.Clamped {
			//is already clamped
			steps = int(math.Pow(2, float64(dir.Bits)))
		} else {
			//NaN, -inf and +inf are needed
			steps = int(math.Pow(2, float64(dir.Bits))) - 3
		}
		if !dir.MaxDefined {
			dir.Max = dir.Step*float64(steps) + dir.Min
		}
		dir.Step = (dir.Max - dir.Min) / float64(steps)
	}

	oneStep := PiecewiseCodingStep{Size: dir.Step, Count: uint64(math.Ceil((dir.Max - dir.Min) / dir.Step))}
	result.Steps = []PiecewiseCodingStep{oneStep}
	return result, nil
}

//GetPiecewisesFromStruct parses by reflect all datatypes with directives to PiecewiseFloats
func GetPiecewisesFromStruct(v interface{}) (PiecewiseFloats, error) {
	result := []PiecewiseCoding{}
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Name() != "" {
			coding, codingErr := createPiecewiseCodingFromStruct(field.Name, field.Type.Name(), field.Tag.Get(SPLURTS))
			if codingErr != nil {
				return result, codingErr
			}
			result = append(result, coding)
		}
	}
	return result, nil
}

//setValuesFromFloatMap is helper function
func (p *PiecewiseFloats) setValuesFromFloatMap(v interface{}, values map[string]float64) error {
	elem := reflect.ValueOf(v).Elem()
	for name, inputValue := range values {
		f := elem.FieldByName(name)
		if f.IsValid() {
			switch f.Type().Name() {
			case "float64", "float32":
				f.SetFloat(inputValue)
			case "int", "int8", "int16", "int32", "int64":
				f.SetInt(int64(inputValue))
			case "uint", "uint8", "uint16", "uint32", "uint64":
				f.SetUint(uint64(inputValue))
			case "bool":
				f.SetBool(0 < inputValue)
			case "string":
				pw, foundErr := p.getCoding(name)
				if foundErr != nil {
					return foundErr
				}
				index := int(inputValue)
				if index == 0 {
					f.SetString("") //First0 is empty/missign
				} else {
					if 0 <= index && index <= len(pw.Enums) {
						f.SetString(pw.Enums[index-1])
					}
				}
			}
		}
	}
	return nil
}

func (p *PiecewiseFloats) getValuesToFloatMap(v interface{}) (map[string]float64, error) {
	result := make(map[string]float64)
	elem := reflect.ValueOf(v)
	typeOfS := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)

		if f.IsValid() {
			name := typeOfS.Field(i).Name
			switch f.Type().Name() {
			case "float64", "float32":
				result[name] = f.Float()
			case "int", "int8", "int16", "int32", "int64":
				result[name] = float64(f.Int())
			case "uint", "uint8", "uint16", "uint32", "uint64":
				result[name] = float64(f.Uint())
			case "bool":
				result[name] = 0
				if f.Bool() {
					result[name] = 1
				}
			case "string":
				pw, foundErr := p.getCoding(name)
				if foundErr != nil {
					return result, foundErr
				}
				stringvalue := f.String()

				if stringvalue == "" {
					result[name] = 0
				} else {
					found := false
					for indexresult, enumstring := range pw.Enums {
						if enumstring == stringvalue {
							result[name] = float64(indexresult + 1)
							found = true
							continue
						}
					}
					if !found {
						return result, fmt.Errorf("Unknown enum %s for %s (valid enums are %#v)", stringvalue, name, pw.Enums)
					}
				}
			}
		}
	}
	return result, nil
}

//UnSplurts converts byte array to wanted target struct  (remember &output when call)
func (p *PiecewiseFloats) UnSplurts(raw []byte, output interface{}) error {
	errInv := p.IsInvalid()
	if errInv != nil {
		return errInv
	}
	variableMap, errDecode := p.Decode(raw, true)
	if errDecode != nil {
		return errDecode
	}
	return p.setValuesFromFloatMap(output, variableMap)
}

func (p *PiecewiseFloats) UnSplurts7bitBytes(raw SevenBitArr, output interface{}) error {
	errInv := p.IsInvalid()
	if errInv != nil {
		return errInv
	}
	variableMap, errDecode := p.Decode7bitBytes(raw, true)
	if errDecode != nil {
		return errDecode
	}
	return p.setValuesFromFloatMap(output, variableMap)
}

//Splurts data to bytes  Get piecewiseFloatStruct by calling GetPiecewisesFromStruct
func (p *PiecewiseFloats) Splurts(input interface{}) ([]byte, error) {
	m, e := p.getValuesToFloatMap(input)
	if e != nil {
		return []byte{}, e
	}
	result, errEncode := p.Encode(m)
	if errEncode != nil {
		return result, fmt.Errorf("Encoding error %v", errEncode.Error())
	}
	return result, nil
}

func (p *PiecewiseFloats) SplurtsHex(input interface{}) (string, error) {
	m, e := p.getValuesToFloatMap(input)
	if e != nil {
		return "", e
	}
	return p.EncodeToHex(m)
}

func (p *PiecewiseFloats) SplurtsHexNybble(input interface{}) (string, error) {
	m, e := p.getValuesToFloatMap(input)
	if e != nil {
		return "", e
	}
	return p.EncodeToHexNybble(m)
}

func (p *PiecewiseFloats) Splurts7bitBytes(input interface{}) (SevenBitArr, error) {
	m, e := p.getValuesToFloatMap(input)
	if e != nil {
		return nil, e
	}
	return p.Encode7bitBytes(m)
}
