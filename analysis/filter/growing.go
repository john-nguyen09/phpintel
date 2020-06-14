package filter

import (
	"math"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	cuckoo "github.com/seiflotfy/cuckoofilter"
)

// This code is inspired by https://github.com/aronszanto/Dynamic-Size-Bloom-Filter

// GrowingFilter is a filter that can be growing based
// on the BloomFilter
type GrowingFilter struct {
	head    *cuckoo.Filter // The current filter
	headCap uint64         // The capacity for head
	filters []*cuckoo.Filter
	opts    Options
}

// NewGrowing creates a growing filter with given options
func NewGrowing(opts Options) *GrowingFilter {
	if opts.NumInitial <= 0 {
		opts.NumInitial = 1
	}
	filter := cuckoo.NewFilter(uint(opts.NumInitial))
	return &GrowingFilter{
		head:    filter,
		headCap: uint64(opts.NumInitial),
		filters: []*cuckoo.Filter{filter},
		opts:    opts,
	}
}

// Insert grows if growing is required and insert data into the bloom filter
func (g *GrowingFilter) Insert(data []byte) *GrowingFilter {
	if g.approximateFillRatio() > float64(g.opts.FillRatio) {
		newFilter := cuckoo.NewFilter(uint(float32(g.opts.NumInitial) * g.opts.ScalingFactor))
		g.head = newFilter
		g.filters = append(g.filters, newFilter)
	}
	g.head.Insert(data)
	return g
}

// Lookup checks through all filters
func (g *GrowingFilter) Lookup(data []byte) bool {
	for _, filter := range g.filters {
		if filter.Lookup(data) {
			return true
		}
	}
	return false
}

// Reset clears all the filters to free memory
func (g *GrowingFilter) Reset() *GrowingFilter {
	g.head = cuckoo.NewFilter(uint(g.opts.NumInitial))
	g.filters = []*cuckoo.Filter{g.head}
	return g
}

// Encode encodes the growing filter into byte slice
func (g *GrowingFilter) Encode(e *storage.Encoder) {
	e.WriteUInt64(g.headCap)
	g.opts.Encode(e)
	e.WriteInt(len(g.filters))
	for _, filter := range g.filters {
		buffer := filter.Encode()
		e.WriteBytes(buffer)
	}
}

// GrowingFilterDecode decodes a growing filter from a byte slice
func GrowingFilterDecode(d *storage.Decoder) *GrowingFilter {
	g := &GrowingFilter{
		headCap: d.ReadUInt64(),
		opts:    OptionsDecode(d),
	}
	filterLength := d.ReadInt()
	for i := 0; i < filterLength; i++ {
		filter, err := cuckoo.Decode(d.ReadBytes())
		if err != nil {
			panic(err)
		}
		g.filters = append(g.filters, filter)
	}
	g.head = g.filters[len(g.filters)-1]
	return g
}

func (g *GrowingFilter) approximateFillRatio() float64 {
	return 1.0 - math.Exp(-float64(g.head.Count())/float64(g.headCap))
}
