/*
Coding floats to ints and ints to fixed size binary blobs

*/

package splurts

import (
	"fmt"
	"math"
)

//PiecewiseCoding  first implementation
type PiecewiseCoding struct {
	Name    string //Needed for floatStruct
	Min     float64
	Steps   []PiecewiseCodingStep
	Clamped bool //No  NaN  -Inf +Inf, just raw value.. Used for flags etc...
	//TODO LittleEndianize bool //Little endianize bytes if possible Splurts is usually using non byte multiple fields
}

func (p PiecewiseCoding) String() string {
	result := fmt.Sprintf("%v: (%v bits", p.Name, p.NumberOfBits())

	if p.Clamped {
		result += ",clamp) "
	}
	result += fmt.Sprintf(" from %v,  ", p.Min)
	for _, step := range p.Steps {
		result += step.String() + " "
	}
	return result
}

//PiecewiseCodingStep size and count  (linter wants comment)
type PiecewiseCodingStep struct {
	Size  float64
	Count uint64
}

//String stringer function for debug printing
func (p PiecewiseCodingStep) String() string {
	return fmt.Sprintf("%v*%vsteps", p.Size, p.Count)
}

//Max helper function
func (p *PiecewiseCoding) Max() float64 {
	result := p.Min
	for _, step := range p.Steps {
		result += step.Size * float64(step.Count)
	}
	return result
}

//IsInvalid Validity checkup.
func (p *PiecewiseCoding) IsInvalid() error {
	if len(p.Steps) == 0 {
		return fmt.Errorf("No steps defined at %v", p.Name)
	}
	for i, step := range p.Steps {
		if (step.Size <= 0) || (step.Count <= 0) {
			return fmt.Errorf("invalid step %#v at index %v", step,i)
		}
	}
	return nil
}

//TotalStepCount helper function
func (p *PiecewiseCoding) TotalStepCount() uint64 {
	n := uint64(0)
	for _, st := range p.Steps {
		n += st.Count
	}
	return n
}

//NumberOfBits how many bits are spent
func (p *PiecewiseCoding) NumberOfBits() int {
	n := p.TotalStepCount()
	if p.Clamped {
		return int(math.Ceil(math.Log2(float64(n))))
	}
	// not defined NaN, -inf and +inf needed, ->three extra steps
	return int(math.Ceil(math.Log2(float64(n + 3))))
}

//MaxCode Maximum code possible.
func (p *PiecewiseCoding) MaxCode() uint64 {
	return uint64(math.Pow(2, float64(p.NumberOfBits())) - 1)
}

//ScaleToUint converts float to step number.
func (p *PiecewiseCoding) ScaleToUint(f float64) uint64 {
	if math.IsNaN(f) {
		return p.MaxCode()
	}
	if f < p.Min {
		return 0 //Coded as -inf or in raw as just 0
	}
	total := p.Min
	stepcounter := uint64(0)
	for _, step := range p.Steps {
		a := total
		total += float64(step.Count) * step.Size
		if f <= total {
			result := stepcounter + uint64(math.Round((f-a)/step.Size)) //Round vs floor vs ceil?
			if !p.Clamped {
				result++
			}
			return result
		}
		stepcounter += uint64(step.Count)
	}
	return p.MaxCode() - 1
}

//BitCode is just wrapper for producing bit string representation from code
func (p *PiecewiseCoding) BitCode(f float64) string {
	formatstring := "%0" + fmt.Sprintf("%v", p.NumberOfBits()) + "b"
	return fmt.Sprintf(formatstring, p.ScaleToUint(f))
}

//HexCode to hex code,
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

/* TODO TRASH!!!
func (p *PiecewiseCoding) HexNybbleCode(f float64) string {
	numberOfHexChars := int(math.Ceil(float64(p.NumberOfBits()) / 4))
	formatstring := "%0" + fmt.Sprintf("%v", numberOfHexChars) + "X"
	return fmt.Sprintf(formatstring, p.ScaleToUint(f))
}

func (p *PiecewiseCoding) HexNybbleCodeMax() string {
	numberOfHexChars := int(math.Ceil(float64(p.NumberOfBits()) / 4))
	result := ""
	for i := 0; i < numberOfHexChars; i++ {
		result += "X"
	}
	return result
}
*/

//ScaleToFloat scales unsigned integer presentation to actual measurement float
func (p *PiecewiseCoding) ScaleToFloat(v uint64) float64 {
	maxv := p.MaxCode()
	if !p.Clamped {
		//NaN for case like where measurement result readout failed due hardware fail
		if v == maxv {
			return math.NaN()
		}
		if v == 0 {
			return math.Inf(-1)
		}
		if v == maxv-1 {
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
	return math.Inf(1)
}
