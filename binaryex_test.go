// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package binaryex

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

type DerivedType uint64

type StructType struct {
	TestField DerivedType
}

type TestType struct {
	BoolField       bool
	IntField        int
	UintField       uint
	Int8Field       int8
	Uint8Field      uint8
	Int16Field      int16
	Uint16Field     uint16
	Int32Field      int32
	Uint32Field     uint32
	Int64Field      int64
	Uint64Field     uint64
	Float32Field    float32
	Float64Field    float64
	Complex64Field  complex64
	Complex128Field complex128
	StringField     string
	ArrayField      [5]byte
	SliceField      []string
	MapField        map[string]int
	StructField     StructType
	TimeField       time.Time
}

func (tt *TestType) Init() {
	tt.BoolField = true
	tt.IntField = -1
	tt.UintField = 1
	tt.Int8Field = -8
	tt.Uint8Field = 8
	tt.Int16Field = -16
	tt.Uint16Field = 16
	tt.Int32Field = -32
	tt.Uint32Field = 32
	tt.Int64Field = -64
	tt.Uint64Field = 64
	tt.Float32Field = 1.61
	tt.Float64Field = 3.14
	tt.Complex64Field = 9
	tt.Complex128Field = 10
	tt.StringField = "String Field"
	tt.ArrayField = [5]byte{1, 2, 3, 4, 5}
	tt.SliceField = []string{"one", "two", "three"}
	tt.MapField = map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	tt.StructField = StructType{
		TestField: 1337,
	}
	tt.TimeField = time.Now()
}

func TestReadWrite(t *testing.T) {

	in := &TestType{}
	in.Init()

	buf := bytes.NewBuffer(nil)
	if err := Write(buf, in); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	fmt.Printf("Buffer: %d\n", buf.Bytes())

	out := &TestType{}
	if err := Read(buf, out); err != nil {
		fmt.Printf("In: %v\n", out)
		t.Fatalf("read failed: %v", err)
	}
	fmt.Printf("IN:\n%v\nOUT:\n%v\n", out, in)
}

func TestNumber(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := 42
	if err := WriteNumber(buf, out); err != nil {
		t.Fatal("WriteNumber Failed", err)
	}
	in := 0
	if err := ReadNumber(buf, &in); err != nil {
		t.Fatal("ReadNumber failed", err)
	}
	if in != out {
		t.Fatalf("Read/Write number missmatch: in: %d, out: %d\n", in, out)
	}
}

func TestString(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := "test"
	if err := WriteString(buf, out); err != nil {
		t.Fatal("WriteString Failed", err)
	}
	in := ""
	if err := ReadString(buf, &in); err != nil {
		t.Fatal("ReadString failed", err)
	}
	if in != out {
		t.Fatalf("Read/Write number missmatch: in: %s, out: %s\n", in, out)
	}
}
