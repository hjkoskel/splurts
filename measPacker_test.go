package splurts

import (
	"math"
	"testing"
)

func TestExample(t *testing.T) {
	structure := PiecewiseFloats([]PiecewiseCoding{
		PiecewiseCoding{
			Name:    "temperature",
			Clamped: false, //Must be able to report missig values if sensor fails
			Min:     -40, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 1, Count: 20},    // -40 to -20 one degree granularity is enough
				PiecewiseCodingStep{Size: 0.1, Count: 400}, //-20 to 20 more decimals are needed 0.1
				PiecewiseCodingStep{Size: 1, Count: 20},    //20 to 35 less decimals are needed 0.1
			},
		},
		PiecewiseCoding{
			Name:    "RH",
			Clamped: false, //Must be able to report missig values if sensor fails
			Min:     0, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 0.1, Count: 1000},
			},
		},

		PiecewiseCoding{
			Name:    "pressure",
			Clamped: false, //Must be able to report missig values if sensor fails
			Min:     80000, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 100, Count: 300}, //110kPa
			},
		},

		PiecewiseCoding{ //Define boolean 0 or 1  (in float)
			Name:    "heater", //Flag is heater on
			Clamped: true,     //Do not spend extra bits
			Min:     0, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 1, Count: 2},
			},
		},
		PiecewiseCoding{ //Define 4 state enum
			Name:    "mode",
			Clamped: true,
			Min:     0, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 1, Count: 4},
			},
		},
	})

	if structure.IsInvalid() != nil {
		t.Errorf("is invalid test fail. Should be not invalid")
	}

	t.Logf("Demo struct is %v bits\n", structure.NumberOfBits())
	for i, pie := range structure {
		t.Logf("  %v: %v %v bits\n", i, pie.Name, pie.NumberOfBits())
	}

	data, err := structure.Encode(map[string]float64{"temperature": 21.3, "RH": 35.3, "pressure": 102401, "heater": 0, "mode": 2})
	if err != nil {
		t.Errorf("Encoding error %v\n", err.Error())
	}

	if len(data) != 4 {
		t.Errorf("data length error %v", len(data))
	}

	parsed, parseErr := structure.Decode(data, true)
	if parseErr != nil {
		t.Errorf("Decode error %v", parseErr)
	}

	temp, hazTemp := parsed["temperature"]
	rh, hazRh := parsed["RH"]
	pressure, hazPressure := parsed["pressure"]
	heater, hazHeater := parsed["heater"]
	mode, hazMode := parsed["mode"]

	if !hazTemp || !hazRh || !hazPressure || !hazHeater || !hazMode {
		t.Errorf("Missing variables")
	}

	if 0.0001 < math.Abs(temp-21) || // Notice rounding
		0.0001 < math.Abs(rh-35.3) ||
		0.0001 < math.Abs(pressure-102400) || //Notice rounding
		0.0001 < math.Abs(heater-0) ||
		0.0001 < math.Abs(mode-2) {
		t.Errorf("value error")
	}

	//Missign data demo
	missingData, err := structure.Encode(map[string]float64{"temperature": 23.3, "RH": 35.3, "heater": 0, "mode": 1})
	if err != nil {
		t.Errorf("Encoding error %v\n", err.Error())
	}

	parsed, parseErr = structure.Decode(missingData, false) //Leave NaN values away
	if parseErr != nil {
		t.Errorf("Encoding error %v\n", parseErr.Error())
	}
	_, hazPressure = parsed["pressure"]
	if hazPressure {
		t.Errorf("Should not have pressure")
	}

	parsed, parseErr = structure.Decode(missingData, true) //NaN values away
	if parseErr != nil {
		t.Errorf("Encoding error %v\n", parseErr.Error())
	}

	nanPressure := float64(0)
	nanPressure, hazPressure = parsed["pressure"]
	if !hazPressure {
		t.Errorf("Should have pressure")
	}
	if !math.IsNaN(nanPressure) {
		t.Errorf("Pressure is not NAN")
	}
	//Invalidty check
	parsed, parseErr = structure.Decode([]byte{4, 2}, true)
	if parseErr == nil {
		t.Errorf("raw data length not checked")
	}
	structure[1].Name = ""
	if structure.IsInvalid() == nil {
		t.Errorf("missing name not checked")
	}
	structure[1].Name = "sadfsadf"
	structure[1].Steps[0].Size = -1
	if structure.IsInvalid() == nil {
		t.Errorf("missing name not checked")
	}
}

func TestHex(t *testing.T) {
	structure := PiecewiseFloats([]PiecewiseCoding{
		PiecewiseCoding{
			Name:    "temperature",
			Clamped: false, //Must be able to report missig values if sensor fails
			Min:     -40, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 1, Count: 20},    // -40 to -20 one degree granularity is enough
				PiecewiseCodingStep{Size: 0.1, Count: 400}, //-20 to 20 more decimals are needed 0.1
				PiecewiseCodingStep{Size: 1, Count: 20},    //20 to 35 less decimals are needed 0.1
			},
		},
		PiecewiseCoding{
			Name:    "RH",
			Clamped: false, //Must be able to report missig values if sensor fails
			Min:     0, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 0.1, Count: 1000},
			},
		},

		PiecewiseCoding{
			Name:    "pressure",
			Clamped: false, //Must be able to report missig values if sensor fails
			Min:     80000, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 100, Count: 300}, //110kPa
			},
		},

		PiecewiseCoding{ //Define boolean 0 or 1  (in float)
			Name:    "heater", //Flag is heater on
			Clamped: true,     //Do not spend extra bits
			Min:     0, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 1, Count: 2},
			},
		},
		PiecewiseCoding{ //Define 4 state enum
			Name:    "mode",
			Clamped: true,
			Min:     0, Steps: []PiecewiseCodingStep{
				PiecewiseCodingStep{Size: 1, Count: 4},
			},
		},
	})

	if structure.IsInvalid() != nil {
		t.Errorf("is invalid test fail. Should be not invalid")
	}

	t.Logf("Demo struct is %v bits\n", structure.NumberOfBits())
	for i, pie := range structure {
		t.Logf("  %v: %v %v bits\n", i, pie.Name, pie.NumberOfBits())
	}

	data, err := structure.EncodeToHex(map[string]float64{"temperature": 21.3, "RH": 35.3, "pressure": 102401, "heater": 0, "mode": 2})
	if err != nil {
		t.Errorf("Encoding error %v\n", err.Error())
	}

	parsed, parseErr := structure.DecodeHex(data, true)
	if parseErr != nil {
		t.Errorf("Decode error %v", parseErr)
	}

	temp, hazTemp := parsed["temperature"]
	rh, hazRh := parsed["RH"]
	pressure, hazPressure := parsed["pressure"]
	heater, hazHeater := parsed["heater"]
	mode, hazMode := parsed["mode"]

	if !hazTemp || !hazRh || !hazPressure || !hazHeater || !hazMode {
		t.Errorf("Missing variables")
	}

	if 0.0001 < math.Abs(temp-21) || // Notice rounding
		0.0001 < math.Abs(rh-35.3) ||
		0.0001 < math.Abs(pressure-102400) || //Notice rounding
		0.0001 < math.Abs(heater-0) ||
		0.0001 < math.Abs(mode-2) {
		t.Errorf("value error")
	}

}
