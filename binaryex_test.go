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

type PointerTypes struct {
	PBoolField   *bool
	PStringField *string
	PStructField *BaseTypes
}

func (pt *PointerTypes) init() {
	bv := true
	pt.PBoolField = &bv
	sv := "teststring"
	pt.PStringField = &sv
	pt.PStructField = &BaseTypes{}
	pt.PStructField.init()
}

type DeepPointerTypes struct {
	PPointerField ***bool
}

type NilTypes struct {
	PBoolField   *bool
	PStringField *string
	PMapField    *map[string]string
}

func (dpt *DeepPointerTypes) init() {
	pv0 := true
	pv1 := &pv0
	pv2 := &pv1
	dpt.PPointerField = &pv2
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

func TestStructPointer(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := PointerTypes{}
	out.init()
	if err := WriteStruct(buf, out); err != nil {
		t.Fatal("WriteStruct pointer failed", err)
	}
	in := PointerTypes{}
	if err := ReadStruct(buf, &in); err != nil {
		t.Fatal("ReadStruct pointer failed", err)
	}

	if !reflect.DeepEqual(PointerTypes(in), PointerTypes(out)) {
		t.Fatalf("Read/Write struct pointer missmatch: in\n%v, out:\n%v\n", in, out)
	}
}

func TestStructDeepPointer(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := DeepPointerTypes{}
	out.init()
	if err := WriteStruct(buf, out); err != nil {
		t.Fatal("WriteStruct deep pointer failed", err)
	}
	in := DeepPointerTypes{}
	if err := ReadStruct(buf, &in); err != nil {
		t.Fatal("ReadStruct deep pointer failed", err)
	}
	if !reflect.DeepEqual(DeepPointerTypes(in), DeepPointerTypes(out)) {
		t.Fatalf("Read/Write struct deep pointer missmatch:\nout:\n%#v\nin:\n%#v\n", out, in)
	}
}

func TestStructNilPointer(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	out := NilTypes{}
	if err := WriteStruct(buf, out); err != nil {
		t.Fatal("WriteStruct nil pointer failed", err)
	}
	in := NilTypes{}
	if err := ReadStruct(buf, &in); err != nil {
		t.Fatal("ReadStruct nil pointer failed", err)
	}

	if *in.PBoolField != false {
		t.Fatal("r/w mismatch")
	}
	if *in.PStringField != "" {
		t.Fatal("r/w mismatch")
	}
}

func BenchmarkReadBool(b *testing.B) {
	b.StopTimer()
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		WriteBool(buf, true)
	}
	out := false
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ReadBool(buf, &out)
	}
}

func BenchmarkWriteBool(b *testing.B) {
	b.StopTimer()
	buf := bytes.NewBuffer(nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		WriteBool(buf, true)
	}
}

func BenchmarkReadNumber(b *testing.B) {
	b.StopTimer()
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		WriteNumber(buf, i)
	}
	out := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ReadNumber(buf, &out)
	}
}

func BenchmarkWriteNumber(b *testing.B) {
	b.StopTimer()
	buf := bytes.NewBuffer(nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		WriteNumber(buf, i)
	}
}

func BenchmarkReadString(b *testing.B) {
	b.StopTimer()
	in := "0123456789"
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		WriteString(buf, in)
	}
	out := ""
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ReadString(buf, &out)
	}
}

func BenchmarkWriteString(b *testing.B) {
	b.StopTimer()
	in := "0123456789"
	buf := bytes.NewBuffer(nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		WriteString(buf, in)
	}
}

func BenchmarkReadArray(b *testing.B) {
	b.StopTimer()
	in := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		WriteArray(buf, in)
	}
	var out [10]byte
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ReadArray(buf, &out)
	}
}

func BenchmarkWriteArray(b *testing.B) {
	b.StopTimer()
	in := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	buf := bytes.NewBuffer(nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		WriteArray(buf, in)
	}
}

func BenchmarkReadSlice(b *testing.B) {
	b.StopTimer()
	in := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		WriteSlice(buf, in)
	}
	var out []byte
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ReadSlice(buf, &out)
	}
}

func BenchmarkWriteSlice(b *testing.B) {
	b.StopTimer()
	in := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	buf := bytes.NewBuffer(nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		WriteSlice(buf, in)
	}
}

func BenchmarkReadMap(b *testing.B) {
	b.StopTimer()
	in := map[string]int{"0123456789": 1}
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		WriteMap(buf, in)
	}
	var out map[string]int
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ReadMap(buf, &out)
	}
}

func BenchmarkWriteMap(b *testing.B) {
	b.StopTimer()
	in := map[string]int{"0123456789": 1}
	buf := bytes.NewBuffer(nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		WriteMap(buf, in)
	}
}

func BenchmarkReadStruct(b *testing.B) {
	b.StopTimer()
	type test struct {
		String string
		Num    int
	}
	in := test{"0123456789", 1}
	buf := bytes.NewBuffer(nil)
	for i := 0; i < b.N; i++ {
		WriteStruct(buf, in)
	}
	var out test
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ReadStruct(buf, &out)
	}
}

func BenchmarkWriteStruct(b *testing.B) {
	b.StopTimer()
	type test struct {
		String string
		Num    int
	}
	in := test{"0123456789", 1}
	buf := bytes.NewBuffer(nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		WriteStruct(buf, in)
	}
}
