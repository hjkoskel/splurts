package messagepack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	/*fmt.Printf("--DELTA--\n  Input: %#v\nOutput:\n", input)
	for _, v := range code {
		fmt.Printf("%02X ", v)
	}
	fmt.Printf("\n")
	*/
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
