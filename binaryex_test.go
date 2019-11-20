// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package binaryex

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

type DerivedType uint64

type StructType struct {
	TimeField time.Time
}

type BaseTypes struct {
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
}

func (tt *BaseTypes) init() {
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
}

type MarshalableTypes struct {
	TimeField time.Time
}

func (mt *MarshalableTypes) init() {
	mt.TimeField = time.Now()
}

type AllTypes struct {
	BaseTypes
	MarshalableTypes
}

func (at *AllTypes) init() {
	at.BaseTypes.init()
	at.MarshalableTypes.init()
}

func TestReadWriteBase(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := BaseTypes{}
	out.init()
	if err := Write(buf, out); err != nil {
		t.Fatal("WriteStruct failed", err)
	}
	in := BaseTypes{}
	if err := Read(buf, &in); err != nil {
		t.Fatal("ReadStruct failed", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Read/Write missmatch: in\n%v, out:\n%v\n", in, out)
	}
}

func TestBool(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := make(map[int]bool)
	out[0] = true
	out[1] = true
	out[2] = false
	if err := WriteBool(buf, out[0]); err != nil {
		t.Fatal("WriteBool failed")
	}
	if err := WriteBool(buf, out[1]); err != nil {
		t.Fatal("WriteBool failed")
	}
	if err := WriteBool(buf, out[2]); err != nil {
		t.Fatal("WriteBool failed")
	}
	in := make(map[int]bool)
	in[0] = false
	in[1] = false
	in[2] = true
	temp := false
	if err := ReadBool(buf, &temp); err != nil {
		t.Fatal("ReadBool failed", err)
	}
	in[0] = temp
	if err := ReadBool(buf, &temp); err != nil {
		t.Fatal("ReadBool failed", err)
	}
	in[1] = temp
	if err := ReadBool(buf, &temp); err != nil {
		t.Fatal("ReadBool failed", err)
	}
	in[2] = temp
	for k, v := range out {
		if in[k] != v {
			t.Fatalf("Read/Write bool missmatch: in %v, out: %v\n", v, in[k])
		}
	}
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

func TestArray(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := [5]byte{0x1, 0x2, 0x3, 0x4, 0x5}
	if err := WriteArray(buf, out); err != nil {
		t.Fatal("WriteArray failed", err)
	}
	in := [5]byte{}
	if err := ReadArray(buf, &in); err != nil {
		t.Fatal("ReadArray failed", err)
	}
	if bytes.Compare(in[:], out[:]) != 0 {
		t.Fatalf("Read/Write array missmatch: in %v, out: %v\n", in, out)
	}
}

func TestSlice(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := []byte{0x1, 0x2, 0x3, 0x4, 0x5}
	if err := WriteSlice(buf, out); err != nil {
		t.Fatal("WriteArray failed", err)
	}
	in := []byte{}
	if err := ReadSlice(buf, &in); err != nil {
		t.Fatal("ReadArray failed", err)
	}
	if bytes.Compare(in, out) != 0 {
		t.Fatalf("Read/Write slice missmatch: in %v, out: %v\n", in, out)
	}
}

func TestMap(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := make(map[string]int)
	out["a"] = 1
	out["b"] = 2
	out["c"] = 3
	if err := WriteMap(buf, out); err != nil {
		t.Fatal("WriteString Failed", err)
	}
	in := make(map[string]int)
	if err := ReadMap(buf, &in); err != nil {
		t.Fatal("ReadString failed", err)
	}
	for k, v := range in {
		if out[k] != v {
			t.Fatalf("Read Write slice missmatch: in %v, out: %v\n", in, out)
		}
	}
}

func TestStructBase(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := BaseTypes{}
	out.init()
	if err := WriteStruct(buf, out); err != nil {
		t.Fatal("WriteStruct failed", err)
	}
	in := BaseTypes{}
	if err := ReadStruct(buf, &in); err != nil {
		t.Fatal("ReadStruct failed", err)
	}

	if !reflect.DeepEqual(BaseTypes(in), BaseTypes(out)) {
		t.Fatalf("Read/Write struct missmatch: in\n%v, out:\n%v\n", in, out)
	}
}

func BenchmarkReadWriteBase(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	val := BaseTypes{}
	for i := 0; i < b.N; i++ {
		val.init()
		if err := Write(buf, val); err != nil {
			b.Fatal(err)
		}
		if err := Read(buf, &val); err != nil {
			b.Fatal(err)
		}
		buf.Truncate(0)
	}
}

func BenchmarkReadWriteAll(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	val := AllTypes{}
	for i := 0; i < b.N; i++ {
		val.init()
		if err := Write(buf, val); err != nil {
			b.Fatal(err)
		}
		if err := Read(buf, &val); err != nil {
			b.Fatal(err)
		}
		buf.Truncate(0)
	}
}
