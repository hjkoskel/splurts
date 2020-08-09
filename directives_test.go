/*
Testing with directives.

Use this as example
*/

package splurts

import (
	"math"
	"testing"
)

type ParticleMeas struct {
	Temperature float64 `splurts:"step=0.1,min=-40,max=40"`
	Humidity    float64 `splurts:"step=0.05,min=0,max=100"`
	Pressure    float64 `splurts:"step=100,min=85000,max=110000"`
	Small       float64 `splurts:"step=0.1,min=0,max=300"`
	Large       float64 `splurts:"step=0.1,min=0,max=300"`
	Extra    float64 `splurts:"step=0.1,min=0,bits=14"`
	Heater      bool    //Flag for heater enabled
}

func TestInfCase(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(ParticleMeas{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

	d := ParticleMeas{
		Temperature: -40.5,
		Humidity:    110,
		Pressure:    80000,
		Small:       301,
		Large:       -0.001,
		Extra:2.1,
	}

	byt, errSplurt := recipe.Splurts(d)
	if errSplurt != nil {
		t.Errorf(errSplurt.Error())
	} else {
		newD := ParticleMeas{}
		t.Logf("Splurted inf case %#v  (%v bytes)\n", byt, len(byt))
		e := recipe.UnSplurts(byt, &newD)
		if e != nil {
			t.Errorf("Unsplurt err %v", e.Error())
		} else {
			t.Logf("New inf=%#v\n\n", newD)

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

			if !math.IsInf(newD.Large, -1) {
				t.Errorf("Inf error")
			}
		}
	}
}

func TestEasyUseCase(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(ParticleMeas{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

	t.Logf("Easy case\n%v", recipe)

	d := ParticleMeas{Temperature: 18.3, Humidity: 23.6, Pressure: 102400, Small: 23.2, Large: 41}
	byt, errSplurt := recipe.Splurts(d)
	if errSplurt != nil {
		t.Errorf(errSplurt.Error())
	} else {
		newD := ParticleMeas{}
		t.Logf("Splurted easy case %#v  (%v bytes)\n", byt, len(byt))
		e := recipe.UnSplurts(byt, &newD)
		if e != nil {
			t.Errorf("Unsplurt err %v", e.Error())
		} else {
			t.Logf("NewD=%#v\n\n", newD)

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

	India    float64 `splurts:"step=0.1,min=-40,max=40,clamped"`
	Juliet   float32 `splurts:"step=0.1,min=-40,max=40,clamped"`
	Kilo     int32   `splurts:"min=7,max=100,clamped"` //integers have default stepsize 1
	Lima     uint32  `splurts:"max=150,clamped"`
	Mike     float64 `splurts:"min=-100,steps=5.0 10|0.5 100|1.5 10,clamped"`
	November int     `splurts:"min=-100,steps=5.0 10|0.5 100|1.5 10,clamped"`
	Oscar    float64 `splurts:"bits=12,min=-45.5,max=40"`
	Papa     float64 `splurts:"bits=12,min=-45.5,max=40,clamped"`
	Quebeck     float64 `splurts:"bits=12,min=-45.5,step=0.3,bits=12,clamped"`
}

func Test7bitCompleteCase(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(AllCases{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

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
	}

}


func TestCompleteCase(t *testing.T) {
	recipe, errRecipe := GetPiecewisesFromStruct(AllCases{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

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
	if errSplurt != nil {
		t.Errorf(errSplurt.Error())
	} else {
		t.Logf("SPLURTS %#v (len=%v)\n", byt, len(byt))
		newD := AllCases{}
		e := recipe.UnSplurts(byt, &newD)
		if e != nil {
			t.Errorf("Unsplurt err %v", e.Error())
		} else {
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
	}

}

//Testing 7bit packing

type SevenBitADC struct {
	AdcVoltage float64 `splurts:"min=-0.320,max=0.320,bits=16,clamped"`
	MessageCounter      byte    `splurts:"min=0,bits=2,clamped"`
}


func TestSevenBitADC(t *testing.T) {
	a:=SevenBitADC{AdcVoltage:-0.32,MessageCounter:1}

	recipe, errRecipe := GetPiecewisesFromStruct(SevenBitADC{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}


	bitArr,toBitArrErr:=recipe.Splurts7bitBytes(a)
	if toBitArrErr!=nil{
		t.Errorf("Splurt error %v",toBitArrErr.Error())
	}

	b:=SevenBitADC{}
	errUnsplurt:=recipe.UnSplurts7bitBytes(bitArr, &b)
	if errUnsplurt!=nil{
		t.Errorf("Unsplurt error %v",errUnsplurt.Error())
	}
	if a.MessageCounter!=b.MessageCounter{
		t.Errorf("Seven bit err %#v vs %#v,  (arr=%#v)",a,b,bitArr)
	}
}
