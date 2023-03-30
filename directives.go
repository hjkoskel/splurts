package splurts

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Keywords in struct. Fixed, based on what kind hardware measures and where
const (
	SPLURTS          = "splurts"
	DIRECTIVECLAMPED = "clamped"
	DIRECTIVEMIN     = "min"
	DIRECTIVEMAX     = "max"
	DIRECTIVESTEP    = "step"
	DIRECTIVESTEPS   = "steps"
	DIRECTIVEBITS    = "bits"   //Use instead of step or steps
	DIRECTIVEENUM    = "enum"   //Used for string datatypes, array of strings of names
	DIRECTIVEINFPOS  = "infpos" //Override inf+ value
	DIRECTIVEINFNEG  = "infneg" //Override inf- value
	DIRECTIVECONST   = "const"  //constant value, set when splurtsing to binary. Required when converting to binary
	DIRECTIVEOMIT    = "omit"   //do not splurt or unsplurt this variable

	DIRECTIVE_META_UNIT    = "unit"    //Unit like kg. Used when plotting and grouping "compatible" metrics together
	DIRECTIVE_META_CAPTION = "caption" //Caption for this metric, optional. Printable text without unit
	//DIRECTIVE_META_THING       = "thing"       //What thing this measures  "ambient", "gas", "oil"...  "inlet","cell" etc.. for grouping
	DIRECTIVE_META_ACCURACY    = "accuracy"    //Numerical value or string pointing to other field by name Plus minus value
	DIRECTIVE_META_MAXINTERVAL = "maxinterval" //During data loggin time NaN are indication that data is missing. This limit tells that it is declared faulty measurement instead of slower rate
	DIRECTIVE_META_BANDWIDTH   = "bandwidth"
)

const (
	DEFAULT_MINEPOCHMS = 0 //1600000000000
	DEFAULT_MAXEPOCHMS = 4300000000000
)

type DirectiveMetadata struct { //Metadata for plotting, printing, reporting, exporting etc... Not directly for computing values
	Unit    string
	Caption string
	// Thing    string NOT YET, think about this
	Accuracy string //Value or metric name

	//Interpolation allowed? If not then step...  linear?  spline
	//NOT YET, think about thisNoInterpolate bool //true.. means that actually this updates when it says so. And guessing values in between is bad by default

	//TODO RLE coding with extra parameter telling how many real updates!
	//NOT YET, think about this MinInterval time.Duration //Nanosec duration what is smallest sampling rate... PARSED FROM STRING  h,min,s,ms
	//NOT YET, think about this Interval    time.Duration //Norminal interval, what should be
	MaxInterval time.Duration //Nanosec duration how long before line cuts in time series when there are NaNs. If not defined NaN terminates immediately
	Bandwidth   float64       //-3dB point, see how fast transients are possible to catch. Is
}

type DirectiveSettings struct { //For parsed
	Omit    bool
	Clamped bool
	Min     float64
	Max     float64
	Bits    int
	Step    float64
	Steps   []PiecewiseCodingStep

	MinDefined bool
	MaxDefined bool

	InfPosDefined bool
	InfNegDefined bool
	InfPos        float64
	InfNeg        float64

	Enums []string
	Const string

	Meta DirectiveMetadata
}

// StepCount, if can not calc then 0
func (p *DirectiveSettings) StepCount() uint64 {
	count := uint64(0)
	for _, a := range p.Steps {
		count += a.Count
	}
	return count
}

func parseByTypenameToFloat64(s string, typename string) (float64, error) {
	switch typename {
	case "float64":
		return strconv.ParseFloat(s, 64)
	case "float32":
		return strconv.ParseFloat(s, 32)
	case "int":
		i, errconst := strconv.ParseUint(s, 0, 32)
		return float64(i), errconst
	case "int8":
		i, errconst := strconv.ParseUint(s, 0, 8)
		return float64(i), errconst
	case "int16":
		i, errconst := strconv.ParseUint(s, 0, 16)
		return float64(i), errconst
	case "int32":
		i, errconst := strconv.ParseUint(s, 0, 32)
		return float64(i), errconst
	case "int64":
		i, errconst := strconv.ParseUint(s, 0, 64)
		return float64(i), errconst
	case "bool":
		b, errconst := strconv.ParseBool(s)
		if b {
			return 1, errconst
		}
		return 0, errconst
	default:
		return 0, fmt.Errorf("trying set const to unknow typename %s", typename)
	}
}

