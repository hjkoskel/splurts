package messagepack

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

// Like recipe... create once. Update data..write..read...search
type MetricArr struct { //One entry from splurts struct
	Meta   MetricMeta
	Coding MetricCoding

	Enums []string
	Steps []MetricStep

	Delta int         //How many deltas
	Data  DeltaRLEVec //TODO WITHOUT COUNT HOW MANY POINTS
}

func (p *MetricArr) Write(buf *bytes.Buffer) error {
	WriteFixmap(buf, 4)
	err := WriteString(buf, MPNAME_META) //1
	if err != nil {
		return err
	}
	err = p.Meta.Write(buf)
	if err != nil {
		return err
	}

	err = WriteString(buf, MPNAME_CODING) //2
	if err != nil {
		return err
	}
	p.Coding.Write(buf)

	if 0 < len(p.Enums) { //3
		err = WriteString(buf, MPNAME_ENUMS)
		if err != nil {
			return err
		}
		WriteArray(buf, uint32(len(p.Enums)))
		for _, s := range p.Enums {
			WriteString(buf, s)
		}
	} else {
		WriteString(buf, MPNAME_STEPS)
		if len(p.Steps) == 0 {
			return fmt.Errorf("no steps defined")
		}
		WriteArray(buf, uint32(len(p.Steps)))
		for _, step := range p.Steps {
			step.Write(buf)
		}
	}

	switch p.Delta { //4
	case 0:
		WriteString(buf, MPCODING_SIMPLEARR)
	case 1:
		WriteString(buf, MPCODING_DELTA1ARR)
	case 2:
		WriteString(buf, MPCODING_DELTA2ARR)
	default:
		return fmt.Errorf("delta %v not supported", p.Delta)
	}
	p.Data.WriteToBuf(buf)
	return nil
}

func ReadMetricArr(buf io.Reader) (MetricArr, error) {
	itemCount, errMap := ReadFixmap(buf)
	if errMap != nil {
		return MetricArr{}, errMap
	}
	result := MetricArr{}

	hadCoding := false
	hadEnums := false
	hadSteps := false
	hadData := false
	hadMeta := false

	for i := 0; i < int(itemCount); i++ {
		itemname, err := ReadString(buf)
		if err != nil {
			return MetricArr{}, fmt.Errorf("error %s on ReadMetricArr item name", err.Error())
		}

		switch itemname {
		case MPNAME_CODING:
			if hadCoding {
				return result, fmt.Errorf("already had coding")
			}
			result.Coding, err = ReadMetricCoding(buf)
			if err != nil {
				return result, err
			}
			hadCoding = true
		case MPNAME_ENUMS:
			if hadEnums {
				return result, fmt.Errorf("already had enums")
			}
			namecount, err := ReadArr(buf)
			if err != nil {
				return result, err
			}
			result.Enums = make([]string, namecount)
			for j := 0; j < int(namecount); j++ {
				result.Enums[j], err = ReadString(buf)
				if err != nil {
					return result, fmt.Errorf("%s error on ReadMetric arr while reading enums", err.Error())
				}
			}
			hadEnums = true
		case MPNAME_META:
			if hadMeta {
				return result, fmt.Errorf("already had meta")
			}
			result.Meta, err = ReadMetricMeta(buf)
			if err != nil {
				return result, err
			}
			hadMeta = true
		case MPNAME_STEPS:
			if hadSteps {
				return result, fmt.Errorf("already had steps")
			}
			stepcount, errStepCount := ReadArr(buf)
			if errStepCount != nil {
				return result, errStepCount
			}
			result.Steps = make([]MetricStep, stepcount)
			var errStep error
			for j := 0; j < int(stepcount); j++ {
				result.Steps[j], errStep = ReadMetricStep(buf)
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
			result.Data, errRead = ReadDeltaRLEVec(buf)
			if errRead != nil {
				return result, fmt.Errorf("error reading %s err=%s", itemname, errRead)
			}
			hadData = true
		}
	}

	return result, nil
}

// ValueAndStepsize, gets value and stepsize at that local part of curve (used for calculating required decimals)
func (p *MetricArr) ValueAndStep(reg int64) (float64, float64, error) {
	if len(p.Steps) == 0 {
		return float64(reg), 1, nil
	}
	counter := int64(0)
	total := float64(0)
	targetIndex := reg
	if !p.Coding.Clamped {
		if reg == 0 {
			return math.Inf(-1), 0, nil
		}
		targetIndex--
	}

	for _, step := range p.Steps {
		if targetIndex < counter+step.Count {
			return p.Coding.Min + total + float64(targetIndex-counter)*step.Step, step.Step, nil
		}
		counter += step.Count
		total += float64(step.Count) * step.Step
	}
	if p.Coding.Clamped {
		return 0, 0, fmt.Errorf("out of range, clamped reg=%v", reg)
	}
	if reg == counter {
		return math.Inf(1), 0, nil
	}
	return math.NaN(), 0, nil
}

func (p *MetricArr) AllValues() ([]float64, error) { //Crude way to just dump... start with this later optimized functions
	regarr, errArr := p.Data.ToArr(p.Delta)
	if errArr != nil {
		return nil, errArr
	}

	result := make([]float64, len(regarr))
	var err error
	for i, reg := range regarr {
		result[i], _, err = p.ValueAndStep(reg)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (p *MetricArr) TotalStepCount() int64 {
	result := int64(0)
	for _, step := range p.Steps {
		result += int64(step.Count)
	}
	return result
}

// for text based protocols, json, csv files
func (p *MetricArr) AllValuesAsString() ([]string, error) {
	regarr, errArr := p.Data.ToArr(p.Delta)
	if errArr != nil {
		return nil, fmt.Errorf("toArr: %v", errArr.Error())
	}
	result := make([]string, len(regarr))
	if 0 < len(p.Enums) {
		for i, reg := range regarr {
			if reg == 0 {
				result[i] = ""
			} else {
				if reg-1 < int64(len(p.Enums)) {
					result[i] = p.Enums[reg-1]
				} else {
					return nil, fmt.Errorf("out of range enum reg=%v got %v enums", reg, len(p.Enums))
				}
			}
		}
		return result, nil
	}

	for i, reg := range regarr {
		f, step, err := p.ValueAndStep(reg) //TODO more optimized...this is for initial testing
		if err != nil {
			return nil, err
		}
		formatstring := fmt.Sprintf("%%.%vf", int(math.Ceil(math.Abs(math.Log10(step)))))
		result[i] = fmt.Sprintf(formatstring, f)
	}
	return result, nil
}
