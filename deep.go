// Package deep provides function deep.Equal which is like reflect.DeepEqual but
// returns a list of differences. This is helpful when comparing complex types
// like structures and maps.
package deep

import (
	"fmt"
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

func NewDiffer() Differ {
	return Differ{
		// options where zero-value equals default value are omitted
		floatPrecision: 10,
		maxDiff:        10,
	}
}

type Delta []string

func (d Delta) Equal(other Delta) bool {
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

func (d Delta) IsEmpty() bool {
	return len(d) == 0
}

// Equal compares variables a and b, recursing into their structure up to
// MaxDepth levels deep (if greater than zero), and returns a list of differences,
// or nil if there are none. Some differences may not be found if an error is
// also returned.
//
// If a type has an Equal method, like time.Equal, it is called to check for
// equality.
//
// When comparing a struct, if a field has the tag `deep:"-"` then it will be
// ignored.
func (d Differ) Equal(a, b any) Delta {
	c := cmp{
		Differ:      d,
		diff:        []string{},
		buff:        []string{},
		floatFormat: fmt.Sprintf("%%.%df", d.floatPrecision),
	}
	return c.delta(a, b)
}

// FloatPrecision is the number of decimal places to round float values
// to when comparing.
func (d Differ) FloatPrecision(p uint) Differ {
	d.floatPrecision = p
	return d
}

// MaxDiff specifies the maximum number of differences to return.
func (d Differ) MaxDiff(m int) Differ {
	d.maxDiff = uint(m)
	return d
}

// MaxDepth specifies the maximum levels of a struct to recurse into,
// if greater than zero. If zero, there is no limit.
func (d Differ) MaxDepth(m uint) Differ {
	d.maxDepth = m
	return d
}

// LogErrors causes errors to be logged to STDERR when true.
func (d Differ) LogErrors(b bool) Differ {
	d.logErrors = b
	return d
}

// CompareUnexportedFields causes unexported struct fields, like s in
// T{s int}, to be compared when true. This does not work for comparing
// error or Time types on unexported fields because methods on unexported
// fields cannot be called.
func (d Differ) CompareUnexportedFields(b bool) Differ {
	d.compareUnexportedFields = b
	return d
}

// CompareFunctions compares functions the same as reflect.DeepEqual:
// only two nil functions are equal. Every other combination is not equal.
// This is disabled by default because previous versions of this package
// ignored functions. Enabling it can possibly report new diffs.
func (d Differ) CompareFunctions(b bool) Differ {
	d.compareFunctions = b
	return d
}

// NilSlicesAreEmpty causes a nil slice to be equal to an empty slice.
func (d Differ) NilSlicesAreEmpty(b bool) Differ {
	d.nilSlicesAreEmpty = b
	return d
}

// NilMapsAreEmpty causes a nil map to be equal to an empty map.
func (d Differ) NilMapsAreEmpty(b bool) Differ {
	d.nilMapsAreEmpty = b
	return d
}

// NilPointersAreZero causes a nil pointer to be equal to a zero value.
func (d Differ) NilPointersAreZero(b bool) Differ {
	d.nilPointersAreZero = b
	return d
}

// IgnoreSliceOrder causes Equal to ignore slice order so that
// []int{1, 2} and []int{2, 1} are equal. Only slices of primitive scalars
// like numbers and strings are supported. Slices of complex types,
// like []T where T is a struct, are undefined because Equal does not
// recurse into the slice value when this flag is enabled.
func (d Differ) IgnoreSliceOrder(b bool) Differ {
	d.ignoreSliceOrder = b
	return d
}
