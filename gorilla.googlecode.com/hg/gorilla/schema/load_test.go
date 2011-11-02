// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"reflect"
	"testing"
)

type TestStruct1 struct {
	// Basic types.
	F01 bool
	F02 float32
	F03 float64
	F04 int
	F05 int8
	F06 int16
	F07 int32
	F08 int64
	F09 string
	F10 uint
	F11 uint8
	F12 uint16
	F13 uint32
	F14 uint64
	// Slices.
	F15 []bool
	F16 []float32
	F17 []float64
	F18 []int
	F19 []int8
	F20 []int16
	F21 []int32
	F22 []int64
	F23 []string
	F24 []uint
	F25 []uint8
	F26 []uint16
	F27 []uint32
	F28 []uint64
	// Maps.
	F29 map[string]bool
	F30 map[string]float32
	F31 map[string]float64
	F32 map[string]int
	F33 map[string]int8
	F34 map[string]int16
	F35 map[string]int32
	F36 map[string]int64
	F37 map[string]string
	F38 map[string]uint
	F39 map[string]uint8
	F40 map[string]uint16
	F41 map[string]uint32
	F42 map[string]uint64
	// Nested structs.
	F43 TestStruct2
}

type TestStruct2 struct {
	F01 string
	F02 **TestStruct2
}

