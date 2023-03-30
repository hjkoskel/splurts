package messagepack

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	METRICNAME_EPOCH = "epoch" //Use this metric as time axis by default
)

// These metadatas are very useful when data is finally plotted
const (
	MPNAME_META             = "meta"
	MPNAME_META_UNIT        = "unit"        //Unit like kg. Used when plotting and grouping "compatible" metrics together
	MPNAME_META_CAPTION     = "caption"     //Caption for this metric, optional. Printable text without unit
	MPNAME_META_ACCURACY    = "accuracy"    //Numerical value or string pointing to other field by name Plus minus value
	MPNAME_META_MAXINTERVAL = "maxinterval" //During data loggin time NaN are indication that data is missing. This limit tells that it is declared faulty measurement instead of slower rate
	MPNAME_META_BANDWIDTH   = "bandwidth"
)
const (
	MPNAME_CODING         = "coding"
	MPNAME_CODING_MIN     = "min"
	MPNAME_CODING_MAX     = "max"
	MPNAME_CODING_CLAMPED = "cla"
)

const (
	MPNAME_ENUMS = "enums"
	MPNAME_STEPS = "steps"

	MPNAME_POINTCOUNT = "n"
	MPNAME_STEPSIZE   = "s"

	MPCODING_SIMPLEARR = "arr"
	MPCODING_DELTA1ARR = "delta1"
	MPCODING_DELTA2ARR = "delta2"
)

func getMessagepackTagString(input interface{}) (map[string]string, error) {
	result := make(map[string]string)
	rt := reflect.TypeOf(input)

	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		vo := reflect.ValueOf(input)
		if vo.Len() == 0 {
			return nil, fmt.Errorf("no slice items")
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
