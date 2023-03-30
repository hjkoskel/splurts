package messagepack

import (
	"fmt"
	"io"
)

type MetricCoding struct {
	Min     float64
	Max     float64
	Clamped bool
}

func ReadMetricCoding(buf io.Reader) (MetricCoding, error) {
	result := MetricCoding{}
	n, errn := ReadFixmap(buf)
	if errn != nil {
		return result, nil
	}
	for i := 0; i < int(n); i++ {
		itemname, errItemName := ReadString(buf)
		if errItemName != nil {
			return MetricCoding{}, fmt.Errorf("%s on ReadMetricCoding item name", errItemName.Error())
		}
		var readErr error
		switch itemname {
		case MPNAME_CODING_MIN:
			result.Min, readErr = ReadNumber(buf)
		case MPNAME_CODING_MAX:
			result.Max, readErr = ReadNumber(buf)
		case MPNAME_CODING_CLAMPED:
			result.Clamped, readErr = ReadBool(buf)
		}
		if readErr != nil {
			return result, readErr
		}
	}
	return result, nil
}

func (p *MetricCoding) Write(w io.Writer) error {
	err := WriteFixmap(w, 3)
	if err != nil {
		return err
	}

	err = WriteString(w, MPNAME_CODING_MIN)
	if err != nil {
		return err
	}
	err = writeFloat64(w, p.Min)
	if err != nil {
		return err
	}
	err = WriteString(w, MPNAME_CODING_MAX)
	if err != nil {
		return err
	}
	err = writeFloat64(w, p.Max)
	if err != nil {
		return err
	}
	err = WriteString(w, MPNAME_CODING_CLAMPED)
	if err != nil {
		return err
	}
	return WriteBool(w, p.Clamped)
}

type MetricStep struct {
	Count int64
	Step  float64
}

func ReadMetricStep(buf io.Reader) (MetricStep, error) {
	result := MetricStep{}
	n, errn := ReadFixmap(buf)
	if errn != nil {
		return result, nil
	}
	for i := 0; i < int(n); i++ {
		itemname, errItemName := ReadString(buf)
		if errItemName != nil {
			return MetricStep{}, fmt.Errorf("%v on ReadMetricStep itemname", errItemName)
		}
		var readErr error
		switch itemname {
		case MPNAME_POINTCOUNT:
			result.Count, readErr = ReadInt(buf)
		case MPNAME_STEPSIZE:
			result.Step, readErr = ReadNumber(buf)
		}
		if readErr != nil {
			return result, readErr
		}
	}
	return result, nil
}

func (p *MetricStep) Write(w io.Writer) error {
	err := WriteFixmap(w, 2)
	if err != nil {
		return err
	}
	err = WriteString(w, MPNAME_POINTCOUNT)
	if err != nil {
		return err
	}
	err = WriteInt(w, p.Count)
	if err != nil {
		return err
	}
	err = WriteString(w, MPNAME_STEPSIZE)
	if err != nil {
		return err
	}
	return writeFloat64(w, p.Step)
}

// Same as in splurts.. But trying to make this lib not depending on splurts
type MetricMeta struct {
	Unit        string
	Caption     string
	Accuracy    string  //Value or metric name
	MaxInterval int64   //Nanosec duration how long before line cuts in time series when there are NaNs. If not defined NaN terminates immediately
	Bandwidth   float64 //-3dB point, see how fast transients are possible to catch. Is
}

func ReadMetricMeta(buf io.Reader) (MetricMeta, error) {
	result := MetricMeta{}

	n, errn := ReadFixmap(buf)
	if errn != nil {
		return result, nil
	}
	for i := 0; i < int(n); i++ {
		itemname, errItemName := ReadString(buf)
		if errItemName != nil {
			return result, fmt.Errorf("%v on ReadMetricMeta itemname", errItemName.Error())
		}
		var readErr error
		switch itemname {
		case MPNAME_META_ACCURACY:
			result.Accuracy, readErr = ReadString(buf)
		case MPNAME_META_BANDWIDTH:
			result.Bandwidth, readErr = ReadNumber(buf)
		case MPNAME_META_CAPTION:
			result.Caption, readErr = ReadString(buf)
		case MPNAME_META_MAXINTERVAL:
			result.MaxInterval, readErr = ReadInt(buf)
		case MPNAME_META_UNIT:
			result.Unit, readErr = ReadString(buf)
		}
		if readErr != nil {
			return result, fmt.Errorf("error on ReadMetricMeta while reading item %s got error %s", itemname, readErr)
		}
	}
	return result, nil
}

func (p *MetricMeta) Write(w io.Writer) error {
	err := WriteFixmap(w, 5)
	if err != nil {
		return err
	}

	err = WriteStringMapString(w, map[string]string{
		MPNAME_META_UNIT:     p.Unit,
		MPNAME_META_CAPTION:  p.Caption,
		MPNAME_META_ACCURACY: p.Accuracy,
	})
	if err != nil {
		return err
	}

	err = WriteString(w, MPNAME_META_MAXINTERVAL)
	if err != nil {
		return err
	}
	err = writeUInt(w, uint64(p.MaxInterval))
	if err != nil {
		return err
	}
	err = WriteString(w, MPNAME_META_BANDWIDTH)
	if err != nil {
		return err
	}
	err = writeFloat64(w, p.Bandwidth)
	if err != nil {
		return err
	}
	return nil
}