func TestLoad(t *testing.T) {
	v := map[string][]string{
		// Basic types.
		"F01": {"true"},
		"F02": {"4.2"},
		"F03": {"4.3"},
		"F04": {"-42"},
		"F05": {"-43"},
		"F06": {"-44"},
		"F07": {"-45"},
		"F08": {"-46"},
		"F09": {"foo"},
		"F10": {"42"},
		"F11": {"43"},
		"F12": {"44"},
		"F13": {"45"},
		"F14": {"46"},
		// Slices.
		"F15": {"true", "false", "true"},
		"F16": {"4.2", "4.3", "4.4"},
		"F17": {"4.5", "4.6", "4.7"},
		"F18": {"-42", "-43", "-44"},
		"F19": {"-45", "-46", "-47"},
		"F20": {"-48", "-49", "-50"},
		"F21": {"-51", "-52", "-53"},
		"F22": {"-54", "-55", "-56"},
		"F23": {"foo", "bar", "baz"},
		"F24": {"42", "43", "44"},
		"F25": {"45", "46", "47"},
		"F26": {"48", "49", "50"},
		"F27": {"51", "52", "53"},
		"F28": {"54", "55", "56"},
		// Maps.
		"F29.a": {"true"},
		"F29.b": {"false"},
		"F29.c": {"true"},
		"F30.a": {"4.2"},
		"F30.b": {"4.3"},
		"F30.c": {"4.4"},
		"F31.a": {"4.5"},
		"F31.b": {"4.6"},
		"F31.c": {"4.7"},
		"F32.a": {"-42"},
		"F32.b": {"-43"},
		"F32.c": {"-44"},
		"F33.a": {"-45"},
		"F33.b": {"-46"},
		"F33.c": {"-47"},
		"F34.a": {"-48"},
		"F34.b": {"-49"},
		"F34.c": {"-50"},
		"F35.a": {"-51"},
		"F35.b": {"-52"},
		"F35.c": {"-53"},
		"F36.a": {"-54"},
		"F36.b": {"-55"},
		"F36.c": {"-56"},
		"F37.a": {"foo"},
		"F37.b": {"bar"},
		"F37.c": {"baz"},
		"F38.a": {"42"},
		"F38.b": {"43"},
		"F38.c": {"44"},
		"F39.a": {"45"},
		"F39.b": {"46"},
		"F39.c": {"47"},
		"F40.a": {"48"},
		"F40.b": {"49"},
		"F40.c": {"50"},
		"F41.a": {"51"},
		"F41.b": {"52"},
		"F41.c": {"53"},
		"F42.a": {"54"},
		"F42.b": {"55"},
		"F42.c": {"56"},
		// Nested structs.
		"F43.F01":         {"foo"},
		"F43.F02.F02.F01": {"bar"},
	}

	s21 := &TestStruct2{F01: "bar", F02: nil}
	s22 := &TestStruct2{F01: "foo", F02: &s21}
	e := TestStruct1{
		// Basic types.
		F01: true,
		F02: 4.2,
		F03: 4.3,
		F04: -42,
		F05: -43,
		F06: -44,
		F07: -45,
		F08: -46,
		F09: "foo",
		F10: 42,
		F11: 43,
		F12: 44,
		F13: 45,
		F14: 46,
		// Slices.
		F15: []bool{true, false, true},
		F16: []float32{4.2, 4.3, 4.4},
		F17: []float64{4.5, 4.6, 4.7},
		F18: []int{-42, -43, -44},
		F19: []int8{-45, -46, -47},
		F20: []int16{-48, -49, -50},
		F21: []int32{-51, -52, -53},
		F22: []int64{-54, -55, -56},
		F23: []string{"foo", "bar", "baz"},
		F24: []uint{42, 43, 44},
		F25: []uint8{45, 46, 47},
		F26: []uint16{48, 49, 50},
		F27: []uint32{51, 52, 53},
		F28: []uint64{54, 55, 56},
		// Maps.
		F29: map[string]bool{"a": true, "b": false, "c": true},
		F30: map[string]float32{"a": 4.2, "b": 4.3, "c": 4.4},
		F31: map[string]float64{"a": 4.5, "b": 4.6, "c": 4.7},
		F32: map[string]int{"a": -42, "b": -43, "c": -44},
		F33: map[string]int8{"a": -45, "b": -46, "c": -47},
		F34: map[string]int16{"a": -48, "b": -49, "c": -50},
		F35: map[string]int32{"a": -51, "b": -52, "c": -53},
		F36: map[string]int64{"a": -54, "b": -55, "c": -56},
		F37: map[string]string{"a": "foo", "b": "bar", "c": "baz"},
		F38: map[string]uint{"a": 42, "b": 43, "c": 44},
		F39: map[string]uint8{"a": 45, "b": 46, "c": 47},
		F40: map[string]uint16{"a": 48, "b": 49, "c": 50},
		F41: map[string]uint32{"a": 51, "b": 52, "c": 53},
		F42: map[string]uint64{"a": 54, "b": 55, "c": 56},
		// Nested structs.
		F43: TestStruct2{F01: "foo", F02: &s22},
	}

	s := &TestStruct1{}
	_ = Load(s, v)

	// Basic types.
	if s.F01 != e.F01 {
		t.Errorf("F01: %v", s.F01)
	}
	if s.F02 != e.F02 {
		t.Errorf("F02: %v", s.F02)
	}
	if s.F03 != e.F03 {
		t.Errorf("F03: %v", s.F03)
	}
	if s.F04 != e.F04 {
		t.Errorf("F04: %v", s.F04)
	}
	if s.F05 != e.F05 {
		t.Errorf("F05: %v", s.F05)
	}
	if s.F06 != e.F06 {
		t.Errorf("F06: %v", s.F06)
	}
	if s.F07 != e.F07 {
		t.Errorf("F07: %v", s.F07)
	}
	if s.F08 != e.F08 {
		t.Errorf("F08: %v", s.F08)
	}
	if s.F09 != e.F09 {
		t.Errorf("F09: %v", s.F09)
	}
	if s.F10 != e.F10 {
		t.Errorf("F10: %v", s.F10)
	}
	if s.F11 != e.F11 {
		t.Errorf("F11: %v", s.F11)
	}
	if s.F12 != e.F12 {
		t.Errorf("F12: %v", s.F12)
	}
	if s.F13 != e.F13 {
		t.Errorf("F13: %v", s.F13)
	}
	if s.F14 != e.F14 {
		t.Errorf("F14: %v", s.F14)
	}
	// Slices.
	if len(s.F15) != 3 || s.F15[0] != e.F15[0] || s.F15[1] != e.F15[1] || s.F15[2] != e.F15[2] {
		t.Errorf("F15: %v", s.F15)
	}
	if len(s.F16) != 3 || s.F16[0] != e.F16[0] || s.F16[1] != e.F16[1] || s.F16[2] != e.F16[2] {
		t.Errorf("F16: %v", s.F16)
	}
	if len(s.F17) != 3 || s.F17[0] != e.F17[0] || s.F17[1] != e.F17[1] || s.F17[2] != e.F17[2] {
		t.Errorf("F17: %v", s.F17)
	}
	if len(s.F18) != 3 || s.F18[0] != e.F18[0] || s.F18[1] != e.F18[1] || s.F18[2] != e.F18[2] {
		t.Errorf("F18: %v", s.F18)
	}
	if len(s.F19) != 3 || s.F19[0] != e.F19[0] || s.F19[1] != e.F19[1] || s.F19[2] != e.F19[2] {
		t.Errorf("F19: %v", s.F19)
	}
	if len(s.F20) != 3 || s.F20[0] != e.F20[0] || s.F20[1] != e.F20[1] || s.F20[2] != e.F20[2] {
		t.Errorf("F20: %v", s.F20)
	}
	if len(s.F21) != 3 || s.F21[0] != e.F21[0] || s.F21[1] != e.F21[1] || s.F21[2] != e.F21[2] {
		t.Errorf("F21: %v", s.F21)
	}
	if len(s.F22) != 3 || s.F22[0] != e.F22[0] || s.F22[1] != e.F22[1] || s.F22[2] != e.F22[2] {
		t.Errorf("F22: %v", s.F22)
	}
	if len(s.F23) != 3 || s.F23[0] != e.F23[0] || s.F23[1] != e.F23[1] || s.F23[2] != e.F23[2] {
		t.Errorf("F23: %v", s.F23)
	}
	if len(s.F24) != 3 || s.F24[0] != e.F24[0] || s.F24[1] != e.F24[1] || s.F24[2] != e.F24[2] {
		t.Errorf("F24: %v", s.F24)
	}
	if len(s.F25) != 3 || s.F25[0] != e.F25[0] || s.F25[1] != e.F25[1] || s.F25[2] != e.F25[2] {
		t.Errorf("F25: %v", s.F25)
	}
	if len(s.F26) != 3 || s.F26[0] != e.F26[0] || s.F26[1] != e.F26[1] || s.F26[2] != e.F26[2] {
		t.Errorf("F26: %v", s.F26)
	}
	if len(s.F27) != 3 || s.F27[0] != e.F27[0] || s.F27[1] != e.F27[1] || s.F27[2] != e.F27[2] {
		t.Errorf("F27: %v", s.F27)
	}
	if len(s.F28) != 3 || s.F28[0] != e.F28[0] || s.F28[1] != e.F28[1] || s.F28[2] != e.F28[2] {
		t.Errorf("F28: %v", s.F28)
	}
	// Maps.
	if len(s.F29) != 3 || s.F29["a"] != e.F29["a"] || s.F29["b"] != e.F29["b"] || s.F29["c"] != e.F29["c"] {
		t.Errorf("F29: %v", s.F29)
	}
	if len(s.F30) != 3 || s.F30["a"] != e.F30["a"] || s.F30["b"] != e.F30["b"] || s.F30["c"] != e.F30["c"] {
		t.Errorf("F30: %v", s.F30)
	}
	if len(s.F31) != 3 || s.F31["a"] != e.F31["a"] || s.F31["b"] != e.F31["b"] || s.F31["c"] != e.F31["c"] {
		t.Errorf("F31: %v", s.F31)
	}
	if len(s.F32) != 3 || s.F32["a"] != e.F32["a"] || s.F32["b"] != e.F32["b"] || s.F32["c"] != e.F32["c"] {
		t.Errorf("F32: %v", s.F32)
	}
	if len(s.F33) != 3 || s.F33["a"] != e.F33["a"] || s.F33["b"] != e.F33["b"] || s.F33["c"] != e.F33["c"] {
		t.Errorf("F33: %v", s.F33)
	}
	if len(s.F34) != 3 || s.F34["a"] != e.F34["a"] || s.F34["b"] != e.F34["b"] || s.F34["c"] != e.F34["c"] {
		t.Errorf("F34: %v", s.F34)
	}
	if len(s.F35) != 3 || s.F35["a"] != e.F35["a"] || s.F35["b"] != e.F35["b"] || s.F35["c"] != e.F35["c"] {
		t.Errorf("F35: %v", s.F35)
	}
	if len(s.F36) != 3 || s.F36["a"] != e.F36["a"] || s.F36["b"] != e.F36["b"] || s.F36["c"] != e.F36["c"] {
		t.Errorf("F36: %v", s.F36)
	}
	if len(s.F37) != 3 || s.F37["a"] != e.F37["a"] || s.F37["b"] != e.F37["b"] || s.F37["c"] != e.F37["c"] {
		t.Errorf("F37: %v", s.F37)
	}
	if len(s.F38) != 3 || s.F38["a"] != e.F38["a"] || s.F38["b"] != e.F38["b"] || s.F38["c"] != e.F38["c"] {
		t.Errorf("F38: %v", s.F38)
	}
	if len(s.F39) != 3 || s.F39["a"] != e.F39["a"] || s.F39["b"] != e.F39["b"] || s.F39["c"] != e.F39["c"] {
		t.Errorf("F39: %v", s.F39)
	}
	if len(s.F40) != 3 || s.F40["a"] != e.F40["a"] || s.F40["b"] != e.F40["b"] || s.F40["c"] != e.F40["c"] {
		t.Errorf("F40: %v", s.F40)
	}
	if len(s.F41) != 3 || s.F41["a"] != e.F41["a"] || s.F41["b"] != e.F41["b"] || s.F41["c"] != e.F41["c"] {
		t.Errorf("F41: %v", s.F41)
	}
	if len(s.F42) != 3 || s.F42["a"] != e.F42["a"] || s.F42["b"] != e.F42["b"] || s.F42["c"] != e.F42["c"] {
		t.Errorf("F42: %v", s.F42)
	}
	// Nested structs.
	if s.F43.F01 != e.F43.F01 || (*(*(*(*s.F43.F02)).F02)).F01 != (*(*(*(*e.F43.F02)).F02)).F01 {
		t.Errorf("F43: %v", s.F43)
	}
}

