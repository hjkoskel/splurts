package messagepack

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/hjkoskel/splurts"
	"github.com/stretchr/testify/assert"
)

type ParticleMeas struct {
	SystemStatus string  `splurts:"enum=UNDEFINED,INITIALIZE,IDLE,MEASURE,STOP,ERROR" messagepack:"stat,rle=3"`
	Temperature  float64 `splurts:"step=0.1,min=-40,max=40" messagepack:"rle=2"`
	StaticSymbol int     `splurts:"bits=7,const=42"`
	Humidity     float64 `splurts:"step=0.05,min=0,max=100" messagepack:"hum,delta=1,rle=2"`
	Pressure     float64 `splurts:"step=100,min=85000,max=110000" messagepack:"name=pres,delta=2"`
	Small        float64 `splurts:"step=0.1,min=0,max=300"  messagepack:"rle=3"`
	Large        float64 `splurts:"step=0.1,min=0,max=300,infpos=99999,infneg=-99999"`
	Extra        float64 `splurts:"step=0.1,min=0,bits=14" messagepack:"name=ext"`
	Heater       bool    `splurts:"enum=NOTHEATING,HEATING"` //Flag for heater enabled
	Emptyvalue   string  `splurts:"enum=NO,MAYBE,YES"`
}

func TestSimple(t *testing.T) {

	testArr := []ParticleMeas{
		ParticleMeas{SystemStatus: "IDLE", Temperature: 22.1, Humidity: 32.4, Pressure: 95200, Small: 69.42, Large: 33, Extra: 2, Heater: false, Emptyvalue: "YES"},
		ParticleMeas{SystemStatus: "MEASURE", Temperature: 22.2, Humidity: 32.6, Pressure: 95100, Small: 67.11, Large: 34, Extra: 3, Heater: true, Emptyvalue: "YES"},
		ParticleMeas{SystemStatus: "MEASURE", Temperature: 22.3, Humidity: 32.6, Pressure: 95100, Small: 67.11, Large: 34, Extra: 3, Heater: true, Emptyvalue: "YES"},
		ParticleMeas{SystemStatus: "MEASURE", Temperature: 22.4, Humidity: 32.6, Pressure: 95100, Small: 67.11, Large: 34, Extra: 3, Heater: true, Emptyvalue: "YES"},
		ParticleMeas{SystemStatus: "MEASURE", Temperature: 22.5, Humidity: 32.6, Pressure: 95100, Small: 67.11, Large: 34, Extra: 3, Heater: true, Emptyvalue: "YES"},
		ParticleMeas{SystemStatus: "MEASURE", Temperature: 22.6, Humidity: 32.6, Pressure: 95500, Small: 67.11, Large: 34, Extra: 3, Heater: true, Emptyvalue: "YES"},
		ParticleMeas{SystemStatus: "STOP", Temperature: 22.3, Humidity: 32.6, Pressure: 95100, Small: 67.11, Large: 34, Extra: 3, Heater: false, Emptyvalue: "YES"},
	}

	//Lets round with splurts (assume that it is working ok)
	recipe, errRecipe := splurts.GetPiecewisesFromStruct(ParticleMeas{})
	if errRecipe != nil {
		t.Errorf("Recipe error %v", errRecipe)
	}

	for i, v := range testArr {
		splurted, splurtErr := recipe.Splurts(v)
		assert.Equal(t, nil, splurtErr)
		unsplurtErr := recipe.UnSplurts(splurted, &testArr[i])
		assert.Equal(t, nil, unsplurtErr)
	}

	code, err := SplurtsArrToMessagepack(recipe, testArr)
	assert.Equal(t, nil, err)

	/*
		fmt.Printf("CODE\n")
		for _, v := range code {
			fmt.Printf("%02X ", v)
		}
		fmt.Printf("\n")
	*/

	metricsBack, backErr := ReadMsgPackMetrics(bytes.NewBuffer(code))
	assert.Equal(t, nil, backErr)
	testmap, _ := recipe.GetValuesToFloatMapArr(testArr)

	testNameMapping := map[string]string{
		"SystemStatus": "stat",
		"Temperature":  "Temperature",
		//"StaticSymbol": "StaticSymbol",
		"Humidity":   "hum",
		"Pressure":   "pres",
		"Small":      "Small",
		"Large":      "Large",
		"Extra":      "ext",
		"Heater":     "Heater",
		"Emptyvalue": "Emptyvalue",
	}

	for inputname, outputname := range testNameMapping {

		arrInput, hazInput := testmap[inputname]
		assert.Equal(t, true, hazInput)
		arrOutput, hazOutput := metricsBack[outputname]
		assert.Equal(t, true, hazOutput)

		//fmt.Printf("VARIABLE %s have %v deltas\n", outputname, arrOutput.Delta)

		outVec, outVecErr := arrOutput.AllValues()
		assert.Equal(t, nil, outVecErr)
		assert.Equal(t, len(arrInput), len(outVec))
		assert.Equal(t, fmt.Sprintf("%.4f", arrInput), fmt.Sprintf("%.4f", outVec))

		//fmt.Printf("\n%s: %.2f  VERSUS %s: %.2f\n\n", inputname, arrInput, outputname, outVec)

	}

}
