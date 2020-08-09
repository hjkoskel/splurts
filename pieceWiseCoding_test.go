package splurts

import (
	"math"
	"testing"
)

func TestValid(t *testing.T) {
	dut := PiecewiseCoding{
		Name:    "", //Needed only for floatStruct
		Min:     12.3,
		Steps:   []PiecewiseCodingStep{},
		Clamped: false,
	}

	if dut.IsInvalid() == nil {
		t.Errorf("invalid test fail")
	}

	dut.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Count: 20}}
	if dut.IsInvalid() == nil {
		t.Errorf("zero size test")
	}

	dut.Steps = []PiecewiseCodingStep{PiecewiseCodingStep{Count: 20, Size: 0.15}}

	if dut.IsInvalid() != nil {
		t.Error("Is invalid even it should not")
	}

}

/*func TestEndian(t *testing.T) {
	dut := PiecewiseCoding{
		Min: 0,
		Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Count: 65536, Size: 1},
		},
		Clamped: false,
	}

}*/

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

	expectedvalue := 10 + 20*0.15 + 10*0.1 + 10*0.2
	if dut.Max() != expectedvalue {
		t.Errorf("Invalid Max")
	}

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
	if dut.NumberOfBits() != 4 {
		t.Errorf("Wrong number of bits on clamped=false %v", dut.NumberOfBits())
	}
	if dut.MaxCode() != 15 { //It is calculated from number of bits
		t.Errorf("Invalid max code %v", dut.MaxCode())
	}

	dut.Clamped = true
	if dut.NumberOfBits() != 3 {
		t.Errorf("Wrong number of bits on clamped=true %v", dut.NumberOfBits())
	}

	dut.Clamped = true
	if dut.MaxCode() != 7 {
		t.Errorf("Invalid max code %v", dut.MaxCode())
	}

	booldut := PiecewiseCoding{
		Name: "", //Needed only for floatStruct
		Min:  0,
		Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Count: 2, Size: 1}, //Boolean have two steps... constant one step
		},
		Clamped: true,
	}
	if booldut.NumberOfBits() != 1 {
		t.Errorf("Boolean do have more than 1 bit")
	}
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

	if dut.BitCode(11.2) != "1110" {
		t.Errorf("End coding error")
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

	if dut.BitCode(5.5) != "0000" {
		t.Errorf("End coding error")
	}

	//1 bit test
	dut1 := PiecewiseCoding{ //Define 4 state enum
		Name:    "mode",
		Clamped: true,
		Min:     0, Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Size: 1, Count: 2},
		},
	}
	t.Logf("1bit is %v  bits=%v totalSTeps=%v\n", dut1.NumberOfBits(), dut1.BitCode(1), dut1.TotalStepCount())

	if dut1.NumberOfBits() != 1 {
		t.Errorf("one bit is not one bit")
	}

	if len(dut1.BitCode(1)) != 1 {
		t.Errorf("one bit bit is not one in length")
	}

	//2 bit test
	dut2 := PiecewiseCoding{ //Define 4 state enum
		Name:    "mode",
		Clamped: true,
		Min:     0, Steps: []PiecewiseCodingStep{
			PiecewiseCodingStep{Size: 1, Count: 4},
		},
	}

	//Clamped mode. Means that values are already clamped
	if len(dut2.BitCode(2)) != 2 {
		t.Errorf("one bit bit is not one in length")
	}

}
