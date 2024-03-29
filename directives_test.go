/*
Testing with directives.

Use this as example
*/

package splurts

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TimeExampleStruct struct {
	ExampleMetric float64   `splurts:"step=0.1,min=-40,max=40"`
	CompleteTime  time.Time //nanosec spend 42bits
	SecondTime    time.Time `splurts:"min=1670000000000,step=1000"` //uses 31 bits, ends at "Tue Apr 06 2106"
}

func TestTime(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(TimeExampleStruct{})
	if errRecipe != nil {
		t.Errorf("recipe error %v", errRecipe)
	}

	testTime := time.UnixMilli(1670523401643)
	d := TimeExampleStruct{CompleteTime: testTime, SecondTime: testTime}
	byt, errsplurt := recipe.Splurts(d)
	assert.Equal(t, nil, errsplurt)

	ref := TimeExampleStruct{}

	errUnsplurt := recipe.UnSplurts(byt, &ref)
	assert.Equal(t, nil, errUnsplurt)
	assert.Equal(t, d.ExampleMetric, ref.ExampleMetric)
	assert.Equal(t, d.CompleteTime, ref.CompleteTime)
	assert.Equal(t, int64(1670523402000), ref.SecondTime.UnixMilli())
}

type ParticleMeas struct {
	SystemStatus string  `splurts:"enum=UNDEFINED,INITIALIZE,IDLE,MEASURE,STOP,ERROR"`
	Temperature  float64 `splurts:"step=0.1,min=-40,max=40"`
	StaticSymbol int     `splurts:"bits=7,const=42"`
	Humidity     float64 `splurts:"step=0.05,min=0,max=100"`
	Pressure     float64 `splurts:"step=100,min=85000,max=110000"`
	Small        float64 `splurts:"step=0.1,min=0,max=300"`
	Large        float64 `splurts:"step=0.1,min=0,max=300,infpos=99999,infneg=-99999"`
	Extra        float64 `splurts:"step=0.1,min=0,bits=14"`
	Heater       bool    `splurts:"enum=NOTHEATING,HEATING"` //Flag for heater enabled
	Emptyvalue   string  `splurts:"enum=NO,MAYBE,YES"`
}