func TestErrors(t *testing.T) {
	values := map[string][]string{
		"F01": {"thisisnotabool"},
		"F02": {"thisisnotafloat"},
	}

	s := &TestStruct1{}
	err := Load(s, values)
	if err == nil {
		t.Fatalf("Expected error, received nil")
	}
	schemaErr, ok := err.(*SchemaError)
	if !ok {
		t.Fatalf("Expecting SchemaError")
	}
	if len(schemaErr.Errors()) != 2 {
		t.Fatalf("Expected 2 entries in SchemaError, got %d", len(schemaErr.Errors()))
	}
	f01Error := schemaErr.Error("F01")
	if f01Error == nil {
		t.Errorf("Expected error for 'F01'")
	}
	f02Error := schemaErr.Error("F02")
	if f02Error == nil {
		t.Errorf("Expected error for 'F02'")
	}
}

// ----------------------------------------------------------------------------

type TestStruct3 struct {
	F01 string `schema-name:"custom-name-01"`
	F02 []int  `schema-name:"custom-name-02"`
}

func TestCustomNames(t *testing.T) {
	values := map[string][]string{
		"custom-name-01": {"foo"},
		"custom-name-02": {"42", "43", "44"},
	}

	s := &TestStruct3{}
	_ = Load(s, values)

	if s.F01 != "foo" {
		t.Errorf("F01: %v", s.F01)
	}

	if len(s.F02) != 3 || s.F02[0] != 42 || s.F02[1] != 43 || s.F02[2] != 44 {
		t.Errorf("F02: %v", s.F02)
	}
}

