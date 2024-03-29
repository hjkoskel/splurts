/*
Coding floats to ints and ints to fixed size binary blobs

*/

package splurts

import (
	"fmt"
	"math"
	"strconv"
)

// PiecewiseCoding  first implementation
type PiecewiseCoding struct {
	Omit    bool   //possible to disable
	Name    string //Needed for floatStruct
	Min     float64
	Steps   []PiecewiseCodingStep
	Clamped bool //No  NaN  -Inf +Inf, just raw value.. Used for flags etc...
	Enums   []string

	InfPosDefined bool
	InfNegDefined bool
	InfPos        float64
	InfNeg        float64

	Const        float64
	ConstDefined bool
	Meta         DirectiveMetadata
}

func (p *PiecewiseCoding) MinStep() float64 {
	if 0 < len(p.Enums) {
		return 1
	}
	if len(p.Steps) == 0 {
		return 0
	}
	minStep := math.Abs(p.Steps[0].Size)
	for _, step := range p.Steps {
		minStep = math.Min(minStep, math.Abs(step.Size))
	}
	return minStep
}

// Decimals Tells how many decimals are required for float. 0=integer 1=0.1 2=0.2
func (p *PiecewiseCoding) Decimals() int {
	if 0 < len(p.Enums) {
		return 0
	}
	if len(p.Steps) == 0 {
		return 0
	}
	/*	minStep := math.Abs(p.Steps[0].Size)
		for _, step := range p.Steps {
			minStep = math.Min(minStep, math.Abs(step.Size))
		}
	*/
	minStep := p.MinStep()
	if 1 < minStep {
		return 0
	}

	return int(math.Ceil(math.Abs(math.Log10(minStep))))
}

func (p *PiecewiseCoding) ToStringValue(f float64) (string, error) {
	if len(p.Enums) == 0 {
		//handling negative zero
		formatstring := fmt.Sprintf("%%.%vf", p.Decimals())
		s := fmt.Sprintf(formatstring, f)

		f2, errInternal := strconv.ParseFloat(s, 64)
		if errInternal != nil {
			return s, errInternal
		}
		if math.Abs(f2) == 0 { //After rounding
			return fmt.Sprintf(formatstring, float64(0)), nil
		}
		return s, nil
	}
	n := int(f)
	if n == 0 {
		return "", nil //First is empty string
	}
	if len(p.Enums) < n || n < 0 {
		return "", fmt.Errorf("have %v enums+empty, index is %v", len(p.Enums), n)
	}
	return p.Enums[n-1], nil
}

func (p PiecewiseCoding) String() string {
	result := fmt.Sprintf("%v: (%v bits", p.Name, p.NumberOfBits())

	if p.Clamped {
		result += ",clamp) "
	} else {
		result += " )"
	}
	result += fmt.Sprintf(" from %v,  ", p.Min)
	for _, step := range p.Steps {
		result += step.String() + " "
	}
	return result
}

// PiecewiseCodingStep size and count  (linter wants comment)
type PiecewiseCodingStep struct {
	Size  float64
	Count uint64
}

// String stringer function for debug printing
func (p PiecewiseCodingStep) String() string {
	return fmt.Sprintf("%v*%vsteps", p.Size, p.Count)
}

// Max helper function
func (p *PiecewiseCoding) Max() float64 {
	result := p.Min
	for _, step := range p.Steps {
		result += step.Size * float64(step.Count)
	}
	return result
}

// IsInvalid Validity checkup.
func (p *PiecewiseCoding) IsInvalid() error {
	if len(p.Name) == 0 {
		return fmt.Errorf("name missing")
	}
	if 0 < len(p.Enums) {
		if !p.Clamped {
			return fmt.Errorf("internal error enums must be clamped not automatic +inf -inf")
		}
		return nil
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("no steps defined at %v", p.Name)
	}

	for i, step := range p.Steps {
		if (step.Size <= 0) || (step.Count <= 0) {
			return fmt.Errorf("invalid step %#v at index %v", step, i)
		}
	}
	return nil
}

// TotalStepCount helper function
func (p *PiecewiseCoding) TotalStepCount() uint64 {
	if 0 < len(p.Enums) {
		return uint64(len(p.Enums)) + 1
	}
	n := uint64(0)
	for _, st := range p.Steps {
		n += st.Count
	}
	return n
}

// NumberOfBits how many bits are spent
func (p *PiecewiseCoding) NumberOfBits() int {
	if p.Omit {
		return 0
	}
	if 0 < len(p.Enums) {
		return int(math.Ceil(math.Log2(float64(len(p.Enums) + 1))))
	}
	n := p.TotalStepCount()
	if p.Clamped {
		return int(math.Ceil(math.Log2(float64(n))))
	}
	// not defined NaN, -inf and +inf needed, ->three extra steps
	return int(math.Ceil(math.Log2(float64(n + 3))))
}

