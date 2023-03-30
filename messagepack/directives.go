package messagepack

import "strconv"

const (
	MPDIRECTIVE       = "messagepack"
	MPDIRECTIVE_NAME  = "name"
	MPDIRECTIVE_DELTA = "delta" //0=no delta coding, 1 once, 2 twice
	MPDIRECTIVE_RLE   = "rle"   //0=not in use 1=more than 1 then use array, at least 3 prefered
)

type Directives struct {
	Name  string
	Delta int64
	Rle   int64
}

func parseDirectives(s string, nameDefault string) (Directives, error) {
	result := Directives{Name: nameDefault, Delta: 0, Rle: 3}
	tagmap := splitTags(s)
	overrideName, hazName := tagmap[MPDIRECTIVE_NAME]
	deltaString, hazDelta := tagmap[MPDIRECTIVE_DELTA]
	rleString, hazRLE := tagmap[MPDIRECTIVE_RLE]

	if hazName && 0 < len(overrideName) {
		result.Name = overrideName
	}
	var errParse error
	if hazDelta {
		result.Delta, errParse = strconv.ParseInt(deltaString, 10, 64)
		if errParse != nil {
			return result, errParse
		}
	}
	if hazRLE {
		result.Rle, errParse = strconv.ParseInt(rleString, 10, 64)
		if errParse != nil {
			return result, errParse
		}
	}
	return result, nil
}
