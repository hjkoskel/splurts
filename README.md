# Splurts golang measurement data coding library

Library for packing up measurement in fixed length binary records.

![Logo](logo.png)

When measuring something with measurement system. For each metric there are
- Range where measured metric is going to be or range where value is significant for specific application
- Sensor + electronics have range and precision
- ADC have number of bits (8/10/12/16/32 etc..). And input range  0v-2.5v etc..
- Signal processing system have maybe input range where it is calibrated.

After (hope so) signal filtering (oversampling).. there is float variable. Usually there multiple metrics that have to be bundled together
(like system status at specific time).

Usually float64 is total overkill (especially when there is 10bit ADC, no oversampling etc..). And storing or transfering data in JSON formatted decimal number spends even many bytes.
Compression algorithms help but it can cause overhead. Some message formats like messagepack, bjson operate by byte basis.

One solution is to scale and pack measurement structs into bytes directly. Unfortunately golang does not provide bit size directive on struct

This library provides way to pack these multiple float metrics into byte array and spend only bits that are needed (depending on range and step size). And also provide function for decoding those bytes back to float map. So scaling is defined in one place

When values are not clamped in signal processing code, this library can store float as NaN -Inf and +Inf codes when value is under or over range.
This feature is sometimes very critical. For example if radiation logger says 3.6 röentgens/h it is it's upper limit. It is nice if log says +Inf than arbitary pre determined limit when actual value is over range. This library also allows NaN value storage on non clamped mode. NaN can be used for signaling that specific value is not available (sensor failure etc...).  Decoding binary have option to report NaN as value or omit it

# Usage by struct directives

The typical use case would be situation where number of floating point measurement results have to be bundled together for logging or communication.

If range is fixed it is possible to write splurts directives on struct (compile time). Like in case of particle measurements

```go
type ParticleMeas struct {
	Temperature float64 `splurts:"step=0.1,min=-40,max=40"`
	Humidity    float64 `splurts:"step=0.05,min=0,max=100"`
	Pressure    float64 `splurts:"step=100,min=85000,max=110000"`
	Small       float64 `splurts:"step=0.1,min=0,max=300"`
	Large       float64 `splurts:"step=0.1,min=0,max=300"`
	Heater      bool    //Status flag for heater enabled
}
```

At the moment splurts supports only non-nested structs

First step is to create PiecewiseFloats from your struct
```go
func GetPiecewisesFromStruct(v interface{}) (PiecewiseFloats, error)
```
PiecewiseFloats documents binary format. On some applications you might want create this struct dynamically at runtime.

Then use Splurts metod for creating byte array.
```go
func (p *PiecewiseFloats) Splurts(input interface{}) ([]byte, error) {
```

When byte array is recieved from communications channel or loaded. Uncompressing is done by creating or loading PiecewiseFloats and doing Unsplurts. (remember & )

```go
func (p *PiecewiseFloats) UnSplurts(raw []byte, output interface{}) error
```

please check unit tests as example how to use library

## Splurt directives

### Simple case
On simple case following parameters are required for float

* *min*, smallest number coded
* *max*, largest number coded
* *step*, stepsize

integer values use step=1 as default. Boolean values do not need any directives
There is optional directive **clamped**. If clamped is present no -Inf, Inf, and NaN codes are not reserved

Also multiple step sizes are supported. It is not recommended to use multiple step sizes on actual range where useful data is.
Might cause problems, disortions etc... on data analysis. Typical use case would be like temperature measurement system have range -40 to 40.
But actual operating temperature is in range -20 to 20.  This would allow to use large step size in zones -40 to -20 and 20 to 40. The significant information on those areas are "how bad it is".

Mandatory parameters are

* *min*, smallest number coded
* *steps*
    * coded in format =size0|count0| size1|count1|size2|count2
      * like 5|10|0.5|20|5|30|   (10 steps 5.0 each then 20 steps each 0.5 then 30 steps 5.0 each)
* *clamped*, optional  (only clamped key)



# All cases example

Following is collection of examples how to use splurts directives

```go
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
}
```
