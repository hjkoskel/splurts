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
)

//Returns if there is problem in coding
func createPiecewiseCodingFromStruct(name string, typename string, tag string) (PiecewiseCoding, error) {
	result := PiecewiseCoding{Name: name}

	maintokens := strings.Split(tag, ",")
	//no tokens if empty string
	if len(tag) == 0 {
		maintokens = []string{}
	}
	parValues := make(map[string]float64)

	directSteps := []PiecewiseCodingStep{}

	for _, tok := range maintokens {
		eqsplit := strings.Split(tok, "=")

		if (len(eqsplit) != 2) && (len(eqsplit) != 1) {
			return result, fmt.Errorf("invalid tag on %v tag %v", name, tag)
		}
		if len(eqsplit) == 1 {
			switch eqsplit[0] { //Too many
			case DIRECTIVECLAMPED:
				result.Clamped = true
			default:
				return result, fmt.Errorf("invalid tag on %v tag %v, unknown token %v", name, tag, tok)
			}
		}
		if len(eqsplit) == 2 {
			switch eqsplit[0] {
			case DIRECTIVEMIN, DIRECTIVEMAX, DIRECTIVESTEP:
				f, ferr := strconv.ParseFloat(eqsplit[1], 64)
				if ferr != nil {
					return result, fmt.Errorf("invalid tag on %v tag %v, invalid token %v", name, tag, tok)
				}
				parValues[eqsplit[0]] = f
			case DIRECTIVESTEPS:
				sSteps := strings.Split(eqsplit[1], "|")
				for nStep, sStep := range sSteps {
					sizecountArr := strings.Split(sStep, " ")
					if len(sizecountArr) != 2 {
						return result, fmt.Errorf("invalid tag on %v tag %v, invalid token %v Only size|count pairs NOT %#v", name, tag, tok, sStep)
					}
					stepSize, parseErrSize := strconv.ParseFloat(sizecountArr[0], 64)
					stepCount, parseErrCount := strconv.ParseInt(sizecountArr[1], 10, 64)
					if parseErrSize != nil || parseErrCount != nil {
						return result, fmt.Errorf("invalid tag on %v tag %v, invalid token %v  error on token step %v", name, tag, tok, nStep)
					}
					directSteps = append(directSteps, PiecewiseCodingStep{Size: stepSize, Count: uint64(stepCount)})
				}
			}
		}
	}

	//Is it basic
	dirMin, hazMin := parValues[DIRECTIVEMIN]
	dirMax, hazMax := parValues[DIRECTIVEMAX]
	dirStep, hazStep := parValues[DIRECTIVESTEP]

	result.Steps = directSteps
	result.Min = dirMin

	switch typename {
	case "float64", "float32": //required min, max and step
		if hazMin && hazMax && hazStep {
			//Basic config
			result.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Size: dirStep, Count: uint64(math.Ceil((dirMax - dirMin) / dirStep))}}
		} else {
			if len(result.Steps) == 0 || !hazMin {
				return result, fmt.Errorf("float type requires \"min\", \"max\", and \"step\" OR \"min\" and \"steps\" directives at %v", name)
			}
		}
	case "int", "int8", "int16", "int32", "int64": //Stepsize is optional. Default one because integer
		//WARNING uint32 and uint64 can lose "steps" when doing float conversion.
		if hazMin && hazMax {
			//Basic config
			if !hazStep {
				dirStep = 1.0
			}
			result.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Size: dirStep, Count: uint64(math.Ceil((dirMax - dirMin) / dirStep))}}
		} else {
			if len(result.Steps) == 0 || !hazMin {
				return result, fmt.Errorf("int type requires \"min\" and \"max\" OR \"min\" and \"steps\" directives at %v", name)
			}
		}
	case "uint", "uint8", "uint16", "uint32", "uint64": //min default is zero
		//WARNING uint32 and uint64 can lose "steps" when doing float conversion.
		if hazMax {
			//Basic config
			if !hazStep {
				dirStep = 1.0
			}
			if !hazMin {
				dirMin = 0
			}
			result.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Size: dirStep, Count: uint64(math.Ceil((dirMax - dirMin) / dirStep))}}
		}
	case "bool": //Special treatment only two values, goes without directives
		result.Clamped = true
		result.Min = 0
		result.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Size: 1, Count: 2}}
	default:
		return result, fmt.Errorf("Type %v on %v is not supported", typename, name)
	}

	//TODO call valid check
	return result, result.IsInvalid()
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
func setValuesFromFloatMap(v interface{}, values map[string]float64) error {
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
			}
		}
	}
	return nil
}

func getValuesToFloatMap(v interface{}) map[string]float64 {
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
			}
		}
	}
	return result
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
	return setValuesFromFloatMap(output, variableMap)
}

//Splurts data to bytes  Get piecewiseFloatStruct by calling GetPiecewisesFromStruct
func (p *PiecewiseFloats) Splurts(input interface{}) ([]byte, error) {
	return p.Encode(getValuesToFloatMap(input))
}
