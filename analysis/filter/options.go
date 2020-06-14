package filter

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
)

// Options contains information for GrowingFilter
type Options struct {
	// NumIntial is the capacity initially
	NumInitial uint32
	// ScalingFactor decides the amount of increase in size
	ScalingFactor float32
	// FalsePositive is the target false positive rate
	FalsePositive float64
	// TighteningRatio is the factor by which the subsequent false positive decreases
	TighteningRatio float32
	// FillRatio approximates ratio of 1s to size of bitset
	FillRatio float32
}

// DefaultOptions returns the default options which is subjectively
// suitable for the author's use-case
func DefaultOptions() Options {
	return Options{
		NumInitial:      100,
		ScalingFactor:   2,
		FalsePositive:   0.01,
		TighteningRatio: 0.8,
		FillRatio:       0.05,
	}
}

// Encode encodes to the encoder
func (o Options) Encode(e *storage.Encoder) {
	e.WriteUInt32(o.NumInitial)
	e.WriteFloat32(o.ScalingFactor)
	e.WriteFloat64(o.FalsePositive)
	e.WriteFloat32(o.TighteningRatio)
	e.WriteFloat32(o.FillRatio)
}

func OptionsDecode(d *storage.Decoder) Options {
	return Options{
		NumInitial:      d.ReadUInt32(),
		ScalingFactor:   d.ReadFloat32(),
		FalsePositive:   d.ReadFloat64(),
		TighteningRatio: d.ReadFloat32(),
		FillRatio:       d.ReadFloat32(),
	}
}