// MaxCode Maximum code possible.
func (p *PiecewiseCoding) MaxCode() uint64 {
	return uint64(math.Pow(2, float64(p.NumberOfBits())) - 1)
}

// Scales float array to uint64 array. TODO optimize for array operation later
func (p *PiecewiseCoding) ScaleToUintArr(fArr []float64) []uint64 {
	//naive solution
	result := make([]uint64, len(fArr))
	for i, f := range fArr {
		result[i] = p.ScaleToUint(f)
	}
	return result
}

// Scales float array to int64 array. TODO optimize for array operation later
func (p *PiecewiseCoding) ScaleToIntArr(fArr []float64) []int64 {
	//naive solution
	result := make([]int64, len(fArr))
	for i, f := range fArr {
		result[i] = int64(p.ScaleToUint(f))
	}
	return result
}

// ScaleToUint converts float to step number.
func (p *PiecewiseCoding) ScaleToUint(f float64) uint64 {

	if 0 < len(p.Enums) {
		return uint64(f)
	}

	if math.IsNaN(f) {
		return p.MaxCode()
	}
	if f < p.Min {
		return 0 //Coded as -inf or in raw as just 0
	}
	total := p.Min
	stepcounter := uint64(0)
	maxcode := p.MaxCode()
	for _, step := range p.Steps {
		a := total
		total += float64(step.Count) * step.Size
		if f <= total {
			result := stepcounter + uint64(math.Round((f-a)/step.Size)) //Round vs floor vs ceil?
			if !p.Clamped {
				result++
			}
			if maxcode < result {
				return maxcode
			}
			return result
		}
		stepcounter += uint64(step.Count)
	}
	return maxcode - 1
}

// BitCode is just wrapper for producing bit string representation from code
func (p *PiecewiseCoding) BitCode(f float64) (string, error) {
	var result string
	formatstring := "%0" + fmt.Sprintf("%v", p.NumberOfBits()) + "b"
	if p.ConstDefined {
		result = fmt.Sprintf(formatstring, p.ScaleToUint(p.Const))
	} else {
		result = fmt.Sprintf(formatstring, p.ScaleToUint(f))
	}
	if len(result) != p.NumberOfBits() {
		return "", fmt.Errorf("bit length %v does not match number of bits %v  (f=%v  piecewise=%#v actual maxCode=%v)\n", len(result), p.NumberOfBits(), f, p, p.MaxCode())
	}
	return result, nil
}

// HexCode to hex code,
func (p *PiecewiseCoding) HexCode(f float64) string {
	numberOfHexChars := int(math.Ceil(float64(p.NumberOfBits()) / 4))
	//Even number
	if numberOfHexChars%2 != 0 {
		numberOfHexChars++
	}

	formatstring := "%0" + fmt.Sprintf("%v", numberOfHexChars) + "X"
	return fmt.Sprintf(formatstring, p.ScaleToUint(f))
}

func (p *PiecewiseCoding) HexCodeMax() string {
	numberOfHexChars := int(math.Ceil(float64(p.NumberOfBits()) / 4))
	//Even number
	if numberOfHexChars%2 != 0 {
		numberOfHexChars++
	}
	result := ""
	for i := 0; i < numberOfHexChars; i++ {
		result += "X"
	}
	return result
}

// ScaleToFloat scales unsigned integer presentation to actual measurement float
func (p *PiecewiseCoding) ScaleToFloat(v uint64) float64 {
	maxv := p.MaxCode()
	if !p.Clamped {
		//NaN for case like where measurement result readout failed due hardware fail
		if v == maxv {
			return math.NaN()
		}
		if v == 0 {
			if p.InfNegDefined {
				return p.InfNeg
			}
			return math.Inf(-1)
		}
		if v == maxv-1 {
			if p.InfPosDefined {
				return p.InfPos
			}
			return math.Inf(1)
		}
	}

	binvalue := uint64(1) // 0=-inf
	if p.Clamped {
		binvalue = 0
	}
	total := p.Min
	for _, step := range p.Steps {
		a := binvalue
		binvalue += uint64(step.Count)
		if v <= binvalue { //Is in this part
			return total + float64(v-a)*step.Size //Round vs floor vs ceil?
		}
		total += float64(step.Count) * step.Size
	}
	if p.Clamped { //Extrapolate up with latest step size. Usually should not need
		return total + float64(v-uint64(p.TotalStepCount()))*p.Steps[len(p.Steps)-1].Size
	}
	if p.InfPosDefined {
		return p.InfPos
	}
	return math.Inf(1)
}
