package messagepack

import (
	"fmt"
	"io"
	"math"
)

type MsgPackMetricCoding struct {
	Min     float64
	Max     float64
	Clamped bool
}

type MsgPackMetricStep struct {
	Count int64
	Step  float64
}

/*
Tää on tarpeen vaan luoda ja sitten käsitellä?

Luodaan messagepackistä

*/

//Like recipe... create once. Update data..write..read...search
type MsgPackMetricArr struct { //One entry from splurts struct
	//name string //comes from splurts directive
	//Variable name in splurts?
	//Other recipe stuff?

	Enums  []string
	Coding MsgPackMetricCoding
	Steps  []MsgPackMetricStep

	Delta int //How many deltas
	//IS FOUND FROM DELTAVEC, USED ONLY WHEN CREATING Rle   int //RLE limit. Or is used at all

	data DeltaRLEVec //TODO WITHOUT COUNT HOW MANY POINTS
}

func ReadMsgPackMetricStep(buf io.Reader) (MsgPackMetricStep, error) {
	result := MsgPackMetricStep{}
	n, errn := ReadFixmap(buf)
	if errn != nil {
		return result, nil
	}
	for i := 0; i < int(n); i++ {
		itemname, errItemName := ReadString(buf)
		if errItemName != nil {
			return MsgPackMetricStep{}, errItemName
		}
		var readErr error
		switch itemname {
		case MPNAME_POINTCOUNT:
			result.Count, readErr = ReadInt(buf)
		case MPNAME_STEPSIZE:
			result.Step, readErr = ReadNumber(buf)
		}
		if readErr != nil {
			return result, readErr
		}
	}
	return result, nil
}

func ReadMsgPackMetricCoding(buf io.Reader) (MsgPackMetricCoding, error) {
	result := MsgPackMetricCoding{}
	n, errn := ReadFixmap(buf)
	if errn != nil {
		return result, nil
	}
	for i := 0; i < int(n); i++ {
		itemname, errItemName := ReadString(buf)
		if errItemName != nil {
			return MsgPackMetricCoding{}, errItemName
		}
		var readErr error
		switch itemname {
		case MPNAME_CODING_MIN:
			result.Min, readErr = ReadNumber(buf)
		case MPNAME_CODING_MAX:
			result.Max, readErr = ReadNumber(buf)
		case MPNAME_CODING_CLAMPED:
			result.Clamped, readErr = ReadBool(buf)
		}
		if readErr != nil {
			return result, readErr
		}
	}
	return result, nil
}

func ReadMsgPackMetricArr(buf io.Reader) (MsgPackMetricArr, error) {
	itemCount, errMap := ReadFixmap(buf)
	if errMap != nil {
		return MsgPackMetricArr{}, errMap
	}
	result := MsgPackMetricArr{Enums: []string{}}

	hadCoding := false
	hadEnums := false
	hadSteps := false
	hadData := false

	for i := 0; i < int(itemCount); i++ {
		itemname, errItemName := ReadString(buf)
		if errItemName != nil {
			return MsgPackMetricArr{}, errItemName
		}

		switch itemname {
		case MPNAME_CODING:
			if hadCoding {
				return result, fmt.Errorf("already had coding")
			}
			var codingErr error
			result.Coding, codingErr = ReadMsgPackMetricCoding(buf)
			if codingErr != nil {
				return result, codingErr
			}
			hadCoding = true
		case MPNAME_ENUMS:
			if hadEnums {
				return result, fmt.Errorf("already had enums")
			}
			namecount, errNamecount := ReadArr(buf)
			if errNamecount != nil {
				return result, errNamecount
			}
			result.Enums = make([]string, namecount)
			var errS error
			for j := 0; j < int(namecount); j++ {
				result.Enums[j], errS = ReadString(buf)
				if errS != nil {
					return result, errS
				}
			}
			hadEnums = true

		case MPNAME_STEPS:
			if hadSteps {
				return result, fmt.Errorf("already had steps")
			}
			stepcount, errStepCount := ReadArr(buf)
			if errStepCount != nil {
				return result, errStepCount
			}
			result.Steps = make([]MsgPackMetricStep, stepcount)
			var errStep error
			for j := 0; j < int(stepcount); j++ {
				result.Steps[j], errStep = ReadMsgPackMetricStep(buf)
				if errStep != nil {
					return result, errStep
				}
			}
			hadSteps = true
		case MPCODING_SIMPLEARR:
			result.Delta = 0
		case MPCODING_DELTA1ARR:
			result.Delta = 1
		case MPCODING_DELTA2ARR:
			result.Delta = 2
		}

		if itemname == MPCODING_SIMPLEARR || itemname == MPCODING_DELTA1ARR || itemname == MPCODING_DELTA2ARR {
			if hadData {
				return result, fmt.Errorf("already had data delta=%v", result.Delta)
			}
			var errRead error
			result.data, errRead = ReadDeltaRLEVec(buf)
			if errRead != nil {
				return result, errRead
			}
			hadData = true
		}
	}

	return result, nil
}
func ReadMsgPackMetrics(buf io.Reader) (map[string]MsgPackMetricArr, error) {
	nMetrics, errNmetrics := ReadFixmap(buf)
	if errNmetrics != nil {
		return nil, errNmetrics
	}
	result := make(map[string]MsgPackMetricArr)
	for i := 0; i < int(nMetrics); i++ {
		name, nameErr := ReadString(buf)
		if nameErr != nil {
			return nil, nameErr
		}
		m, mErr := ReadMsgPackMetricArr(buf)
		if mErr != nil {
			return nil, mErr
		}
		result[name] = m
	}

	return result, nil
}

