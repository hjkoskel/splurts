package messagepack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricArr(t *testing.T) {
	vec, errVec := CreateDeltaRLEVec([]int64{1, 2, 3, 4, 5, 6, 100, 102, 105, 110, 120, 130, 140, 140, 140, 140}, 2, 2)
	assert.Equal(t, nil, errVec)

	dut := MetricArr{ //One entry from splurts struct
		Meta:   MetricMeta{},
		Coding: MetricCoding{},
		//Enums: []string{},
		Steps: []MetricStep{
			{Count: 100, Step: 0.25},
		},
		Delta: 1,
		Data:  vec,
	}

	b := new(bytes.Buffer)
	err := dut.Write(b)
	assert.Equal(t, nil, err)
	code := b.Bytes()

	ref, refErr := ReadMetricArr(bytes.NewBuffer(code))
	assert.Equal(t, nil, refErr)
	assert.Equal(t, dut, ref)

}
