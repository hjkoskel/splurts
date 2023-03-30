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
	SystemStatus string  `splurts:"enum=INITIALIZE,IDLE,MEASURE,STOP,ERROR"`
	Temperature float64 `splurts:"step=0.1,min=-40,max=40"`
	StaticSymbol int     `splurts:"bits=7,const=42"`
	Humidity    float64 `splurts:"step=0.05,min=0,max=100"`
	Pressure    float64 `splurts:"step=100,min=85000,max=110000"`
	Small       float64 `splurts:"step=0.1,min=0,max=300"`
	Large       float64 `splurts:"step=0.1,min=0,max=300,infpos=99999,infneg=-99999"`

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
func (p *PiecewiseFloats) Splurts(input interface{}) ([]byte, error)
```

When byte array is recieved from communications channel or loaded. Uncompressing is done by creating or loading PiecewiseFloats and doing Unsplurts. (remember & )

```go
func (p *PiecewiseFloats) UnSplurts(raw []byte, output interface{}) error
```

Splurts library also provides rounding functions for formatting struct float values in efficient way. Typical use would be situation where data is inserted to database by generating insert command
```go
func (p *PiecewiseFloats) ToStrings(input interface{}, quotes bool) (map[string]string, error) 
```

please check unit tests as example how to use library

## Splurt directives

### Simple case
On simple case following parameters are required for float

* *min*, smallest number coded
* *max*, largest number coded
* *step*, stepsize

integer values use step=1 as default. Boolean values do not need any directives
There is optional directive **clamped**. If clamped is present no -Inf, Inf, and NaN codes are not reserved.

But if clamped is not defined, value can become +Inf and -Inf. Some systems (influxdb, JSON etc..) can not handle +Inf and -Inf values. For that reason it is possible to define optional float64 parameters
**infpos** for replacing +Inf  and **infneg** for replacing -Inf

Also multiple step sizes are supported. It is not recommended to use multiple step sizes on actual range where useful data is.
Might cause problems, disortions etc... on data analysis. Typical use case would be like temperature measurement system have range -40 to 40.
But actual operating temperature is in range -20 to 20.  This would allow to use large step size in zones -40 to -20 and 20 to 40. The significant information on those areas are "how bad it is".

Mandatory parameters are

* *min*, smallest number coded
* *steps*
    * coded in format =size0|count0| size1|count1|size2|count2
      * like 5|10|0.5|20|5|30|   (10 steps 5.0 each then 20 steps each 0.5 then 30 steps 5.0 each)
* *clamped*, optional  (only clamped key)


## Enums

Enum directive allows to translate string constant to number. Empty value is coded as 0. Enums are working only with string type

```go
	SystemStatus string  `splurts:"enum=UNDEFINED,INITIALIZE,IDLE,MEASURE,STOP,ERROR"`
```

## Consts
It is possible to define struct variable to constant with directive **constant** . When splurtsing struct to binary, value is overridden by constant definition. When unsplurtsing it is required that binary contains that constant value.

This feature can be used when communication packets are 

## Using time.Time

Latest feature allows to use time.Time variables on struct. Time is stored in millisecond unix epoch format.
User can limit step size (step=1000 means that only seconds are stored) and/or limit min and max values so less bits are consumed

This is still experimental feature

```go
type TimeExampleStruct struct {
	ExampleMetric float64   `splurts:"step=0.1,min=-40,max=40"`
	CompleteTime  time.Time //nanosec spend 42bits
	SecondTime    time.Time `splurts:"min=1670000000000,step=1000"` //uses 31 bits, ends at "Tue Apr 06 2106"
}
```

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

# Exporting

Exporting struct correctly, expecially in case of text format is sometimes hard.
One typical problem is that too few or too many decimals are printed on text formatted output

## Export to CSV format
Function ToCsv takes struct with splurts defined values. Or array/slice of structs with splurt directives.
Parameter separator tells how columns are separated. Typical separator is by tabulator "\t" or ";" for half comma.

Columns is array of variable names that tell order of the colums. This array can left empty. Empty columns mean that all fields are exported to CSV. **skipNaNRows** allows to set code to skip rows in CSV with NaN values. This feature is needed for avoiding NaN values on export.

Values like +Inf and -Inf can be avoided by setting **infpos** and **infneg** on struct

```go
func (p *PiecewiseFloats) ToCsv(input interface{}, separator string, columns []string, skipNaNRows bool) (string, error) {
```
# Messagepack

Experimental feature:

- *SplurtsArrToMessagepack(* creates message pack formatted binary from array of splurts structs.
- *ReadMsgPackMetrics(* reads back message pack struct

Check messagepack/messagepack_test.go as example. Later there will be code for extracting messagepacked metrics for different languages

# Splurt library explained by chatGTP
The text you've provided describes a library that is designed to efficiently store and transmit measurement data in a binary format. The library is intended to be used in systems where multiple metrics are being measured, and where the range and precision of those metrics are known.

The library is designed to take advantage of the fact that different metrics will have different ranges and precisions, and to use only the number of bits that are necessary to represent each metric. For example, if a metric has a range of 0 to 2.5 volts, and an ADC with 10-bit resolution is used to measure that metric, the library will only use 10 bits to represent the measurement value.

The library also provides functions for encoding and decoding the binary data, and allows for the storage of special values such as NaN, Inf, and -Inf, when the measured value is outside of the range. This feature can be critical in situations such as radiation measurement where it is important to know when a value exceeds the measurement range.

Additionally, the library allows handling with clamping of values and NaN during decoding and packing it in a byte array. And also compressing it to reduce bytes used.

# Incoming new features

- support for time.Time conversion to binary format
	- epoch time... timezones..precisions
	- Bad... use epoch+bootcounter?
	- relative time uses omited absolute field?
- name directive support (usually user might want lower case variable name for export)
- directives that link variables
	- One variable is "errorOf" or timedelta of some other variable
	- Variable is derived some other value (maybe not present) like there are can be
		-min/max/avg/std...  each point to same value (that is not present anymore)
		-helps when creating plots or compressing
	- Variable is official timestamp (with defined... is epoch)
- physics related directives for plots and reporting
	- *precision* directive for constant precision (how tightly packed, how many decimals in text printout, error bars etc..) Overrides stepsize for presentation
		- TODO absolute vs percent
		- Comes from stepsize
	- accuracy directive, realistic or tells how much meaningful decimals there are  (not actual use yet, but reserve directive). OR give plusminus symbol to printout.
		- TODO absolute vs percent!!
	- Unit, SI-units and derivates
		- allow smart conversion? 
		- Quality, "voltage","temperature"
		- Grouping and tagging in plot and database exports
	- MeasBw, what is bandwidth in Hz where actual info is.  Used when zooming out (prevent antialiasing). Also reaction time. Can used also when doing lossy compression (filtering, dropping points etc...)
	- SignalBw, signal of bandwidth in measurement (wider than measBw) signal+extra noise band
	- MeasBwStart, where bandwidth starts, default is 0
	- fs, nominal update rate
	- time window, one point. How much it covers in time (like RMS value from window)
	- Equations,  for derivering values
- Tags, for time series database export
- endianess support
- sparse mode
- plotdata export
	- ranges based on maximum ranges of values
	- time axis conversion (that is pain)
- matlab export code data+plots
- protobuf like extraction/de-extraction code generation. OR data export
	- golang, C, javascript, assemblyscript
	- also data export to constant byte arrays