func parseDirectives(tag string, typename string) (DirectiveSettings, error) {
	result := DirectiveSettings{}

	if typename == "Time" {
		//Lets set some good defaults
		//result.Bits = 64
		result.Step = 1.0
		result.Clamped = true
		result.Min = DEFAULT_MINEPOCHMS
		result.Max = DEFAULT_MAXEPOCHMS
		if len(tag) == 0 {
			return result, nil //Ok to have no directives for time.Time
		}
	}
	if len(tag) == 0 {
		return DirectiveSettings{}, fmt.Errorf("no directives") //bool does not need directives. It is just bit
	}

	maintokens := strings.Split(tag, ",")

	//Default step is 1
	result.Step = 1.0
	for tokindex, tok := range maintokens {
		eqsplit := strings.Split(tok, "=")
		if (len(eqsplit) != 2) && (len(eqsplit) != 1) {
			return result, fmt.Errorf("invalid tag %v", tag)
		}

		if eqsplit[0] == DIRECTIVEENUM {
			result.Enums = maintokens[tokindex:]
			result.Enums[0] = strings.Replace(result.Enums[0], DIRECTIVEENUM, "", 1)
			result.Enums[0] = strings.Replace(result.Enums[0], "=", "", 1)
			result.Steps = []PiecewiseCodingStep{
				{Size: 1, Count: uint64(len(result.Enums))},
			}
			result.Clamped = true
			return result, nil
		}

		if len(eqsplit) == 1 {
			switch eqsplit[0] { //Too many
			case DIRECTIVECLAMPED:
				result.Clamped = true
			case DIRECTIVEOMIT:
				result.Omit = true
			default:
				return result, fmt.Errorf("invalid tag %v, unknown token %v", tag, tok)
			}
		}
		if len(eqsplit) == 2 {
			switch eqsplit[0] {
			case DIRECTIVECONST:
				result.Const = eqsplit[1]
				//Check validity of constant
				_, errconst := parseByTypenameToFloat64(result.Const, typename)
				if errconst != nil {
					return result, fmt.Errorf("type %v can not have constant %v, err=%v", typename, result.Const, errconst.Error())
				}

			case DIRECTIVEINFPOS:
				f, ferr := strconv.ParseFloat(eqsplit[1], 64)
				if ferr != nil {
					return result, fmt.Errorf("invalid tag %v, invalid token %v parse error %v", tag, tok, ferr.Error())
				}
				result.InfPos = f
				result.InfPosDefined = true

			case DIRECTIVEINFNEG:
				f, ferr := strconv.ParseFloat(eqsplit[1], 64)
				if ferr != nil {
					return result, fmt.Errorf("invalid tag %v, invalid token %v parse error %v", tag, tok, ferr.Error())
				}
				result.InfNeg = f
				result.InfNegDefined = true

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
			case DIRECTIVE_META_UNIT: // = "unit"         //Unit like kg. Used when plotting and grouping "compatible" metrics together
				result.Meta.Unit = eqsplit[1]
			case DIRECTIVE_META_CAPTION: // = "caption"   //Caption for this metric, optional. Printable text without unit
				result.Meta.Caption = eqsplit[1]
			//case DIRECTIVE_META_THING: // =  "thing"       //What thing this measures  "ambient", "gas", "oil"...  "inlet","cell" etc.. for grouping
			//	result.Meta.Thing = eqsplit[1]
			case DIRECTIVE_META_ACCURACY: // = "accuracy" //Numerical value or string pointing to other field by name Plus minus value
				result.Meta.Accuracy = eqsplit[1]
			case DIRECTIVE_META_MAXINTERVAL:
				var err error
				result.Meta.MaxInterval, err = time.ParseDuration(eqsplit[1])
				if err != nil {
					return result, err
				}
			}
		}
	}

	//Sanity check
	if result.Clamped && (result.InfPosDefined || result.InfNegDefined) {
		return result, fmt.Errorf("clamped and infpos/infneg can not be defined at same time")
	}
	return result, nil
}

