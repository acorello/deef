package deep

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

var (
	// ErrMaxRecursion is logged when WithMaxDepth is reached.
	ErrMaxRecursion = errors.New("recursed to WithMaxDepth")

	// ErrTypeMismatch is logged when Equal passed two different types of values.
	ErrTypeMismatch = errors.New("variables are different reflect.Type")

	// ErrNotHandled is logged when a primitive Go kind is not handled.
	ErrNotHandled = errors.New("cannot compare the reflect.Kind")
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

type cmp struct {
	Differ
	diff        []string
	buff        []string
	floatFormat string
}

func (c *cmp) delta(a any, b any) Delta {
	if a == nil && b == nil {
		return nil
	} else if a == nil {
		c.saveDiff("<nil pointer>", b)
	} else if b == nil {
		c.saveDiff(a, "<nil pointer>")
	}
	if len(c.diff) > 0 {
		return c.diff
	}

	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)
	c.delta_(aVal, bVal, 0)
	return c.diff
}

func (c *cmp) delta_(a, b reflect.Value, level uint) {
	if c.maxDepth > 0 && level > c.maxDepth {
		c.logError(ErrMaxRecursion)
		return
	}

	// Check if one value is nil, e.g. T{x: *X} and T.x is nil
	if !a.IsValid() || !b.IsValid() {
		if a.IsValid() && !b.IsValid() {
			c.saveDiff(a.Type(), "<nil pointer>")
		} else if !a.IsValid() && b.IsValid() {
			c.saveDiff("<nil pointer>", b.Type())
		}
		return
	}

	// If different types, they can't be equal
	aType := a.Type()
	bType := b.Type()
	if aType != bType {
		// Built-in types don't have a name, so don't report [3]int != [2]int as " != "
		if aType.Name() == "" || aType.Name() != bType.Name() {
			c.saveDiff(aType, bType)
		} else {
			// Type names can be the same, e.g. pkg/v1.Error and pkg/v2.Error
			// are both exported as pkg, so unless we include the full pkg path
			// the diff will be "pkg.Error != pkg.Error"
			// https://github.com/go-test/deep/issues/39
			aFullType := aType.PkgPath() + "." + aType.Name()
			bFullType := bType.PkgPath() + "." + bType.Name()
			c.saveDiff(aFullType, bFullType)
		}
		c.logError(ErrTypeMismatch)
		return
	}

	// Primitive https://golang.org/pkg/reflect/#Kind
	aKind := a.Kind()
	bKind := b.Kind()

	// Do a and b have underlying elements? Yes if they're ptr or interface.
	aElem := aKind == reflect.Ptr || aKind == reflect.Interface
	bElem := bKind == reflect.Ptr || bKind == reflect.Interface

	// If both types implement the error interface, compare the error strings.
	// This must be done before dereferencing because errors.NewWithDefaults() returns a
	// pointer to a struct that implements the interface:
	//   func (e *errorString) Error() string {
	// And we check CanInterface as a hack to make sure the underlying method
	// is callable because https://github.com/golang/go/issues/32438
	// Issues:
	//   https://github.com/go-test/deep/issues/31
	//   https://github.com/go-test/deep/issues/45
	if (aType.Implements(errorType) && bType.Implements(errorType)) &&
		((!aElem || !a.IsNil()) && (!bElem || !b.IsNil())) &&
		(a.CanInterface() && b.CanInterface()) {
		aString := a.MethodByName("Error").Call(nil)[0].String()
		bString := b.MethodByName("Error").Call(nil)[0].String()
		if aString != bString {
			c.saveDiff(aString, bString)
		}
		return
	}

	// Dereference pointers and any
	if aElem || bElem {
		if aElem {
			a = a.Elem()
		}
		if bElem {
			b = b.Elem()
		}
		if aElem && c.nilPointersAreZero && !a.IsValid() && b.IsValid() {
			a = reflect.Zero(b.Type())
		}
		if bElem && c.nilPointersAreZero && !b.IsValid() && a.IsValid() {
			b = reflect.Zero(a.Type())
		}
		c.delta_(a, b, level+1)
		return
	}

	switch aKind {

	/////////////////////////////////////////////////////////////////////
	// Iterable kinds
	/////////////////////////////////////////////////////////////////////

	case reflect.Struct:
		/*
			The variables are structs like:
				type T struct {
					FirstName string
					LastName  string
				}
			Type = <pkg>.T, Kind = reflect.Struct

			Iterate through the fields (FirstName, LastName), recurse into their values.
		*/

		// Types with an Equal() method, like time.Time, only if struct field
		// is exported (CanInterface)
		if eqFunc := a.MethodByName("Equal"); eqFunc.IsValid() && eqFunc.CanInterface() {
			// Handle https://github.com/go-test/deep/issues/15:
			// Don't call T.Equal if the method is from an embedded struct, like:
			//   type Foo struct { time.Time }
			// First, we'll encounter Equal(Ttime, time.Time) but if we pass b
			// as the 2nd arg we'll panic: "Call using pkg.Foo as type time.Time"
			// As far as I can tell, there's no way to see that the method is from
			// time.Time not Foo. So we check the type of the 1st (0) arg and skip
			// unless it's b type. Later, we'll encounter the time.Time anonymous/
			// embedded field and then we'll have Equal(time.Time, time.Time).
			funcType := eqFunc.Type()
			if funcType.NumIn() == 1 && funcType.In(0) == bType {
				retVals := eqFunc.Call([]reflect.Value{b})
				if !retVals[0].Bool() {
					c.saveDiff(a, b)
				}
				return
			}
		}

		for i := 0; i < a.NumField(); i++ {
			if aType.Field(i).PkgPath != "" && !c.compareUnexportedFields {
				continue // skip unexported field, e.g. s in type T struct {s string}
			}

			if aType.Field(i).Tag.Get("deep") == "-" {
				continue // field wants to be ignored
			}

			c.push(aType.Field(i).Name) // push field name to buff

			// Get the Value for each field, e.g. FirstName has Type = string,
			// Kind = reflect.String.
			af := a.Field(i)
			bf := b.Field(i)

			// Recurse to compare the field values
			c.delta_(af, bf, level+1)

			c.pop() // pop field name from buff

			if c.isAtMaxDiff() {
				break
			}
		}
	case reflect.Map:
		/*
			The variables are maps like:
				map[string]int{
					"foo": 1,
					"bar": 2,
				}
			Type = map[string]int, Kind = reflect.Map

			Or:
				type T map[string]int{}
			Type = <pkg>.T, Kind = reflect.Map

			Iterate through the map keys (foo, bar), recurse into their values.
		*/

		if a.IsNil() || b.IsNil() {
			if c.nilMapsAreEmpty {
				if a.IsNil() && b.Len() != 0 {
					c.saveDiff("<nil map>", b)
					return
				} else if a.Len() != 0 && b.IsNil() {
					c.saveDiff(a, "<nil map>")
					return
				}
			} else {
				if a.IsNil() && !b.IsNil() {
					c.saveDiff("<nil map>", b)
				} else if !a.IsNil() && b.IsNil() {
					c.saveDiff(a, "<nil map>")
				}
			}
			return
		}

		if a.Pointer() == b.Pointer() {
			return
		}

		for _, key := range a.MapKeys() {
			c.push(fmt.Sprintf("map[%v]", key))

			aVal := a.MapIndex(key)
			bVal := b.MapIndex(key)
			if bVal.IsValid() {
				c.delta_(aVal, bVal, level+1)
			} else {
				c.saveDiff(aVal, "<does not have key>")
			}

			c.pop()

			if c.isAtMaxDiff() {
				return
			}
		}

		for _, key := range b.MapKeys() {
			if aVal := a.MapIndex(key); aVal.IsValid() {
				continue
			}

			c.push(fmt.Sprintf("map[%v]", key))
			c.saveDiff("<does not have key>", b.MapIndex(key))
			c.pop()
			if c.isAtMaxDiff() {
				return
			}
		}
	case reflect.Array:
		n := a.Len()
		for i := 0; i < n; i++ {
			c.push(fmt.Sprintf("array[%d]", i))
			c.delta_(a.Index(i), b.Index(i), level+1)
			c.pop()
			if c.isAtMaxDiff() {
				break
			}
		}
	case reflect.Slice:
		if c.nilSlicesAreEmpty {
			if a.IsNil() && b.Len() != 0 {
				c.saveDiff("<nil slice>", b)
				return
			} else if a.Len() != 0 && b.IsNil() {
				c.saveDiff(a, "<nil slice>")
				return
			}
		} else {
			if a.IsNil() && !b.IsNil() {
				c.saveDiff("<nil slice>", b)
				return
			} else if !a.IsNil() && b.IsNil() {
				c.saveDiff(a, "<nil slice>")
				return
			}
		}

		// Equal if same underlying pointer and same length, this latter handles
		//   foo := []int{1, 2, 3, 4}
		//   a := foo[0:2] // == {1,2}
		//   b := foo[2:4] // == {3,4}
		// a and b are same pointer but different slices (lengths) of the underlying
		// array, so not equal.
		aLen := a.Len()
		bLen := b.Len()
		if a.Pointer() == b.Pointer() && aLen == bLen {
			return
		}

		if c.ignoreSliceOrder {
			// Compare slices by value and value count; ignore order.
			// Value equality is impliclity established by the maps:
			// any value v1 will hash to the same map value if it's equal
			// to another value v2. Then equality is determiend by value
			// count: presuming v1==v2, then the slics are equal if there
			// are equal numbers of v1 in each slice.
			am := map[any]int{}
			for i := 0; i < a.Len(); i++ {
				am[a.Index(i).Interface()] += 1
			}
			bm := map[any]int{}
			for i := 0; i < b.Len(); i++ {
				bm[b.Index(i).Interface()] += 1
			}
			c.cmpMapValueCounts(a, b, am, bm, true)  // a cmp b
			c.cmpMapValueCounts(b, a, bm, am, false) // b cmp a
		} else {
			// Compare slices by order
			n := aLen
			if bLen > aLen {
				n = bLen
			}
			for i := 0; i < n; i++ {
				c.push(fmt.Sprintf("slice[%d]", i))
				if i < aLen && i < bLen {
					c.delta_(a.Index(i), b.Index(i), level+1)
				} else if i < aLen {
					c.saveDiff(a.Index(i), "<no value>")
				} else {
					c.saveDiff("<no value>", b.Index(i))
				}
				c.pop()
				if c.isAtMaxDiff() {
					break
				}
			}
		}

	/////////////////////////////////////////////////////////////////////
	// Primitive kinds
	/////////////////////////////////////////////////////////////////////

	case reflect.Float32, reflect.Float64:
		// Round floats to WithFloatPrecision decimal places to compare with
		// user-defined precision. As is commonly know, floats have "imprecision"
		// such that 0.1 becomes 0.100000001490116119384765625. This cannot
		// be avoided; it can only be handled. Issue 30 suggested that floats
		// be compared using an epsilon: equal = |a-b| < epsilon.
		// In many cases the result is the same, but I think epsilon is a little
		// less clear for users to reason about. See issue 30 for details.
		aval := fmt.Sprintf(c.floatFormat, a.Float())
		bval := fmt.Sprintf(c.floatFormat, b.Float())
		if aval != bval {
			c.saveDiff(a.Float(), b.Float())
		}
	case reflect.Bool:
		if a.Bool() != b.Bool() {
			c.saveDiff(a.Bool(), b.Bool())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() != b.Int() {
			c.saveDiff(a.Int(), b.Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if a.Uint() != b.Uint() {
			c.saveDiff(a.Uint(), b.Uint())
		}
	case reflect.String:
		if a.String() != b.String() {
			c.saveDiff(a.String(), b.String())
		}
	case reflect.Func:
		if c.compareFunctions {
			if !a.IsNil() || !b.IsNil() {
				aVal, bVal := "nil func", "nil func"
				if !a.IsNil() {
					aVal = "func"
				}
				if !b.IsNil() {
					bVal = "func"
				}
				c.saveDiff(aVal, bVal)
			}
		}
	default:
		c.logError(ErrNotHandled)
	}
}

func (c *cmp) isAtMaxDiff() bool {
	return uint(len(c.diff)) >= c.MaxDiff()
}

func (c *cmp) MaxDiff() uint {
	return c.Differ.maxDiff
}

func (c *cmp) push(name string) {
	c.buff = append(c.buff, name)
}

func (c *cmp) pop() {
	if len(c.buff) > 0 {
		c.buff = c.buff[0 : len(c.buff)-1]
	}
}

func (c *cmp) saveDiff(aval, bval any) {
	if len(c.buff) > 0 {
		varName := strings.Join(c.buff, ".")
		c.diff = append(c.diff, fmt.Sprintf("%s: %v != %v", varName, aval, bval))
	} else {
		c.diff = append(c.diff, fmt.Sprintf("%v != %v", aval, bval))
	}
}

func (c *cmp) cmpMapValueCounts(a, b reflect.Value, am, bm map[any]int, a2b bool) {
	for v := range am {
		aCount, _ := am[v]
		bCount, _ := bm[v]

		if aCount != bCount {
			c.push(fmt.Sprintf("(unordered) slice[]=%v: value count", v))
			if a2b {
				c.saveDiff(fmt.Sprintf("%d", aCount), fmt.Sprintf("%d", bCount))
			} else {
				c.saveDiff(fmt.Sprintf("%d", bCount), fmt.Sprintf("%d", aCount))
			}
			c.pop()
		}
		delete(am, v)
		delete(bm, v)
	}
}

func (c *cmp) logError(err error) {
	if c.logErrors {
		log.Println(err)
	}
}