func TestArrayOfStructsValue(t *testing.T) {
	testArr := []ParticleMeas{
		{SystemStatus: "IDLE", Temperature: 22.1, Humidity: 32.4, Pressure: 95200, Small: 69.42, Large: 33, Extra: 2, Heater: false, Emptyvalue: "YES"},
		{SystemStatus: "MEASURE", Temperature: 22.2, Humidity: 32.6, Pressure: 95100, Small: 67.11, Large: 34, Extra: 3, Heater: true, Emptyvalue: "YES"},
		{SystemStatus: "STOP", Temperature: 22.3, Humidity: 32.6, Pressure: 95100, Small: 67.11, Large: 34, Extra: 3, Heater: false, Emptyvalue: "YES"},
	}

	recipe, errRecipe := GetPiecewisesFromStruct(ParticleMeas{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

	m, mErr := recipe.GetValuesToFloatMapArr(testArr)
	if mErr != nil {
		t.Errorf("map err %v", mErr.Error())
	}

	var wanted = map[string][]float64{
		"Emptyvalue":   {3, 3, 3},
		"Extra":        {2, 3, 3},
		"Heater":       {0, 1, 0},
		"Humidity":     {32.4, 32.6, 32.6},
		"Large":        {33, 34, 34},
		"Pressure":     {95200, 95100, 95100},
		"Small":        {69.42, 67.11, 67.11},
		"StaticSymbol": {42, 42, 42},
		"SystemStatus": {3, 4, 5},
		"Temperature":  {22.1, 22.2, 22.3}}

	assert.Equal(t, wanted, m)

}

func TestNotMatchingConst(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(ParticleMeas{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

	d := ParticleMeas{}
	byt, _ := recipe.Splurts(d)
	byt = make([]byte, len(byt)) //zero out
	e := recipe.UnSplurts(byt, d)
	assert.Equal(t, fmt.Errorf("const field StaticSymbol is -Inf not 42"), e)
}

func TestStringConversion(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(ParticleMeas{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

	d := ParticleMeas{
		SystemStatus: "MEASURE",
		Temperature:  -40.149,
		Humidity:     110.1291,
		Pressure:     80000.51,
		Small:        301.222,
		Large:        -0.001, //TODO NO NEGATIVE string values
		Extra:        2.1,
	}

	testString, _ := recipe.ToStrings(d, true)
	v, haz := testString["SystemStatus"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "\"MEASURE\"", v)
	v, haz = testString["Temperature"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "-40.1", v)
	v, haz = testString["Humidity"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "110.13", v)
	v, haz = testString["Pressure"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "80001", v)
	v, haz = testString["Small"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "301.2", v)
	v, haz = testString["Large"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "0.0", v)
	v, haz = testString["Extra"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "2.1", v)
	v, haz = testString["Heater"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "0", v)
	v, haz = testString["Emptyvalue"]
	assert.Equal(t, true, haz)
	assert.Equal(t, "\"\"", v)
}

func TestInfCase(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(ParticleMeas{})
	assert.Equal(t, nil, errRecipe)

	d := ParticleMeas{
		SystemStatus: "MEASURE",
		Temperature:  -40.1225,
		Humidity:     110,
		Pressure:     80000,
		Small:        301,
		Large:        -0.001,
		Extra:        2.1,
	}

	byt, errSplurt := recipe.Splurts(d)
	assert.Equal(t, nil, errSplurt)
	newD := ParticleMeas{}
	t.Logf("Splurted inf case %#v  (%v bytes)\n", byt, len(byt))
	e := recipe.UnSplurts(byt, &newD)
	assert.Equal(t, nil, e)
	t.Logf("New inf=%#v\n\n", newD)

	assert.Equal(t, newD.StaticSymbol, 42)

	if !math.IsInf(newD.Temperature, -1) {
		t.Errorf("Inf error")
	}
	if !math.IsInf(newD.Humidity, 1) {
		t.Errorf("Inf error")
	}
	if !math.IsInf(newD.Pressure, -1) {
		t.Errorf("Inf error")
	}

	if !math.IsInf(newD.Small, 1) {
		t.Errorf("Inf error")
	}

	/*if !math.IsInf(newD.Large, -1) {
		t.Errorf("Inf error")
	}*/
	if newD.Large != -99999 {
		t.Errorf("neg inf not working")
	}

}

func TestEasyUseCase(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(ParticleMeas{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

	d := ParticleMeas{SystemStatus: "IDLE", Temperature: 18.3, Humidity: 23.6, Pressure: 102400, Small: 23.2, Large: 41}
	byt, errSplurt := recipe.Splurts(d)
	assert.Equal(t, nil, errSplurt)

	newD := ParticleMeas{}
	t.Logf("Splurted easy case %#v  (%v bytes)\n", byt, len(byt))
	e := recipe.UnSplurts(byt, &newD)

	assert.Equal(t, nil, e)

	t.Logf("NewD=%#v\n\n", newD)

	if newD.SystemStatus != d.SystemStatus {
		t.Errorf("invalid system status %s vs %s", newD.SystemStatus, d.SystemStatus)
	}

	if 0.00001 < math.Abs(float64(newD.Temperature-d.Temperature)) {
		t.Errorf("Invalid Temperature")
	}

	if 0.00001 < math.Abs(float64(newD.Humidity-d.Humidity)) {
		t.Errorf("Invalid")
	}

	if 0.00001 < math.Abs(float64(newD.Pressure-d.Pressure)) {
		t.Errorf("Invalid")
	}

	if 0.00001 < math.Abs(float64(newD.Small-d.Small)) {
		t.Errorf("Invalid")
	}

	if 0.00001 < math.Abs(float64(newD.Large-d.Large)) {
		t.Errorf("Invalid")
	}

}

//--------------------------
type Fail0 struct {
	V float64
}

type Fail1 struct {
	V int
}

type Fail2 struct {
	V float64 `splurts:"step=0.1 min=-40,max=40"`
}

type Fail3 struct {
	V string
}

type Fail4 struct {
	V float64 `splurts:"step=0.1 min=-40,max=40"`
}

type Fail5 struct {
	V float64 `splurts:"step=0.1,min=-40,max=40,asdf"`
}

type Fail6 struct {
	V float64 `splurts:"step=0.1,min=-40,max=q"`
}

type Fail7 struct {
	V float64 `splurts:"min=-100,steps=5.0 10|100|1.5 10"`
}

type Fail8 struct {
	V float64 `splurts:"min=-100,steps=5.0 10|a b|1.5 10"`
}

type Fail9 struct {
	V float64 `splurts:"step=0.1,min=-40,max=999999999999999999999999999999999"`
}

func TestFails(t *testing.T) {
	_, e := GetPiecewisesFromStruct(Fail0{})
	if e == nil {
		t.Error("Fail 0 not failed")
	}

	_, e = GetPiecewisesFromStruct(Fail1{})
	if e == nil {
		t.Error("Fail 1 not failed")
	}

	_, e = GetPiecewisesFromStruct(Fail2{})
	if e == nil {
		t.Error("Fail 2 not failed")
	}

	_, e = GetPiecewisesFromStruct(Fail3{})
	if e == nil {
		t.Error("Fail 3 not failed")
	}

	_, e = GetPiecewisesFromStruct(Fail4{})
	if e == nil {
		t.Error("Fail 4 not failed")
	}

	_, e = GetPiecewisesFromStruct(Fail5{})
	if e == nil {
		t.Error("Fail 5 not failed")
	}
	_, e = GetPiecewisesFromStruct(Fail6{})
	if e == nil {
		t.Error("Fail 6 not failed")
	}

	_, e = GetPiecewisesFromStruct(Fail7{})
	if e == nil {
		t.Error("Fail 7 not failed")
	}

	_, e = GetPiecewisesFromStruct(Fail8{})
	if e == nil {
		t.Error("Fail 8 not failed")
	}

}

//More complete case
type AllCases struct {
	Alpha   float64 `splurts:"step=0.1,min=-40,max=40"`
	Bravo   float32 `splurts:"step=0.1,min=-40,max=40"`
	Charlie int32   `splurts:"min=7,max=100"` //integers have default stepsize 1
	Delta   uint32  `splurts:"max=150"`
	Echo    bool
	Foxtrot float64 `splurts:"min=-100,steps=5.0 10|0.5 100|1.5 10"`
	Golf    int     `splurts:"min=-10,steps=2.0 5|1.0 100"`
	Hotel   float32 `splurts:"step=0.1,min=-40,max=40,clamped"`

	LeaveThisOut     float64 `splurts:"omit"`
	LeaveThisOutAlso float64 `splurts:"omit"`

	India    float64 `splurts:"step=0.1,min=-40,max=40,clamped"`
	Juliet   float32 `splurts:"step=0.1,min=-40,max=40,clamped"`
	Kilo     int32   `splurts:"min=7,max=100,clamped"` //integers have default stepsize 1
	Lima     uint32  `splurts:"max=150,clamped"`
	Mike     float64 `splurts:"min=-100,steps=5.0 10|0.5 100|1.5 10,clamped"`
	November int     `splurts:"min=-100,steps=5.0 10|0.5 100|1.5 10,clamped"`
	Oscar    float64 `splurts:"bits=12,min=-45.5,max=40"`
	Papa     float64 `splurts:"bits=12,min=-45.5,max=40,clamped"`
	Quebeck  float64 `splurts:"bits=12,min=-45.5,step=0.3,bits=12,clamped"`
}

func Test7bitCompleteCase(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(AllCases{})
	assert.Equal(t, nil, errRecipe)

	t.Logf("\nCOMPLETE RECIPE IS \n%v\n\n", recipe)

	//In range
	d := AllCases{
		Alpha:   2.3,
		Bravo:   10.1,
		Charlie: 42,
		Delta:   123,
		Echo:    true,
		Foxtrot: 13.5,
		Golf:    5,
		Hotel:   26.4,

		India:    20.4,
		Juliet:   1.1,
		Kilo:     25,
		Lima:     42,
		Mike:     -21.5,
		November: -90,
		Oscar:    42.69,
		Papa:     112.5,
	}

	byt, errSplurt := recipe.Splurts7bitBytes(d)
	if errSplurt != nil {
		t.Errorf(errSplurt.Error())
	} else {
		t.Logf("SPLURTS %#v (len=%v)\n", byt, len(byt))
		newD := AllCases{}
		e := recipe.UnSplurts7bitBytes(byt, &newD)
		if e != nil {
			t.Errorf("Unsplurt err %v", e.Error())
		} else {
			t.Logf("NewD=%#v\n\n", newD)
			if 0.00001 < math.Abs(newD.Alpha-d.Alpha) {
				t.Errorf("Invalid Alpha")
			}
			if 0.00001 < math.Abs(float64(newD.Bravo-d.Bravo)) {
				t.Errorf("Invalid Bravo")
			}
			if 0.00001 < math.Abs(float64(newD.Charlie-d.Charlie)) {
				t.Errorf("Invalid Charlie")
			}
			if 0.00001 < math.Abs(float64(newD.Delta-d.Delta)) {
				t.Errorf("Invalid Delta")
			}
			if newD.Echo != d.Echo {
				t.Errorf("Invalid Echo")
			}
			if 0.00001 < math.Abs(float64(newD.Foxtrot-d.Foxtrot)) {
				t.Errorf("Invalid Foxtrot %v vs %v", newD.Foxtrot, d.Foxtrot)
			}
			if 0.00001 < math.Abs(float64(newD.Golf-d.Golf)) {
				t.Errorf("Invalid Golf %v vs %v", newD.Golf, d.Golf)
			}
			if 0.00001 < math.Abs(float64(newD.Hotel-d.Hotel)) {
				t.Errorf("Invalid %v vs %v", newD.Hotel, d.Hotel)
			}
			if 0.00001 < math.Abs(float64(newD.India-d.India)) {
				t.Errorf("Invalid India")
			}
			if 0.00001 < math.Abs(float64(newD.Juliet-d.Juliet)) {
				t.Errorf("Invalid Julia")
			}
			if 0.00001 < math.Abs(float64(newD.Kilo-d.Kilo)) {
				t.Errorf("Invalid Kilo")
			}
			if 0.00001 < math.Abs(float64(newD.Lima-d.Lima)) {
				t.Errorf("Invalid Lima")
			}
			if 0.00001 < math.Abs(float64(newD.Mike-d.Mike)) {
				t.Errorf("Invalid Mike %v vs %v", newD.Mike, d.Mike)
			}
			if 0.00001 < math.Abs(float64(newD.November-d.November)) {
				t.Errorf("Invalid November %v vs %v", newD.November, d.November)
			}
		}
	}

}

func TestCompleteCase(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(AllCases{})
	assert.Equal(t, nil, errRecipe)

	t.Logf("\nCOMPLETE RECIPE IS \n%v\n\n", recipe)

	//In range
	d := AllCases{
		Alpha:   2.3,
		Bravo:   10.1,
		Charlie: 42,
		Delta:   123,
		Echo:    true,
		Foxtrot: 13.5,
		Golf:    5,
		Hotel:   26.4,

		India:    20.4,
		Juliet:   1.1,
		Kilo:     25,
		Lima:     42,
		Mike:     -21.5,
		November: -90,
		Oscar:    42.69,
		Papa:     112.5,
	}

	byt, errSplurt := recipe.Splurts(d)
	assert.Equal(t, nil, errSplurt)

	t.Logf("SPLURTS %#v (len=%v)\n", byt, len(byt))
	newD := AllCases{}
	e := recipe.UnSplurts(byt, &newD)
	assert.Equal(t, nil, e)

	t.Logf("NewD=%#v\n\n", newD)
	if 0.00001 < math.Abs(newD.Alpha-d.Alpha) {
		t.Errorf("Invalid")
	}
	if 0.00001 < math.Abs(float64(newD.Bravo-d.Bravo)) {
		t.Errorf("Invalid")
	}
	if 0.00001 < math.Abs(float64(newD.Charlie-d.Charlie)) {
		t.Errorf("Invalid")
	}
	if 0.00001 < math.Abs(float64(newD.Delta-d.Delta)) {
		t.Errorf("Invalid")
	}
	if newD.Echo != d.Echo {
		t.Errorf("Invalid")
	}

	if 0.00001 < math.Abs(float64(newD.Foxtrot-d.Foxtrot)) {
		t.Errorf("Invalid Foxtrot %v vs %v", newD.Foxtrot, d.Foxtrot)
	}

	if 0.00001 < math.Abs(float64(newD.Golf-d.Golf)) {
		t.Errorf("Invalid Golf %v vs %v", newD.Golf, d.Golf)
	}

	if 0.00001 < math.Abs(float64(newD.Hotel-d.Hotel)) {
		t.Errorf("Invalid %v vs %v", newD.Hotel, d.Hotel)
	}

	if 0.00001 < math.Abs(float64(newD.India-d.India)) {
		t.Errorf("Invalid")
	}

	if 0.00001 < math.Abs(float64(newD.Juliet-d.Juliet)) {
		t.Errorf("Invalid")
	}

	if 0.00001 < math.Abs(float64(newD.Kilo-d.Kilo)) {
		t.Errorf("Invalid")
	}

	if 0.00001 < math.Abs(float64(newD.Lima-d.Lima)) {
		t.Errorf("Invalid")
	}

	if 0.00001 < math.Abs(float64(newD.Mike-d.Mike)) {
		t.Errorf("Invalid Mike %v vs %v", newD.Mike, d.Mike)
	}

	if 0.00001 < math.Abs(float64(newD.November-d.November)) {
		t.Errorf("Invalid November %v vs %v", newD.November, d.November)
	}

}

//Testing 7bit packing

type SevenBitADC struct {
	AdcVoltage     float64 `splurts:"min=-0.320,max=0.320,bits=16,clamped"`
	MessageCounter byte    `splurts:"min=0,bits=2,clamped"`
}

func TestSevenBitADC(t *testing.T) {
	a := SevenBitADC{AdcVoltage: -0.32, MessageCounter: 1}

	recipe, errRecipe := GetPiecewisesFromStruct(SevenBitADC{})
	assert.Equal(t, nil, errRecipe)

	bitArr, toBitArrErr := recipe.Splurts7bitBytes(a)
	assert.Equal(t, nil, toBitArrErr)

	b := SevenBitADC{}
	errUnsplurt := recipe.UnSplurts7bitBytes(bitArr, &b)
	assert.Equal(t, nil, errUnsplurt)

	if a.MessageCounter != b.MessageCounter {
		t.Errorf("Seven bit err %#v vs %#v,  (arr=%#v)", a, b, bitArr)
	}
}

type Simple7 struct {
	A float64 `splurts:"min=0,step=0.01,bits=9,clamped"` //us
	B float64 `splurts:"min=0,step=0.01,bits=9,clamped"`
	C int     `splurts:"min=0,bits=8,clamped"`            //TODO BITS?
	D float64 `splurts:"min=0,step=0.01,bits=16,clamped"` //us how long coil is running
	E float64 `splurts:"min=-0.320,max=0.320,bits=16,clamped"`
	F int     `splurts:"min=0,bits=5,clamped"`
}

func Test7bitSimpleCase(t *testing.T) {
	a := Simple7{A: 5, B: 1.5, C: 5, D: 15, E: 0, F: 0}
	recipe, errRecipe := GetPiecewisesFromStruct(Simple7{})
	assert.Equal(t, nil, errRecipe)

	bitArr, toBitArrErr := recipe.Splurts7bitBytes(a)
	assert.Equal(t, nil, toBitArrErr)

	b := SevenBitADC{}
	errUnsplurt := recipe.UnSplurts7bitBytes(bitArr, &b)
	assert.Equal(t, nil, errUnsplurt)

}

type InvalidInfStruct struct {
	A     int     `splurts:"min=0,bits=5,clamped"`
	Wrong float64 `splurts:"step=0.1,min=0,max=300,infpos=99999,infneg=-99999,clamped"`
	B     int     `splurts:"min=0,bits=5,clamped"`
}

func TestInvalidInfConf(t *testing.T) {
	_, errInf := GetPiecewisesFromStruct(InvalidInfStruct{})
	assert.NotEqual(t, nil, errInf)
}
func TestCsv(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(ParticleMeas{})
	assert.Equal(t, nil, errRecipe)
	meas1 := ParticleMeas{
		SystemStatus: "MEASURE",
		Temperature:  -40.149,
		Humidity:     110.1291,
		Pressure:     80000.51,
		Small:        301.222,
		Large:        -0.001,
		Extra:        2.1,
	}
	arr := []ParticleMeas{meas1, meas1}
	meas1.Temperature = math.NaN()
	arr = append(arr, meas1)
	meas1.Temperature = 3
	meas1.Large = math.Inf(1)
	arr = append(arr, meas1)
	meas1.Temperature = 5
	meas1.Large = math.Inf(-1)
	arr = append(arr, meas1)
	meas1.Large = 0
	meas1.Small = math.Inf(1)
	arr = append(arr, meas1)

	txt, errCsv := recipe.ToCsv(arr, "\t", []string{}, true)
	assert.Equal(t, nil, errCsv)
	assert.Equal(t, "MEASURE\t-40.1\t42\t110.13\t80001\t301.2\t0.0\t2.1\t0\nMEASURE\t-40.1\t42\t110.13\t80001\t301.2\t0.0\t2.1\t0\nMEASURE\t3.0\t42\t110.13\t80001\t301.2\t99999.0\t2.1\t0\nMEASURE\t5.0\t42\t110.13\t80001\t301.2\t-99999.0\t2.1\t0\nMEASURE\t5.0\t42\t110.13\t80001\t+Inf\t0.0\t2.1\t0\n", txt)

	txtTempHum, errTempHum := recipe.ToCsv(arr, "\t", []string{"Temperature", "Humidity"}, true)
	assert.Equal(t, nil, errTempHum)

	assert.Equal(t, "-40.1\t110.13\n-40.1\t110.13\n3.0\t110.13\n5.0\t110.13\n5.0\t110.13\n", txtTempHum)
}
