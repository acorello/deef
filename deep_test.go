package deep_test

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
	"unsafe"

	"github.com/acorello/deep"
	v1 "github.com/acorello/deep/test/v1"
	v2 "github.com/acorello/deep/test/v2"
)

func TestString(t *testing.T) {
	d := deep.NewWithDefaults()
	diff := d.Equal("foo", "foo")
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = d.Equal("foo", "bar")
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "foo != bar" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestFloat(t *testing.T) {
	d := deep.NewWithDefaults()
	diff := d.Equal(1.1, 1.1)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = d.Equal(1.1234561, 1.1234562)
	if diff.IsEmpty() {
		t.Error("no diff")
	}

	dFP6, err := deep.New(deep.FloatPrecision(6))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff = dFP6.Equal(1.1234561, 1.1234562)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = dFP6.Equal(1.123456, 1.123457)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "1.123456 != 1.123457" {
		t.Error("wrong diff:", diff[0])
	}

}

func TestInt(t *testing.T) {
	d := deep.NewWithDefaults()
	diff := d.Equal(1, 1)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = d.Equal(1, 2)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != ("1 != 2") {
		t.Error("wrong diff:", diff[0])
	}
}

func TestUint(t *testing.T) {
	d := deep.NewWithDefaults()
	diff := d.Equal(uint(2), uint(2))
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = d.Equal(uint(2), uint(3))
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "2 != 3" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestBool(t *testing.T) {
	d := deep.NewWithDefaults()
	diff := d.Equal(true, true)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = d.Equal(false, false)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = d.Equal(true, false)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "true != false" { // unless you're fipar
		t.Error("wrong diff:", diff[0])
	}
}

func TestTypeMismatch(t *testing.T) {
	type T1 int // same type kind (int)
	type T2 int // but different type
	var t1 T1 = 1
	var t2 T2 = 1
	d := deep.NewWithDefaults()
	diff := d.Equal(t1, t2)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "deep_test.T1 != deep_test.T2" {
		t.Error("wrong diff:", diff[0])
	}

	// Same pkg name but differnet full paths
	// https://github.com/go-test/deep/issues/39
	err1 := v1.Error{}
	err2 := v2.Error{}
	diff = d.Equal(err1, err2)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "github.com/acorello/deep/test/v1.Error != github.com/acorello/deep/test/v2.Error" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestKindMismatch(t *testing.T) {
	var x int = 100
	var y float64 = 100
	d, err := deep.New(deep.LogErrors(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := d.Equal(x, y)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "int != float64" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestDeepRecursion(t *testing.T) {
	type s3 struct {
		S int
	}
	type s2 struct {
		S s3
	}
	type s1 struct {
		S s2
	}
	foo := map[string]s1{
		"foo": { // 1
			S: s2{ // 2
				S: s3{ // 3
					S: 42, // 4
				},
			},
		},
	}
	bar := map[string]s1{
		"foo": {
			S: s2{
				S: s3{
					S: 100,
				},
			},
		},
	}
	// No diffs because MaxDepth=2 prevents seeing the diff at 3rd level down
	dMaxDepth2, err := deep.New(deep.MaxDepth(2))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := dMaxDepth2.Equal(foo, bar)
	if !diff.IsEmpty() {
		t.Errorf("got %d diffs, expected none: %v", len(diff), diff)
	}

	dMaxDepth4, err := deep.New(deep.MaxDepth(4))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff = dMaxDepth4.Equal(foo, bar)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "map[foo].S.S.S: 42 != 100" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestMaxDiff(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7}
	b := []int{0, 0, 0, 0, 0, 0, 0}

	wantDiffLen := 3
	dMaxDiff3, err := deep.New(deep.MaxDiff(wantDiffLen))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := dMaxDiff3.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diffs")
	}

	if gotDiffLen := len(diff); gotDiffLen != wantDiffLen {
		t.Errorf("got %d diffs, expected %d", gotDiffLen, wantDiffLen)
	}

	type fiveFields struct {
		a int // unexported fields require ^
		b int
		c int
		d int
		e int
	}
	t1 := fiveFields{1, 2, 3, 4, 5}
	t2 := fiveFields{0, 0, 0, 0, 0}
	dMaxDiff3UnexportedTrue, err := dMaxDiff3.With(deep.CompareUnexportedFields(true))
	if err != nil {
		t.Fatal("error constructing differ: ", err)
	}
	diff = dMaxDiff3UnexportedTrue.Equal(t1, t2)
	if diff.IsEmpty() {
		t.Fatal("no diffs")
	}
	if len(diff) != wantDiffLen {
		t.Errorf("got %d diffs, expected %d", len(diff), wantDiffLen)
	}

	// Same keys, too many diffs
	m1 := map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}
	m2 := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	}
	diff = dMaxDiff3UnexportedTrue.Equal(m1, m2)
	if diff.IsEmpty() {
		t.Fatal("no diffs")
	}
	if len(diff) != wantDiffLen {
		t.Log(diff)
		t.Errorf("got %d diffs, expected %d", len(diff), wantDiffLen)
	}

	// Too many missing keys
	m1 = map[int]int{
		1: 1,
		2: 2,
	}
	m2 = map[int]int{
		1: 1,
		2: 2,
		3: 0,
		4: 0,
		5: 0,
		6: 0,
		7: 0,
	}
	diff = dMaxDiff3UnexportedTrue.Equal(m1, m2)
	if diff.IsEmpty() {
		t.Fatal("no diffs")
	}
	if len(diff) != wantDiffLen {
		t.Log(diff)
		t.Errorf("got %d diffs, expected %d", len(diff), wantDiffLen)
	}
}

func TestNotHandled(t *testing.T) {
	// UnsafePointer is pretty much the only kind not handled now
	v := []int{1}
	a := unsafe.Pointer(&v)
	b := unsafe.Pointer(&v)
	// UnsafePointer added in Go 1.88. Use these lines once this pkg
	// no longer supports Go 1.17.
	//a := reflect.ValueOf(v).UnsafePointer()
	//b := reflect.ValueOf(v).UnsafePointer()
	diff := deep.NewWithDefaults().Equal(a, b)
	if len(diff) > 0 {
		t.Error("got diffs:", diff)
	}
}

func TestStruct(t *testing.T) {
	type s1 struct {
		id     int
		Name   string
		Number int
	}
	sa := s1{
		id:     1,
		Name:   "foo",
		Number: 2,
	}
	sb := sa
	d := deep.NewWithDefaults()
	diff := d.Equal(sa, sb)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	sb.Name = "bar"
	diff = d.Equal(sa, sb)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "Name: foo != bar" {
		t.Error("wrong diff:", diff[0])
	}

	sb.Number = 22
	diff = d.Equal(sa, sb)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 2 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "Name: foo != bar" {
		t.Error("wrong diff:", diff[0])
	}
	if diff[1] != "Number: 2 != 22" {
		t.Error("wrong diff:", diff[1])
	}

	sb.id = 11
	diff = d.Equal(sa, sb)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 2 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "Name: foo != bar" {
		t.Error("wrong diff:", diff[0])
	}
	if diff[1] != "Number: 2 != 22" {
		t.Error("wrong diff:", diff[1])
	}
}

func TestStructWithTags(t *testing.T) {
	type s1 struct {
		same                    int
		modified                int
		sameIgnored             int `deep:"-"`
		modifiedIgnored         int `deep:"-"`
		ExportedSame            int
		ExportedModified        int
		ExportedSameIgnored     int `deep:"-"`
		ExportedModifiedIgnored int `deep:"-"`
	}
	type s2 struct {
		s1
		same                    int
		modified                int
		sameIgnored             int `deep:"-"`
		modifiedIgnored         int `deep:"-"`
		ExportedSame            int
		ExportedModified        int
		ExportedSameIgnored     int `deep:"-"`
		ExportedModifiedIgnored int `deep:"-"`
		recurseInline           s1
		recursePtr              *s2
	}
	sa := s2{
		s1: s1{
			same:                    0,
			modified:                1,
			sameIgnored:             2,
			modifiedIgnored:         3,
			ExportedSame:            4,
			ExportedModified:        5,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 7,
		},
		same:                    0,
		modified:                1,
		sameIgnored:             2,
		modifiedIgnored:         3,
		ExportedSame:            4,
		ExportedModified:        5,
		ExportedSameIgnored:     6,
		ExportedModifiedIgnored: 7,
		recurseInline: s1{
			same:                    0,
			modified:                1,
			sameIgnored:             2,
			modifiedIgnored:         3,
			ExportedSame:            4,
			ExportedModified:        5,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 7,
		},
		recursePtr: &s2{
			same:                    0,
			modified:                1,
			sameIgnored:             2,
			modifiedIgnored:         3,
			ExportedSame:            4,
			ExportedModified:        5,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 7,
		},
	}
	sb := s2{
		s1: s1{
			same:                    0,
			modified:                10,
			sameIgnored:             2,
			modifiedIgnored:         30,
			ExportedSame:            4,
			ExportedModified:        50,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 70,
		},
		same:                    0,
		modified:                10,
		sameIgnored:             2,
		modifiedIgnored:         30,
		ExportedSame:            4,
		ExportedModified:        50,
		ExportedSameIgnored:     6,
		ExportedModifiedIgnored: 70,
		recurseInline: s1{
			same:                    0,
			modified:                10,
			sameIgnored:             2,
			modifiedIgnored:         30,
			ExportedSame:            4,
			ExportedModified:        50,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 70,
		},
		recursePtr: &s2{
			same:                    0,
			modified:                10,
			sameIgnored:             2,
			modifiedIgnored:         30,
			ExportedSame:            4,
			ExportedModified:        50,
			ExportedSameIgnored:     6,
			ExportedModifiedIgnored: 70,
		},
	}

	d, err := deep.New(deep.CompareUnexportedFields(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	gotDelta := d.Equal(sa, sb)

	want := deep.Delta([]string{
		"s1.modified: 1 != 10",
		"s1.ExportedModified: 5 != 50",
		"modified: 1 != 10",
		"ExportedModified: 5 != 50",
		"recurseInline.modified: 1 != 10",
		"recurseInline.ExportedModified: 5 != 50",
		"recursePtr.modified: 1 != 10",
		"recursePtr.ExportedModified: 5 != 50",
	})
	if !gotDelta.Equal(want) {
		t.Errorf("got %v, want %v", gotDelta, want)
	}
}

func TestNestedStruct(t *testing.T) {
	type s2 struct {
		Nickname string
	}
	type s1 struct {
		Name  string
		Alias s2
	}
	sa := s1{
		Name:  "Robert",
		Alias: s2{Nickname: "Bob"},
	}
	sb := sa
	d := deep.NewWithDefaults()
	diff := d.Equal(sa, sb)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	sb.Alias.Nickname = "Bobby"
	diff = d.Equal(sa, sb)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "Alias.Nickname: Bob != Bobby" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestMap(t *testing.T) {
	ma := map[string]int{
		"foo": 1,
		"bar": 2,
	}
	mb := map[string]int{
		"foo": 1,
		"bar": 2,
	}
	d := deep.NewWithDefaults()
	diff := d.Equal(ma, mb)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = d.Equal(ma, ma)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	mb["foo"] = 111
	diff = d.Equal(ma, mb)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "map[foo]: 1 != 111" {
		t.Error("wrong diff:", diff[0])
	}

	delete(mb, "foo")
	diff = d.Equal(ma, mb)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "map[foo]: 1 != <does not have key>" {
		t.Error("wrong diff:", diff[0])
	}

	diff = d.Equal(mb, ma)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "map[foo]: <does not have key> != 1" {
		t.Error("wrong diff:", diff[0])
	}

	var mc map[string]int
	diff = d.Equal(ma, mc)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	// handle hash order randomness
	if diff[0] != "map[foo:1 bar:2] != <nil map>" && diff[0] != "map[bar:2 foo:1] != <nil map>" {
		t.Error("wrong diff:", diff[0])
	}

	diff = d.Equal(mc, ma)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "<nil map> != map[foo:1 bar:2]" && diff[0] != "<nil map> != map[bar:2 foo:1]" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestArray(t *testing.T) {
	a := [3]int{1, 2, 3}
	b := [3]int{1, 2, 3}

	differ := deep.NewWithDefaults()
	diff := differ.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = differ.Equal(a, a)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	b[2] = 333
	diff = differ.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "array[2]: 3 != 333" {
		t.Error("wrong diff:", diff[0])
	}

	c := [3]int{1, 2, 2}
	diff = differ.Equal(a, c)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "array[2]: 3 != 2" {
		t.Error("wrong diff:", diff[0])
	}

	var d [2]int
	diff = differ.Equal(a, d)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "[3]int != [2]int" {
		t.Error("wrong diff:", diff[0])
	}

	e := [12]int{}
	f := [12]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	diff = differ.Equal(e, f)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	const DefaultMaxDiff = 10
	if len(diff) != DefaultMaxDiff {
		t.Error("not enough diffs:", diff)
	}
	for i := 0; i < DefaultMaxDiff; i++ {
		if diff[i] != fmt.Sprintf("array[%d]: 0 != %d", i+1, i+1) {
			t.Error("wrong diff:", diff[i])
		}
	}
}

func TestSlice(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}

	d := deep.NewWithDefaults()
	diff := d.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	diff = d.Equal(a, a)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	b[2] = 333
	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[2]: 3 != 333" {
		t.Error("wrong diff:", diff[0])
	}

	b = b[0:2]
	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[2]: 3 != <no value>" {
		t.Error("wrong diff:", diff[0])
	}

	diff = d.Equal(b, a)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[2]: <no value> != 3" {
		t.Error("wrong diff:", diff[0])
	}

	var c []int
	diff = d.Equal(a, c)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "[1 2 3] != <nil slice>" {
		t.Error("wrong diff:", diff[0])
	}

	diff = d.Equal(c, a)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "<nil slice> != [1 2 3]" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestSiblingSlices(t *testing.T) {
	father := []int{1, 2, 3, 4}
	a := father[0:3]
	b := father[0:3]

	d := deep.NewWithDefaults()
	diff := d.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
	diff = d.Equal(b, a)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	a = father[0:3]
	b = father[0:2]
	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[2]: 3 != <no value>" {
		t.Error("wrong diff:", diff[0])
	}

	a = father[0:2]
	b = father[0:3]

	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[2]: <no value> != 3" {
		t.Error("wrong diff:", diff[0])
	}

	a = father[0:2]
	b = father[2:4]

	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 2 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[0]: 1 != 3" {
		t.Error("wrong diff:", diff[0])
	}
	if diff[1] != "slice[1]: 2 != 4" {
		t.Error("wrong diff:", diff[1])
	}

	a = father[0:0]
	b = father[1:1]

	diff = d.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
	diff = d.Equal(b, a)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
}

func TestEmptySlice(t *testing.T) {
	a := []int{1}
	b := []int{}
	var c []int

	// Non-empty is not equal to empty.
	d := deep.NewWithDefaults()
	diff := d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[0]: 1 != <no value>" {
		t.Error("wrong diff:", diff[0])
	}

	// Empty is not equal to non-empty.
	diff = d.Equal(b, a)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[0]: <no value> != 1" {
		t.Error("wrong diff:", diff[0])
	}

	// Empty is not equal to nil.
	diff = d.Equal(b, c)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "[] != <nil slice>" {
		t.Error("wrong diff:", diff[0])
	}

	// Nil is not equal to empty.
	diff = d.Equal(c, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "<nil slice> != []" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestNilSlicesAreEmpty(t *testing.T) {

	a := []int{1}
	b := []int{}
	var c []int

	// Empty is equal to nil.
	d, err := deep.New(deep.NilSlicesAreEmpty(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := d.Equal(b, c)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// Nil is equal to empty.
	diff = d.Equal(c, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// Non-empty is not equal to nil.
	diff = d.Equal(a, c)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "[1] != <nil slice>" {
		t.Error("wrong diff:", diff[0])
	}

	// Nil is not equal to non-empty.
	diff = d.Equal(c, a)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "<nil slice> != [1]" {
		t.Error("wrong diff:", diff[0])
	}

	// Non-empty is not equal to empty.
	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[0]: 1 != <no value>" {
		t.Error("wrong diff:", diff[0])
	}

	// Empty is not equal to non-empty.
	diff = d.Equal(b, a)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "slice[0]: <no value> != 1" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestNilMapsAreEmpty(t *testing.T) {

	a := map[int]int{1: 1}
	b := map[int]int{}
	var c map[int]int

	// Empty is equal to nil.
	d, err := deep.New(deep.NilMapsAreEmpty(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := d.Equal(b, c)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// Nil is equal to empty.
	diff = d.Equal(c, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// Non-empty is not equal to nil.
	diff = d.Equal(a, c)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "map[1:1] != <nil map>" {
		t.Error("wrong diff:", diff[0])
	}

	// Nil is not equal to non-empty.
	diff = d.Equal(c, a)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "<nil map> != map[1:1]" {
		t.Error("wrong diff:", diff[0])
	}

	// Non-empty is not equal to empty.
	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "map[1]: 1 != <does not have key>" {
		t.Error("wrong diff:", diff[0])
	}

	// Empty is not equal to non-empty.
	diff = d.Equal(b, a)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "map[1]: <does not have key> != 1" {
		t.Error("wrong diff:", diff[0])
	}
}

func TestNilInterface(t *testing.T) {
	type T struct{ i int }

	a := &T{i: 1}
	d := deep.NewWithDefaults()
	diff := d.Equal(nil, a)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "<nil pointer> != &{1}" {
		t.Error("wrong diff:", diff[0])
	}

	diff = d.Equal(a, nil)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "&{1} != <nil pointer>" {
		t.Error("wrong diff:", diff[0])
	}

	diff = d.Equal(nil, nil)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
}

func TestPointer(t *testing.T) {
	type T struct{ i int }

	a, b := &T{i: 1}, &T{i: 1}
	d := deep.NewWithDefaults()
	diff := d.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	a, b = nil, &T{}
	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "<nil pointer> != deep_test.T" {
		t.Error("wrong diff:", diff[0])
	}

	a, b = &T{}, nil
	diff = d.Equal(a, b)
	if diff.IsEmpty() {
		t.Fatal("no diff")
	}
	if len(diff) != 1 {
		t.Error("too many diff:", diff)
	}
	if diff[0] != "deep_test.T != <nil pointer>" {
		t.Error("wrong diff:", diff[0])
	}

	a, b = nil, nil
	diff = d.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}
}

func TestTime(t *testing.T) {
	// In an interable kind (i.e. a struct)
	type sTime struct {
		T time.Time
	}
	now := time.Now()
	got := sTime{T: now}
	expect := sTime{T: now.Add(1 * time.Second)}
	d := deep.NewWithDefaults()
	diff := d.Equal(got, expect)
	if len(diff) != 1 {
		t.Error("expected 1 diff:", diff)
	}

	// Directly
	a := now
	b := now
	diff = d.Equal(a, b)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// https://github.com/go-test/deep/issues/15
	type Time15 struct {
		time.Time
	}
	a15 := Time15{now}
	b15 := Time15{now}
	diff = d.Equal(a15, b15)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	later := now.Add(1 * time.Second)
	b15 = Time15{later}
	diff = d.Equal(a15, b15)
	if len(diff) != 1 {
		t.Errorf("got %d diffs, expected 1: %s", len(diff), diff)
	}

	// No diff in Equal should not affect diff of other fields (Foo)
	type Time17 struct {
		time.Time
		Foo int
	}
	a17 := Time17{Time: now, Foo: 1}
	b17 := Time17{Time: now, Foo: 2}
	diff = d.Equal(a17, b17)
	if len(diff) != 1 {
		t.Errorf("got %d diffs, expected 1: %s", len(diff), diff)
	}
}

func TestTimeUnexported(t *testing.T) {
	// https://github.com/go-test/deep/issues/18
	// Can't call Call() on exported Value func

	now := time.Now()
	type hiddenTime struct {
		t time.Time
	}
	htA := &hiddenTime{t: now}
	htB := &hiddenTime{t: now}
	d, err := deep.New(deep.CompareUnexportedFields(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := d.Equal(htA, htB)
	if len(diff) > 0 {
		t.Error("should be equal:", diff)
	}

	// This doesn't call time.Time.Equal(), it compares the unexported fields
	// in time.Time, causing a diff like:
	// [t.wall: 13740788835924462040 != 13740788836998203864 t.ext: 1447549 != 1001447549]
	later := now.Add(1 * time.Second)
	htC := &hiddenTime{t: later}
	diff = d.Equal(htA, htC)

	expected := 1
	if _, ok := reflect.TypeOf(htA.t).FieldByName("ext"); ok {
		expected = 2
	}
	if len(diff) != expected {
		t.Errorf("got %d diffs, expected %d: %s", len(diff), expected, diff)
	}
}

func TestInterface(t *testing.T) {
	a := map[string]any{
		"foo": map[string]string{
			"bar": "a",
		},
	}
	b := map[string]any{
		"foo": map[string]string{
			"bar": "b",
		},
	}
	diff := deep.NewWithDefaults().Equal(a, b)
	if len(diff) == 0 {
		t.Fatalf("expected 1 diff, got zero")
	}
	if len(diff) != 1 {
		t.Errorf("expected 1 diff, got %d: %s", len(diff), diff)
	}
}

func TestInterface2(t *testing.T) {
	defer func() {
		if val := recover(); val != nil {
			t.Fatalf("panic: %v", val)
		}
	}()

	a := map[string]any{
		"bar": 1,
	}
	b := map[string]any{
		"bar": 1.23,
	}
	diff := deep.NewWithDefaults().Equal(a, b)
	if len(diff) == 0 {
		t.Fatalf("expected 1 diff, got zero")
	}
	if len(diff) != 1 {
		t.Errorf("expected 1 diff, got %d: %s", len(diff), diff)
	}
}

func TestInterface3(t *testing.T) {
	type Value struct{ int }
	a := map[string]any{
		"foo": &Value{},
	}
	b := map[string]any{
		"foo": 1.23,
	}
	diff := deep.NewWithDefaults().Equal(a, b)
	if len(diff) == 0 {
		t.Fatalf("expected 1 diff, got zero")
	}

	if len(diff) != 1 {
		t.Errorf("expected 1 diff, got: %s", diff)
	}
}

func TestError(t *testing.T) {
	a := errors.New("it broke")
	b := errors.New("it broke")

	d := deep.NewWithDefaults()
	diff := d.Equal(a, b)
	if len(diff) != 0 {
		t.Fatalf("expected zero diffs, got %d: %s", len(diff), diff)
	}

	b = errors.New("it fell apart")
	diff = d.Equal(a, b)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "it broke != it fell apart" {
		t.Errorf("got '%s', expected 'it broke != it fell apart'", diff[0])
	}

	// Both errors set
	type tWithError struct {
		Error error
	}
	t1 := tWithError{
		Error: a,
	}
	t2 := tWithError{
		Error: b,
	}
	diff = d.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "Error: it broke != it fell apart" {
		t.Errorf("got '%s', expected 'Error: it broke != it fell apart'", diff[0])
	}

	// Both errors nil
	t1 = tWithError{
		Error: nil,
	}
	t2 = tWithError{
		Error: nil,
	}
	diff = d.Equal(t1, t2)
	if len(diff) != 0 {
		t.Log(diff)
		t.Fatalf("expected 0 diff, got %d: %s", len(diff), diff)
	}

	// One error is nil
	t1 = tWithError{
		Error: errors.New("foo"),
	}
	t2 = tWithError{
		Error: nil,
	}
	diff = d.Equal(t1, t2)
	if len(diff) != 1 {
		t.Log(diff)
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "Error: *errors.errorString != <nil pointer>" {
		t.Errorf("got '%s', expected 'Error: *errors.errorString != <nil pointer>'", diff[0])
	}
}

func TestErrorWithOtherFields(t *testing.T) {
	a := errors.New("it broke")
	b := errors.New("it fell apart")

	// Both errors set
	type tWithError struct {
		Error error
		Other string
	}
	t1 := tWithError{
		Error: a,
		Other: "ok",
	}
	t2 := tWithError{
		Error: b,
		Other: "ok",
	}
	d := deep.NewWithDefaults()
	diff := d.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "Error: it broke != it fell apart" {
		t.Errorf("got '%s', expected 'Error: it broke != it fell apart'", diff[0])
	}

	// Both errors nil
	t1 = tWithError{
		Error: nil,
		Other: "ok",
	}
	t2 = tWithError{
		Error: nil,
		Other: "ok",
	}
	diff = d.Equal(t1, t2)
	if len(diff) != 0 {
		t.Log(diff)
		t.Fatalf("expected 0 diff, got %d: %s", len(diff), diff)
	}

	// Different Other value
	t1 = tWithError{
		Error: nil,
		Other: "ok",
	}
	t2 = tWithError{
		Error: nil,
		Other: "nope",
	}
	diff = d.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "Other: ok != nope" {
		t.Errorf("got '%s', expected 'Other: ok != nope'", diff[0])
	}

	// Different Other value, same error
	t1 = tWithError{
		Error: a,
		Other: "ok",
	}
	t2 = tWithError{
		Error: a,
		Other: "nope",
	}
	diff = d.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "Other: ok != nope" {
		t.Errorf("got '%s', expected 'Other: ok != nope'", diff[0])
	}
}

type primKindError string

func (e primKindError) Error() string {
	return string(e)
}

func TestErrorPrimitiveKind(t *testing.T) {
	// The primKindError type above is valid and used by Go, e.g.
	// url.EscapeError and url.InvalidHostError. Before fixing this bug
	// (https://github.com/go-test/deep/issues/31), we presumed a and b
	// were ptr or interface (and not nil), so a.Elem() worked. But when
	// a/b are primitive kinds, Elem() causes a panic.
	var err1 primKindError = "abc"
	var err2 primKindError = "abc"
	d := deep.NewWithDefaults()
	diff := d.Equal(err1, err2)
	if len(diff) != 0 {
		t.Fatalf("expected zero diffs, got %d: %s", len(diff), diff)
	}

	err2 = "def"
	diff = d.Equal(err1, err2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
}

func TestErrorUnexported(t *testing.T) {
	// https://github.com/go-test/deep/issues/45
	type foo struct {
		bar error
	}
	e1 := foo{bar: fmt.Errorf("error")}
	e2 := foo{bar: fmt.Errorf("error")}
	d, err := deep.New(deep.CompareUnexportedFields(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	d.Equal(e1, e2)
}

func TestNil(t *testing.T) {
	type student struct {
		name string
		age  int
	}

	mark := student{"mark", 10}
	var someNilThing any = nil
	d := deep.NewWithDefaults()
	diff := d.Equal(someNilThing, mark)
	if diff.IsEmpty() {
		t.Error("Nil value to comparison should not be equal")
	}
	diff = d.Equal(mark, someNilThing)
	if diff.IsEmpty() {
		t.Error("Nil value to comparison should not be equal")
	}
	diff = d.Equal(someNilThing, someNilThing)
	if diff != nil {
		t.Error("Nil value to comparison should not be equal")
	}
}

var testFunc = func() {}

func TestFunc(t *testing.T) {
	// https://github.com/go-test/deep/issues/46
	type TestStruct struct {
		Function func()
	}
	t1 := TestStruct{
		Function: testFunc,
	}
	t2 := TestStruct{
		Function: testFunc,
	}

	// CompareFunctions is off by default, so this should report no diff:
	d := deep.NewWithDefaults()
	diff := d.Equal(t1, t2)
	if len(diff) != 0 {
		t.Fatalf("expected 0 diff when CompareFunctions=false, got %d: %s", len(diff), diff)
	}

	dCompFunc, err := deep.New(deep.CompareFunctions(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}

	// Two funcs are not equal (even if they're the same func)
	diff = dCompFunc.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "Function: func != func" {
		t.Errorf("got '%s', expected 'Function: func != func'", diff[0])
	}

	// One func nil, the other set: not equal
	t1.Function = nil
	diff = dCompFunc.Equal(t1, t2)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "Function: nil func != func" {
		t.Errorf("got '%s', expected 'Function: nil func != func'", diff[0])
	}

	// Two nil funcs are equal
	t1.Function = nil
	t2.Function = nil
	diff = dCompFunc.Equal(t1, t2)
	if len(diff) != 0 {
		t.Errorf("expected 0 diff, got %d: %s", len(diff), diff)
	}
}

func TestSliceOrderString(t *testing.T) {
	// https://github.com/go-test/deep/issues/28

	// These are equal if we ignore order
	a := []string{"foo", "bar"}
	b := []string{"bar", "foo"}
	d, err := deep.New(deep.IgnoreSliceOrder(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := d.Equal(a, b)
	if len(diff) != 0 {
		t.Fatalf("expected 0 diff, got %d: %s", len(diff), diff)
	}

	// Equal with dupes
	a = []string{"foo", "foo", "bar"}
	b = []string{"bar", "foo", "foo"}
	diff = d.Equal(a, b)
	if len(diff) != 0 {
		t.Fatalf("expected 0 diff, got %d: %s", len(diff), diff)
	}

	// NOT equal with dupes
	a = []string{"foo", "foo", "bar"}
	b = []string{"bar", "bar", "foo"}
	diff = d.Equal(a, b)
	if len(diff) != 2 {
		t.Fatalf("expected 2 diff, got %d: %s", len(diff), diff)
	}
	m1 := "(unordered) slice[]=foo: value count: 2 != 1"
	m2 := "(unordered) slice[]=bar: value count: 1 != 2"
	if diff[0] != m1 && diff[0] != m2 {
		t.Errorf("got %s, expected '%s' or '%s'", diff[0], m1, m2)
	}
	if diff[1] != m1 && diff[1] != m2 {
		t.Errorf("got %s, expected '%s' or '%s'", diff[1], m1, m2)
	}

	// NOT equal with one missing
	a = []string{"foo", "bar"}
	b = []string{"bar", "foo", "gone"}
	diff = d.Equal(a, b)
	if len(diff) != 1 {
		t.Fatalf("expected 2 diff, got %d: %s", len(diff), diff)
	}
	if diff[0] != "(unordered) slice[]=gone: value count: 0 != 1" {
		t.Errorf("got %s, expected ''", diff[0])
	}

	// NOT equal at all
	a = []string{"foo", "bar"}
	b = []string{"x"}
	diff = d.Equal(a, b)
	if len(diff) != 3 {
		t.Fatalf("expected 2 diff, got %d: %s", len(diff), diff)
	}
	sort.Strings(diff)
	if diff[0] != "(unordered) slice[]=bar: value count: 1 != 0" {
		t.Errorf("got %s, expected '(unordered) slice[]=bar: value count: 1 != 0'", diff[0])
	}
	if diff[1] != "(unordered) slice[]=foo: value count: 1 != 0" {
		t.Errorf("got %s, expected '(unordered) slice[]=foo: value count: 1 != 0", diff[1])
	}
	if diff[2] != "(unordered) slice[]=x: value count: 0 != 1" {
		t.Errorf("got %s, expected '(unordered) slice[]=x: value count: 0 != 1'", diff[2])
	}
}

func TestSliceOrderStruct(t *testing.T) {
	// https://github.com/go-test/deep/issues/28
	// This is NOT supported but Go is so wonderful that it just happens to work.
	// But again: not supported. So if this test starts to fail or be a problem,
	// it can and should be removed becuase the docs say it's not supported.
	type T struct{ i int }
	a := []T{
		{i: 1},
		{i: 2},
	}
	b := []T{
		{i: 2},
		{i: 1},
	}
	d, err := deep.New(deep.IgnoreSliceOrder(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := d.Equal(a, b)
	if len(diff) != 0 {
		t.Fatalf("expected 0 diff, got %d: %s", len(diff), diff)
	}
}

func TestNilPointersAreZero(t *testing.T) {

	type T struct {
		S *string
	}

	a := T{S: nil}
	b := T{S: new(string)}

	d, err := deep.New(deep.NilPointersAreZero(true))
	if err != nil {
		t.Fatal("error constructing differ:", err)
	}
	diff := d.Equal(a, b)
	if len(diff) != 0 {
		t.Fatalf("expected 0 diff, got %d: %s", len(diff), diff)
	}

	*b.S = "hello"
	diff = d.Equal(a, b)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}

	a.S = new(string)
	*a.S = "hi again"
	b.S = nil
	diff = d.Equal(a, b)
	if len(diff) != 1 {
		t.Fatalf("expected 1 diff, got %d: %s", len(diff), diff)
	}
}
