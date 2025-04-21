// Package deef provides function deef.Diff which is like reflect.DeepEqual but
// returns a list of differences. This is helpful when comparing complex types
// like structures and maps.
package deef

import (
	"fmt"
	"slices"
)

type Comparison struct {
	compareFunctions        bool
	compareUnexportedFields bool
	floatPrecision          uint
	ignoreSliceOrder        bool
	logErrors               bool
	maxDepth                uint
	maxDiff                 uint
	nilMapsAreEmpty         bool
	nilPointersAreZero      bool
	nilSlicesAreEmpty       bool
}

func NewComparison() (c Comparison) {
	c, err := NewComparisonWith()
	if err != nil {
		panic(fmt.Errorf("factory method with default params failed: %w", err))
	}
	return c
}

func NewComparisonWith(opts ...Option) (c Comparison, err error) {
	c = Comparison{
		// options where zero-value equals default value are omitted
		floatPrecision: 10,
		maxDiff:        10,
	}
	for opt := range slices.Values(opts) {
		err = opt(&c)
		if err != nil {
			break
		}
	}
	return c, err
}

// Compare returns the differences between a and b. If they are compound types it recursively compares each
// of their members up to configured MaxDepth (default is unlimited).
//
// If any type encountered has an `Equal` method it is used to determine equality.
//
// When comparing a struct, if a field has the tag `deef:"-"` then it will be ignored.
//
// Some differences may not be found if an error is also returned.
func (c Comparison) Compare(a, b any) Diff {
	cmp := comparator{
		config:      c,
		diff:        []string{},
		buff:        []string{},
		floatFormat: fmt.Sprintf("%%.%df", c.floatPrecision),
	}
	return cmp.compare(a, b)
}

type Diff []string

func (d Diff) Equal(other Diff) bool {
	if len(d) != len(other) {
		return false
	}
	for i := range d {
		if d[i] != other[i] {
			return false
		}
	}
	return true
}

func (d Diff) IsEmpty() bool {
	return len(d) == 0
}

// Option changes a Comparison configuration
type Option func(*Comparison) error

// FloatPrecision is the number of decimal places to round float values
// to when comparing.
func FloatPrecision(p uint) Option {
	return func(d *Comparison) error {
		d.floatPrecision = p
		return nil
	}
}

// MaxDiff specifies the maximum number of differences to return.
func MaxDiff(m int) Option {
	return func(d *Comparison) error {
		d.maxDiff = uint(m)
		return nil
	}
}

// MaxDepth specifies the maximum levels of a struct to recurse into,
// if greater than zero. If zero, there is no limit.
func MaxDepth(m uint) Option {
	return func(c *Comparison) error {
		c.maxDepth = m
		return nil
	}
}

// LogErrors causes errors to be logged to STDERR when true.
func LogErrors(b bool) Option {
	return func(differ *Comparison) error {
		differ.logErrors = b
		return nil
	}
}

// CompareUnexportedFields causes unexported struct fields, like s in
// T{s int}, to be compared when true. This does not work for comparing
// error or Time types on unexported fields because methods on unexported
// fields cannot be called.
func CompareUnexportedFields(b bool) Option {
	return func(differ *Comparison) error {
		differ.compareUnexportedFields = b
		return nil
	}
}

// CompareFunctions compares functions the same as reflect.DeepEqual:
// only two nil functions are equal. Every other combination is not equal.
// This is disabled by default because previous versions of this package
// ignored functions. Enabling it can possibly report new diffs.
func CompareFunctions(b bool) Option {
	return func(differ *Comparison) error {
		differ.compareFunctions = b
		return nil
	}
}

// NilSlicesAreEmpty causes a nil slice to be equal to an empty slice.
func NilSlicesAreEmpty(b bool) Option {
	return func(differ *Comparison) error {
		differ.nilSlicesAreEmpty = b
		return nil
	}
}

// NilMapsAreEmpty causes a nil map to be equal to an empty map.
func NilMapsAreEmpty(b bool) Option {
	return func(differ *Comparison) error {
		differ.nilMapsAreEmpty = b
		return nil
	}
}

// NilPointersAreZero causes a nil pointer to be equal to a zero value.
func NilPointersAreZero(b bool) Option {
	return func(differ *Comparison) error {
		differ.nilPointersAreZero = b
		return nil
	}
}

// IgnoreSliceOrder causes Comparison to ignore slice order so that
// []int{1, 2} and []int{2, 1} are equal. Only slices of primitive scalars
// like numbers and strings are supported. Slices of complex types,
// like []T where T is a struct, are undefined because Diff does not
// recurse into the slice value when this flag is enabled.
func IgnoreSliceOrder(b bool) Option {
	return func(differ *Comparison) error {
		differ.ignoreSliceOrder = b
		return nil
	}
}