// ----------------------------------------------------------------------------

type stringType string

type TestStruct4 struct {
	F01 stringType
	F02 []stringType
	F03 map[string]stringType
}

func convStringType(v reflect.Value) reflect.Value {
	return reflect.ValueOf(stringType(v.String()))
}

func TestCompositeType(t *testing.T) {
	v := map[string][]string{
		"F01":     {"foo"},
		"F02":     {"foo", "bar", "baz"},
		"F03.foo": {"bar"},
		"F03.baz": {"ding"},
	}

	target := new(TestStruct4)

	AddTypeConverter(reflect.TypeOf(stringType("")), convStringType)

	err := Load(target, v)

	if err != nil {
		t.Errorf("TestComposite. Error: %v", err)
	}

	if target.F01 != stringType(v["F01"][0]) {
		t.Errorf("Expected %v got %v", stringType(v["F01"][0]), target.F01)
	}

	if len(target.F02) != 3 || target.F02[0] != stringType(v["F02"][0]) || target.F02[1] != stringType(v["F02"][1]) || target.F02[2] != stringType(v["F02"][2]) {
		t.Errorf("F02: %v", target.F02)
	}

	if len(target.F03) != 2 || target.F03["foo"] != stringType(v["F03.foo"][0]) || target.F03["baz"] != stringType(v["F03.baz"][0]) {
		t.Errorf("F03: %v", target.F03)
	}
}

// ----------------------------------------------------------------------------
// Example from the docs.

type Phone struct {
	Label  string
	Number string
}

type Person struct {
	Name   string
	Phones []Phone
}

func TestMultiStructField(t *testing.T) {
	v := map[string][]string{
		"Name":          {"Moe"},
		"Phones.Label":  {"home", "office"},
		"Phones.Number": {"111-111", "222-222"},
	}

	person := new(Person)
	err := Load(person, v)

	if err != nil {
		t.Errorf("TestMultiStructField. Error: %v", err)
	}

	if person.Name != v["Name"][0] {
		t.Errorf("Expected %v, got %v", v["Name"][0], person.Name)
	}

	if person.Phones == nil || len(person.Phones) != 2 {
		t.Errorf("Expected 2 items in person.Phones, got %v", person.Phones)
	} else {
		if person.Phones[0].Label != v["Phones.Label"][0] {
			t.Errorf("Expected %v, got %v", v["Phones.Label"][0], person.Phones[0].Label)
		}
		if person.Phones[1].Label != v["Phones.Label"][1] {
			t.Errorf("Expected %v, got %v", v["Phones.Label"][1], person.Phones[1].Label)
		}
		if person.Phones[0].Number != v["Phones.Number"][0] {
			t.Errorf("Expected %v, got %v", v["Phones.Number"][0], person.Phones[0].Number)
		}
		if person.Phones[1].Number != v["Phones.Number"][1] {
			t.Errorf("Expected %v, got %v", v["Phones.Number"][1], person.Phones[1].Number)
		}
	}
}