func createPiecewiseCodingFromStruct(name string, typename string, tag string) (PiecewiseCoding, error) {
	if typename == "bool" { //Bool does not need any other directives
		return PiecewiseCoding{Name: name, Min: 0, Steps: []PiecewiseCodingStep{
			{Size: 1, Count: 2},
		}, Clamped: true}, nil
	}

	dir, dirErr := parseDirectives(tag, typename)
	if dirErr != nil {
		return PiecewiseCoding{}, fmt.Errorf("%v fail %v", name, dirErr.Error())
	}

	result := PiecewiseCoding{
		Omit:    dir.Omit,
		Name:    name,
		Min:     dir.Min,
		Steps:   dir.Steps,
		Clamped: dir.Clamped,
		Enums:   dir.Enums,

		InfPosDefined: dir.InfPosDefined,
		InfNegDefined: dir.InfNegDefined,
		InfPos:        dir.InfPos,
		InfNeg:        dir.InfNeg,

		//Const: dir.Const,
		Meta: dir.Meta,
	}

	if 0 < len(dir.Const) {
		var parseErr error
		result.Const, parseErr = parseByTypenameToFloat64(dir.Const, typename)
		if parseErr != nil {
			return result, parseErr
		}
		result.ConstDefined = true
	}

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

// GetPiecewisesFromStruct parses by reflect all datatypes with directives to PiecewiseFloats
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

// setValuesFromFloatMap is helper function
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
			case "Time":
				timeParsed := time.UnixMilli(int64(inputValue))
				f.Set(reflect.ValueOf(timeParsed))

			default:
				return fmt.Errorf("unknown type %v", f.Type().Name())
			}
		}
	}
	return nil
}

func (p *PiecewiseFloats) GetValuesToFloatMapArr(v interface{}) (map[string][]float64, error) {
	result := make(map[string][]float64)

	rt := reflect.TypeOf(v)

	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		vo := reflect.ValueOf(v)
		count := vo.Len()
		for i := 0; i < count; i++ {
			item := vo.Index(i)
			m, errm := p.GetValuesToFloatMapArr(item.Interface())
			if errm != nil {
				return result, errm
			}
			for name, varvalue := range m {
				a, haz := result[name]
				if haz {
					a = append(a, varvalue...)
					result[name] = a
				} else {
					result[name] = varvalue
				}
			}
		}
	case reflect.Struct:
		m, errm := p.GetValuesToFloatMap(v)
		if errm != nil {
			return nil, errm
		}
		for name, varvalue := range m {
			result[name] = []float64{varvalue}
		}
	}
	return result, nil
}

func (p *PiecewiseFloats) GetValuesToFloatMap(v interface{}) (map[string]float64, error) {
	result := make(map[string]float64)
	elem := reflect.ValueOf(v)
	typeOfS := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)

		if f.IsValid() {

			name := typeOfS.Field(i).Name

			pw, foundErr := p.getCoding(name)
			if foundErr != nil {
				return result, foundErr
			}

			if pw.Omit {
				continue
			}

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
						return result, fmt.Errorf("unknown enum %s for %s (valid enums are %#v)", stringvalue, name, pw.Enums)
					}
				}
			case "Time":
				timevalue := f.Interface().(time.Time)
				//timevalue := reflect.ValueOf(f).Interface().(time.Time)
				result[name] = float64(timevalue.UnixMilli()) //Lets use milliseconds for compatibility and overflow reasons with javascript
			default:
				return result, fmt.Errorf("unknown type %s at %v", f.Type().Name(), f)
			}

			f := result[name]
			if pw.InfPosDefined && math.IsInf(f, 1) {
				result[name] = pw.InfPos
			}
			if pw.InfNegDefined && math.IsInf(f, -1) {
				result[name] = pw.InfNeg
			}
			if pw.ConstDefined {
				result[name] = pw.Const
			}
		}
	}
	return result, nil
}

