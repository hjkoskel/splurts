# Messagepack for splurts

** INITIAL RELEASE FOR DEVELOPMENT, MESSAGEPACK FEATURE NOT READY FOR PRODUCTION *****

For transfering storing raw binary series of splurts structs, one alternative is to use messagepack.

This library provides way to store and retieve splurts structs and array of splurt structs in fixed messagepack format.


## Basic usage in go

User have array of structs with splurts directives
```go
func SplurtsArrToMetricArrMap(pw splurts.PiecewiseFloats, input interface{}) (MetricArrMap, error) {
```

Then user have MetricArrMap.  It can be

Saved as binary blob
```go
func (p *MetricArrMap) Write(buf *bytes.Buffer) error {
```
and read back when needed
```go
func ReadMetricsArrMap(buf io.Reader) (MetricArrMap, error) {
```
Or just printed out in text format, with correct number of decimals
```go
func (p *MetricArrMap) TabulateValues(colNames []string, separator string) (string, error) {
```

## Messagepack directives

Most of directives are coming from basic splurt directives. Like min, max, bits, metadata etc...

Under *messagepack* there are three directives
- *name* for renaming metric for messagepack
- *delta* 0=no delta coding, 1=once, 2 twice
- *rle*,  0=not in use 1=more than 1 then use array, at least 3 prefered

Delta coding stores initial value and then how values change from value to value
https://en.wikipedia.org/wiki/Delta_encoding
Delta coding is very useful things like timestamps or sweeped values.

RLE means run length encoding
https://en.wikipedia.org/wiki/Run-length_encoding




## Hierarchy

Splurt struct is mapped as MetricArrMap
MetricArr is one variable from struct (includes values from each array item)

MetricArr contains 

Each MetricArr stores data in DeltaRLEVec

DeltaRLEVec supports delta coding (one and two times pass) and run length compression

## Messagepack format

Messagepack format by itself packs only one struct. In real use case part of message pack is packed inside larger document.
### MetricArr
- *Meta* Meta information, **MetricMeta**
- *Coding* Information how values are coded into integers. **MetricCoding**
- *Enums*, alternative to coding. Array of strings telling value of value defined
- *Steps* array of **MetricStep**
- *Delta* Is delta coding used and how many deltas (derivates) are used 0,1 or 2
- *data* Actual delta encoded (or not) and RLE compressed data, byte array

## MetricMeta
Metadata is for plots etc... good practice to transfer in measurement document for archive purposes
- *Unit*
- *Caption*
- *Accuracy*
- *MaxInterval* , Duration how long before line cuts in time series when there are NaNs.
- *Bandwidth*, float in Hz see how fast transients are possible to catch.

**MetricCoding**
- *Min*, minimum possible value
- *Max*, maximum possible value
- *Clamped*, if clamped not +-inf not needed

**MetricStep**
- *Count*, how many symbols are used for this step
- *Step*, how much one symbol is in


