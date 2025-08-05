// Package deep provides Differ.Compare which is like reflect.DeepEqual but
// returns a list of differences. This is helpful when comparing complex types
// like structures and maps.
package deep

import (
	"fmt"
	"slices"
)

type Differ struct {
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

func NewWithDefaults() (d Differ) {
	d, err := New()
	if err != nil {
		panic(fmt.Errorf("factory method with default params failed: %w", err))
	}
	return d
}

func New(opts ...Opt) (d Differ, err error) {
	d = Differ{
		// options where zero-value equals default value are omitted
		floatPrecision: 10,
		maxDiff:        10,
	}
	for opt := range slices.Values(opts) {
		err = opt(&d)
		if err != nil {
			break
		}
	}
	return d, err
}

type Diffs []string

func (d Diffs) Equal(other Diffs) bool {
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

func (d Diffs) IsEmpty() bool {
	return len(d) == 0
}

// Compare compares variables a and b, recursing into their structure up to
// WithMaxDepth levels deep (if greater than zero), and returns a list of differences,
// or nil if there are none. Some differences may not be found if an error is
// also returned.
//
// If a type has an Compare method, like time.Compare, it is called to check for
// equality.
//
// When comparing a struct, if a field has the tag `deep:"-"` then it will be
// ignored.
func (d Differ) Compare(a, b any) Diffs {
	c := cmp{
		Differ:      d,
		diff:        []string{},
		buff:        []string{},
		floatFormat: fmt.Sprintf("%%.%df", d.floatPrecision),
	}
	return c.delta(a, b)
}

func (d Differ) But(opts ...Opt) (r Differ, err error) {
	r = d
	for opt := range slices.Values(opts) {
		err = opt(&r)
		if err != nil {
			break
		}
	}
	return r, err
}

type Opt func(*Differ) error

// WithFloatPrecision is the number of decimal places to round float values
// to when comparing.
func WithFloatPrecision(p uint) Opt {
	return func(d *Differ) error {
		d.floatPrecision = p
		return nil
	}
}

// WithMaxDiff specifies the maximum number of differences to return.
func WithMaxDiff(m int) Opt {
	return func(d *Differ) error {
		d.maxDiff = uint(m)
		return nil
	}
}

// WithMaxDepth specifies the maximum levels of a struct to recurse into,
// if greater than zero. If zero, there is no limit.
func WithMaxDepth(m uint) Opt {
	return func(d *Differ) error {
		d.maxDepth = m
		return nil
	}
}

// WithLogErrors causes errors to be logged to STDERR when true.
func WithLogErrors(b bool) Opt {
	return func(differ *Differ) error {
		differ.logErrors = b
		return nil
	}
}

// WithCompareUnexportedFields causes unexported struct fields, like s in
// T{s int}, to be compared when true. This does not work for comparing
// error or Time types on unexported fields because methods on unexported
// fields cannot be called.
func WithCompareUnexportedFields(b bool) Opt {
	return func(differ *Differ) error {
		differ.compareUnexportedFields = b
		return nil
	}
}

// WithCompareFunctions compares functions the same as reflect.DeepEqual:
// only two nil functions are equal. Every other combination is not equal.
// This is disabled by default because previous versions of this package
// ignored functions. Enabling it can possibly report new diffs.
func WithCompareFunctions(b bool) Opt {
	return func(differ *Differ) error {
		differ.compareFunctions = b
		return nil
	}
}

// WithNilSlicesAreEmpty causes a nil slice to be equal to an empty slice.
func WithNilSlicesAreEmpty(b bool) Opt {
	return func(differ *Differ) error {
		differ.nilSlicesAreEmpty = b
		return nil
	}
}

// WithNilMapsAreEmpty causes a nil map to be equal to an empty map.
func WithNilMapsAreEmpty(b bool) Opt {
	return func(differ *Differ) error {
		differ.nilMapsAreEmpty = b
		return nil
	}
}

// WithNilPointersAreZero causes a nil pointer to be equal to a zero value.
func WithNilPointersAreZero(b bool) Opt {
	return func(differ *Differ) error {
		differ.nilPointersAreZero = b
		return nil
	}
}

// WithIgnoreSliceOrder causes Compare to ignore slice order so that
// []int{1, 2} and []int{2, 1} are equal. Only slices of primitive scalars
// like numbers and strings are supported. Slices of complex types,
// like []T where T is a struct, are undefined because Compare does not
// recurse into the slice value when this flag is enabled.
func WithIgnoreSliceOrder(b bool) Opt {
	return func(differ *Differ) error {
		differ.ignoreSliceOrder = b
		return nil
	}
}
