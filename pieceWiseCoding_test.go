package splurts

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {
	dut := PiecewiseCoding{
		Name:    "a", //Needed only for floatStruct
		Min:     12.3,
		Steps:   []PiecewiseCodingStep{},
		Clamped: false,
	}

	if dut.IsInvalid() == nil {
		t.Errorf("invalid test fail")
	}

	dut.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Count: 20}}
	assert.NotEqual(t, nil, dut.IsInvalid())
	dut.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Count: 20, Size: 0.15}}
	assert.Equal(t, nil, dut.IsInvalid())
}

func TestEnums(t *testing.T) {
	dut := PiecewiseCoding{
		Name:  "b",
		Enums: []string{"ONE", "two", "three", "FOUR"},
	}

	assert.Equal(t, 3, dut.NumberOfBits())
	assert.Equal(t, uint64(5), dut.TotalStepCount())

	code, codeErr := dut.BitCode(2)
	assert.Equal(t, nil, codeErr)
	assert.Equal(t, "010", code)
}

func TestDecimals(t *testing.T) {
	dut := PiecewiseCoding{Steps: []PiecewiseCodingStep{PiecewiseCodingStep{Size: 0.0123}}}
	assert.Equal(t, 2, dut.Decimals())
	dut = PiecewiseCoding{Steps: []PiecewiseCodingStep{PiecewiseCodingStep{Size: 1.5}}}
	assert.Equal(t, 0, dut.Decimals())
}

func TestMax(t *testing.T) {
	//minimum is direct value
	dut := PiecewiseCoding{
		Name: "", //Needed only for floatStruct
		Min:  10,
		Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Count: 20, Size: 0.15},
			PiecewiseCodingStep{Count: 10, Size: 0.10},
			PiecewiseCodingStep{Count: 10, Size: 0.20},
		},
		Clamped: false,
	}
	assert.Equal(t, 10+20*0.15+10*0.1+10*0.2, dut.Max())
}

func TestNumberOfBits(t *testing.T) {
	//minimum is direct value
	dut := PiecewiseCoding{
		Name: "", //Needed only for floatStruct
		Min:  10,
		Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Count: 2, Size: 0.15},
			PiecewiseCodingStep{Count: 2, Size: 0.10},
			PiecewiseCodingStep{Count: 3, Size: 0.20},
		},
		Clamped: false,
	}
	assert.Equal(t, 4, dut.NumberOfBits())
	assert.Equal(t, uint64(15), dut.MaxCode())

	dut.Clamped = true
	assert.Equal(t, 3, dut.NumberOfBits())

	dut.Clamped = true
	assert.Equal(t, uint64(7), dut.MaxCode())

	booldut := PiecewiseCoding{
		Name: "", //Needed only for floatStruct
		Min:  0,
		Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Count: 2, Size: 1}, //Boolean have two steps... constant one step
		},
		Clamped: true,
	}
	assert.Equal(t, 1, booldut.NumberOfBits())
}

func floatvectest(t *testing.T, got []float64, wanted []float64) {
	if len(got) != len(wanted) {
		t.Errorf("Expected decoding \n%#v\nnot\n%#v", wanted, got)
		t.FailNow()
	}
	for i, f := range wanted {
		if math.IsNaN(f) {
			if !math.IsNaN(got[i]) {
				if got[i] != f {
					t.Errorf("ERR On index %v f=%v", i, f)
					t.Errorf("Expected decoding \n%#v\nnot\n%#v", wanted, got)
					t.FailNow()
				}

			}
		} else {
			if 0.000001 < math.Abs(got[i]-f) {
				t.Errorf("ERR On index %v f=%v", i, f)
				t.Errorf("Expected decoding \n%#v\nnot\n%#v", wanted, got)
				t.FailNow()
			}
		}
	}
}

func TestDecode(t *testing.T) {
	//minimum is direct value
	dut := PiecewiseCoding{
		Name: "", //Needed only for floatStruct
		Min:  10,
		Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Count: 2, Size: 0.15},
			PiecewiseCodingStep{Count: 2, Size: 0.10},
			PiecewiseCodingStep{Count: 3, Size: 0.20},
		},
		Clamped: false, //False= non clampled so is that -Inf +Inf are included.
	}

	wanted := []float64{
		math.Inf(-1),
		10, //min value
		10.15, 10.3,
		10.4, 10.5,
		10.7, 10.9, 11.1,
		math.Inf(1), math.Inf(1), math.Inf(1), math.Inf(1), math.Inf(1), math.Inf(1), math.NaN()}
	got := []float64{}

	for i := range wanted {
		got = append(got, dut.ScaleToFloat(uint64(i)))
	}
	floatvectest(t, got, wanted)

	dut.Clamped = true //Values are clamped when entering so no special inf needed

	wanted = []float64{
		10, //min value
		10.15, 10.3,
		10.4, 10.5,
		10.7, 10.9, 11.1,
		11.2999999, 11.5, 11.7, 11.9, 12.1, 12.3, 12.5}
	got = []float64{}

	for i := range wanted {
		got = append(got, dut.ScaleToFloat(uint64(i)))
	}
	floatvectest(t, got, wanted)

}

func TestCode(t *testing.T) {
	//minimum is direct value
	dut := PiecewiseCoding{
		Name: "", //Needed only for floatStruct
		Min:  10,
		Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Count: 2, Size: 0.15},
			PiecewiseCodingStep{Count: 2, Size: 0.10},
			PiecewiseCodingStep{Count: 3, Size: 0.20},
		},
		Clamped: false, //False= non clampled so is that -Inf +Inf are included.
	}

	code, codeErr := dut.BitCode(11.2)
	if code != "1110" || codeErr != nil {
		t.Errorf("End coding error %v", codeErr)
	}
	if dut.ScaleToFloat(dut.ScaleToUint(11.2)) != math.Inf(1) {
		t.Errorf("Inf err")
	}

	testpoints := []float64{10, 10.15, 10.4, 10.7, 11.1}

	for _, f := range testpoints {
		if dut.ScaleToFloat(dut.ScaleToUint(f)) != f {
			t.Errorf("code decode err")
		}
	}

	if !math.IsNaN(dut.ScaleToFloat(dut.ScaleToUint(math.NaN()))) {
		t.Errorf("Nan err")
	}

	code, codeErr = dut.BitCode(5.5)
	if code != "0000" || codeErr != nil {
		t.Errorf("End coding error %v", codeErr)
	}

	//1 bit test
	dut1 := PiecewiseCoding{ //Define 4 state enum
		Name:    "mode",
		Clamped: true,
		Min:     0, Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Size: 1, Count: 2},
		},
	}

	code, codeErr = dut1.BitCode(1)
	if codeErr != nil {
		t.Errorf("End coding error %v", codeErr)
	}

	t.Logf("1bit is %v  bits=%v totalSTeps=%v\n", dut1.NumberOfBits(), code, dut1.TotalStepCount())

	if dut1.NumberOfBits() != 1 {
		t.Errorf("one bit is not one bit")
	}

	code, _ = dut1.BitCode(1)
	assert.Equal(t, 1, len(code))

	//2 bit test
	dut2 := PiecewiseCoding{ //Define 4 state enum
		Name:    "mode",
		Clamped: true,
		Min:     0, Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Size: 1, Count: 4},
		},
	}

	//Clamped mode. Means that values are already clamped
	code, _ = dut2.BitCode(2)
	assert.Equal(t, 2, len(code))
}
