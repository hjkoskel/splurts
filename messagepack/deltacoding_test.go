package messagepack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDeltaRLEVec(t *testing.T) {
	inputdata := []int64{6, 2, 2, 4, 1, 4, 4, 4}

	dat, err := CreateDeltaRLEVec([]int64{6, 2, 2, 4, 1, 4, 4, 4}, 0, 0)

	assert.Equal(t, err, nil)
	ref, errRef := dat.ToArr(0)
	assert.Equal(t, errRef, nil)
	assert.Equal(t, inputdata, ref)

	b := new(bytes.Buffer)
	wErr := dat.WriteToBuf(b)
	assert.Equal(t, wErr, nil)
	readback, errReadBack := ReadDeltaRLEVec(b)
	assert.Equal(t, errReadBack, nil)
	assert.Equal(t, dat, readback)

	dat, err = CreateDeltaRLEVec([]int64{6, 2, 2, 4, 1, 4, 4, 4}, 0, 2)
	assert.Equal(t, err, nil)
	ref, errRef = dat.ToArr(0)
	assert.Equal(t, errRef, nil)
	assert.Equal(t, inputdata, ref)

	dat, err = CreateDeltaRLEVec([]int64{6, 2, 2, 4, 1, 4, 4, 4}, 1, 2)
	assert.Equal(t, err, nil)
	ref, errRef = dat.ToArr(1)
	assert.Equal(t, errRef, nil)
	assert.Equal(t, inputdata, ref)

	dat, err = CreateDeltaRLEVec([]int64{6, 2, 2, 4, 1, 4, 4, 4}, 2, 2)
	assert.Equal(t, err, nil)
	ref, errRef = dat.ToArr(2)
	assert.Equal(t, errRef, nil)
	assert.Equal(t, inputdata, ref)

	b = new(bytes.Buffer)
	wErr = dat.WriteToBuf(b)
	assert.Equal(t, wErr, nil)

	readback, errReadBack = ReadDeltaRLEVec(b)
	assert.Equal(t, errReadBack, nil)
	assert.Equal(t, dat, readback)
}

func TestDeltaUndelta(t *testing.T) {
	inputdata := []int64{6, 2, 2, 4, 1, 4, 4, 4}
	delted := DeltaVec(inputdata)
	outputdata := UnDeltaVec(delted)

	assert.Equal(t, inputdata, outputdata)

	assert.Equal(t, inputdata, UnDeltaVec(UnDeltaVec(DeltaVec(DeltaVec(inputdata)))))

}

func TestOneDelta(t *testing.T) {
	b := new(bytes.Buffer)
	inputdata := []int64{6, 2, 2, 4, 1, 4, 4, 4}

	err := ArrToRLEMessagepack(b, inputdata, 3)
	assert.Equal(t, nil, err)

	code := b.Bytes()
	codebuf := bytes.NewBuffer(code)

	unpacked, unpackerr := RLEMessagepackToArr(codebuf)
	assert.Equal(t, nil, unpackerr)

	assert.Equal(t, inputdata, unpacked)
}

func TestRLE(t *testing.T) {
	b := new(bytes.Buffer)
	inputdata := []int64{6, 2, 2, 4, 4, 4, 4, 4, 4, 4, 1, 4, 4, 4}
	err := ArrToRLEMessagepack(b, inputdata, 3)
	assert.Equal(t, nil, err)

	code := b.Bytes()
	codebuf := bytes.NewBuffer(code)

	unpacked, unpackerr := RLEMessagepackToArr(codebuf)
	assert.Equal(t, nil, unpackerr)

	assert.Equal(t, inputdata, unpacked)

}