func (p *MsgPackMetricArr) Value(reg int64) (float64, error) {
	if len(p.Steps) == 0 {
		return float64(reg), nil
		//return 0, fmt.Errorf("no steps")
	}
	counter := int64(0)
	total := float64(0)
	targetIndex := reg
	if !p.Coding.Clamped {
		if reg == 0 {
			return math.Inf(-1), nil
		}
		targetIndex--
	}

	for _, step := range p.Steps {
		if targetIndex < counter+step.Count {
			return p.Coding.Min + total + float64(targetIndex-counter)*step.Step, nil
		}
		counter += step.Count
		total += float64(step.Count) * step.Step
	}
	if p.Coding.Clamped {
		return 0, fmt.Errorf("out of range")
	}
	if reg == counter {
		return math.Inf(1), nil
	}
	return math.NaN(), nil
}

func (p *MsgPackMetricArr) AllValues() ([]float64, error) { //Crude way to just dump... start with this later optimized functions
	regarr, errArr := p.data.ToArr(p.Delta)
	if errArr != nil {
		return nil, errArr
	}
	result := make([]float64, len(regarr))
	var err error
	for i, reg := range regarr {
		result[i], err = p.Value(reg)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

/*
//inputOneVar not actual arr. Or recursive is arr?
func CreateMsgPackMetricStruct(input interface{}) (MsgPackMetricStruct, error) {
	//todo get this as parameter? or is overhead neglible?
	pwArr, errPiecewises := splurts.GetPiecewisesFromStruct(input)

	result := make(map[string]MsgPackMetricArr)

	if errPiecewises != nil {
		return result, errPiecewises
	}

	t := reflect.TypeOf(input)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		r := MsgPackMetricArr{}

		tagstring := field.Tag.Get(MPDIRECTIVE)
		if tagstring != "" {
			tags := splitTags(tagstring)
			s, hazs := tags[MPDIRECTIVE_NAME]
			if hazs {
				r.Name = s
			}
			s, hazs = tags[MPDIRECTIVE_DELTA]
			if hazs {
				i, parseErr := strconv.ParseInt(s, 10, 64)
				if parseErr != nil {
					return result, fmt.Errorf("error parsing tagstring %s, %s have invalid value %v, err=%v", tagstring, MPDIRECTIVE_DELTA, s, parseErr.Error())
				}
				r.Delta = int(i)
			}

			s, hazs = tags[MPDIRECTIVE_RLE]
			if hazs {
				i, parseErr := strconv.ParseInt(s, 10, 64)
				if parseErr != nil {
					return result, fmt.Errorf("error parsing tagstring %s, %s have invalid value %v, err=%v", tagstring, MPDIRECTIVE_DELTA, s, parseErr.Error())
				}
				r.Rle = int(i)
			}

		}
		result[field.Name] = r
	}

	finalResult := make(map[string]MsgPackMetricArr)

	//TODO or call lower level function?
	for _, q := range pwArr {
		r, hazr := result[q.Name] //TODO NAME OR CAPTION? CHECK
		if !hazr {
			continue //NO FOR FINAL RESULT?
		}
		r.Coding.Clamped = q.Clamped
		r.Coding.Min = q.Min
		r.Coding.Max = q.Max()
		r.Enums = q.Enums
		r.Steps = make([]MsgPackMetricStep, len(q.Steps))

		finalResult[q.Name] = r
	}
	return finalResult, nil
}

//Set from array of structs
func (p *MsgPackMetricStruct) Set(arr interface{}) error {
	rt := reflect.TypeOf(arr)

	kind := rt.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		vo := reflect.ValueOf(arr)
		if vo.Len() == 0 {
			return fmt.Errorf("no slice items\n")
		}
		count := vo.Len()

		for i := 0; i < count; i++ {
			item := vo.Index(i)
			item
		}
	}

}
*/
