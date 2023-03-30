package messagepack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricCoding(t *testing.T) {
	dut := MetricCoding{Min: 3.1, Max: 102.1, Clamped: true}
	b := new(bytes.Buffer)
	err := dut.Write(b)
	assert.Equal(t, nil, err)
	code := b.Bytes()

	ref, refErr := ReadMetricCoding(bytes.NewBuffer(code))
	assert.Equal(t, nil, refErr)
	assert.Equal(t, dut, ref)
}

func TestMetricStep(t *testing.T) {
	dut := MetricStep{Count: 69, Step: 4.2}
	b := new(bytes.Buffer)
	err := dut.Write(b)
	assert.Equal(t, nil, err)
	code := b.Bytes()

	ref, refErr := ReadMetricStep(bytes.NewBuffer(code))
	assert.Equal(t, nil, refErr)
	assert.Equal(t, dut, ref)
}

func TestMetricMeta(t *testing.T) {
	dut := MetricMeta{
		Unit:        "kg",
		Caption:     "punnittelu",
		Accuracy:    "weightErr",
		MaxInterval: 1000 * 1000 * 1000,
		Bandwidth:   123.2,
	}
	b := new(bytes.Buffer)
	err := dut.Write(b)
	assert.Equal(t, nil, err)
	code := b.Bytes()
	ref, refErr := ReadMetricMeta(bytes.NewBuffer(code))
	assert.Equal(t, nil, refErr)
	assert.Equal(t, dut, ref)

}
