# Deep Variable Comparison for Humans

This package is a refactoring of [go-test/deep](https://github.com/go-test/deep). Changes:

- The config and logic is bound to an object rather than the package. This is to avoid the problems of a global shared
  configuration.
- The available options have been transformed into "functional options", you can pass them to the factory function or to
  the `func (Comparator) But` method, if you want to build an amended one.
- Renamed `deep.Equal` to `(Comparator)Compare`, because `Equal` is conventionally given to methods returning a boolean.
  IMO it's also more accurate for what it does.
- The returned value is of a named type: `Diff`, that simply wraps a `map[string]string` (the original return type) and
  offer some utility methods like `(Diff) IsEmpty`.

Beside these changes the aim and comparison logic is identical to the original: in fact all tests have been preserved.

## Note

I've refactored this package to help myself with writing tests and my PR on the original package has not been reviewed
in months.

Having given it a new name I'm restarting the versioning from v1. The 

A massive thank you to Daniel Nichter for having done most of the work!

## Usage

```go
package main_test

import (
	"testing"
	"github.com/acorello/deef"
)

type T struct {
	Name    string
	Numbers []float64
}

func TestDeepEqual(t *testing.T) {
	// Can you spot the difference?
	t1 := T{
		Name:    "Isabella",
		Numbers: []float64{1.13459, 2.29343, 3.010100010},
	}
	t2 := T{
		Name:    "Isabella",
		Numbers: []float64{1.13459, 2.29843, 3.010100010},
	}

	c := deef.NewWithDefaults()

	if diff := c.Compare(t1, t2); !diff.IsEmpty() {
		t.Error(diff)
	}
}
```

```
$ go test
--- FAIL: TestDeepEqual (0.00s)
        main_test.go:25: [Numbers.slice[1]: 2.29343 != 2.29843]
```

The difference is in `Numbers.slice[1]`: the two values aren't equal using Go `==`.
