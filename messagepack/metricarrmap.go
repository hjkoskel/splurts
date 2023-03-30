/*
Metric array map.

Metricarrmap is created from struct or array of splurts structs
*/
package messagepack

import (
	"fmt"
	"io"
	"strings"

	"github.com/hjkoskel/splurts"
)

type MetricArrMap map[string]MetricArr

func (p *MetricArrMap) MetricNames() ([]string, error) {
	result := make([]string, len(*p))
	i := 0
	for name := range *p {
		result[i] = name
		i++
	}
	return result, nil
}

func (p *MetricArrMap) TabulateValues(colNames []string, separator string) (string, error) {
	if len(colNames) == 0 || colNames == nil {
		var errNames error
		colNames, errNames = p.MetricNames()
		if errNames != nil {
			return "", errNames
		}
	}
	if len(colNames) == 0 {
		return "", nil
	}

	ma := map[string]MetricArr(*p)
	resultData := map[string][]string{}

	var stringConvErr error
	for _, name := range colNames {
		arr, haz := ma[name]
		if !haz {
			return "", fmt.Errorf("column %s not found", name)
		}

		resultData[name], stringConvErr = arr.AllValuesAsString()
		if stringConvErr != nil {
			return "", fmt.Errorf("error converting %v err=%v", name, stringConvErr.Error())
		}
	}
	//Check lengths
	requiredLen := 0
	for name, a := range resultData {
		if requiredLen == 0 {
			requiredLen = len(a)
		}
		if len(a) != requiredLen {
			return "", fmt.Errorf("expected length of %v but %s is %v length", requiredLen, name, len(a))
		}
	}
	var sb strings.Builder
	for i := 0; i < requiredLen; i++ {
		cols := make([]string, len(colNames))
		for col, name := range colNames {
			cols[col] = resultData[name][i]
		}
		sb.WriteString(strings.Join(cols, separator) + "\n")
	}
	return sb.String(), nil
}

func ReadMetricsArrMap(buf io.Reader) (MetricArrMap, error) {
	nMetrics, errNmetrics := ReadFixmap(buf)
	if errNmetrics != nil {
		return nil, errNmetrics
	}
	result := make(map[string]MetricArr)
	for i := 0; i < int(nMetrics); i++ {
		name, nameErr := ReadString(buf)
		if nameErr != nil {
			return nil, nameErr
		}
		m, mErr := ReadMetricArr(buf)
		if mErr != nil {
			return nil, mErr
		}
		result[name] = m
	}
	return result, nil
}

func (p *MetricArrMap) Write(w io.Writer) error {
	err := WriteFixmap(w, uint32(len(*p)))
	if err != nil {
		return err
	}
	for metname, met := range *p {
		err = WriteString(w, metname)
		if err != nil {
			return err
		}

		err = met.Write(w)
		if err != nil {
			return err
		}
	}
	return nil
}

// SplurtsArrToMetricArrMap creates MetricArrMap
func SplurtsArrToMetricArrMap(pw splurts.PiecewiseFloats, input interface{}) (MetricArrMap, error) {
	valmap, errmap := pw.GetValuesToFloatMapArr(input)
	if errmap != nil {
		return nil, errmap
	}

	tagsmap, errnamemap := getMessagepackTagString(input)
	if errnamemap != nil {
		return nil, errnamemap
	}

	result := make(map[string]MetricArr)
	for _, p := range pw {
		if p.Omit || p.ConstDefined { //skip omits and consts
			continue
		}

		arr, haz := valmap[p.Name]
		if !haz {
			return nil, fmt.Errorf("name %s not found", p.Name)
		}

		tag := tagsmap[p.Name]
		packdirect, errDirective := parseDirectives(tag, p.Name)
		if errDirective != nil {
			return result, errDirective
		}

		entryMeta := MetricMeta{
			Unit:        p.Meta.Unit,
			Caption:     p.Meta.Caption,
			Accuracy:    p.Meta.Accuracy,
			MaxInterval: p.Meta.MaxInterval.Nanoseconds(),
			Bandwidth:   p.Meta.Bandwidth,
		}

		entrySteps := make([]MetricStep, len(p.Steps))
		for i, step := range p.Steps {
			entrySteps[i] = MetricStep{
				Count: int64(step.Count),
				Step:  step.Size,
			}
		}

		entry := MetricArr{
			Meta:   entryMeta,
			Enums:  p.Enums,
			Coding: MetricCoding{Min: p.Min, Max: p.Max(), Clamped: p.Clamped},
			Steps:  entrySteps,
			Delta:  int(packdirect.Delta),
		}

		var dataConvErr error
		//TODO int64 register values required!
		entry.Data, dataConvErr = CreateDeltaRLEVec(p.ScaleToIntArr(arr), entry.Delta, int(packdirect.Rle))
		if dataConvErr != nil {
			return result, dataConvErr
		}
		result[packdirect.Name] = entry
	}
	return result, nil
}