// UnSplurts converts byte array to wanted target struct  (remember &output when call)
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

// Splurts data to bytes  Get piecewiseFloatStruct by calling GetPiecewisesFromStruct
func (p *PiecewiseFloats) Splurts(input interface{}) ([]byte, error) {
	m, e := p.GetValuesToFloatMap(input)
	if e != nil {
		return []byte{}, e
	}
	result, errEncode := p.Encode(m)
	if errEncode != nil {
		return result, fmt.Errorf("encoding error %v", errEncode.Error())
	}
	return result, nil
}

func (p *PiecewiseFloats) SplurtsHex(input interface{}) (string, error) {
	m, e := p.GetValuesToFloatMap(input)
	if e != nil {
		return "", e
	}
	return p.EncodeToHex(m)
}

func (p *PiecewiseFloats) SplurtsHexNybble(input interface{}) (string, error) {
	m, e := p.GetValuesToFloatMap(input)
	if e != nil {
		return "", e
	}
	return p.EncodeToHexNybble(m)
}

func (p *PiecewiseFloats) Splurts7bitBytes(input interface{}) (SevenBitArr, error) {
	m, e := p.GetValuesToFloatMap(input)
	if e != nil {
		return nil, e
	}
	return p.Encode7bitBytes(m)
}

// ToStrings writes values with required number of decimals and enums in string format
func (p *PiecewiseFloats) ToStrings(input interface{}, quotes bool) (map[string]string, error) {
	result := make(map[string]string)
	m, e := p.GetValuesToFloatMap(input)
	if e != nil {
		return nil, e
	}

	for _, pw := range *p {
		v, haz := m[pw.Name]
		if !haz {
			return nil, fmt.Errorf("internal error name %s in PiecewiseFloats is not found as value", pw.Name)
		}
		sval, svalErr := pw.ToStringValue(v)
		if svalErr != nil {
			return nil, fmt.Errorf("variable %v with value %v conversion to string fail err=%v", pw.Name, v, svalErr.Error())
		}
		if 0 < len(pw.Enums) && quotes {
			result[pw.Name] = "\"" + sval + "\""
		} else {
			result[pw.Name] = sval
		}
	}
	return result, nil
}

func (p *PiecewiseFloats) Names() []string {
	result := make([]string, len(*p))
	for i, a := range *p {
		result[i] = a.Name
	}
	return result
}

func (p *PiecewiseFloats) ToCsv(input interface{}, separator string, columns []string, skipNaNRows bool) (string, error) {
	rt := reflect.TypeOf(input)

	if len(columns) == 0 {
		columns = p.Names()
	}
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		vo := reflect.ValueOf(input)
		count := vo.Len()
		var sb strings.Builder
		for i := 0; i < count; i++ {
			item := vo.Index(i)
			s, csvErr := p.ToCsv(item.Interface(), separator, columns, skipNaNRows)
			if csvErr != nil {
				return sb.String(), fmt.Errorf("failed on row %v err=%v", i, csvErr.Error())
			}
			if 0 < len(s) {
				_, errWrite := sb.WriteString(s + "\n")
				if errWrite != nil {
					return "", errWrite
				}
			}
		}
		return sb.String(), nil

	case reflect.Struct:
		valuemap, errStrings := p.ToStrings(input, false)
		if errStrings != nil {
			return "", errStrings
		}
		var sb strings.Builder
		for _, name := range columns {
			v, haz := valuemap[name]
			if !haz {
				return "", fmt.Errorf("internal error no name %v in map %#v", v, valuemap)
			}
			if v == "NaN" && skipNaNRows {
				return "", nil //Skip this
			}

			if 0 < len(v) {
				if 0 < sb.Len() {
					sb.WriteString(separator)
				}
				sb.WriteString(v)
			}

		}
		return sb.String(), nil
	}
	return "", fmt.Errorf("not supported %v", rt.Kind())
}
