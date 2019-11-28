// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package binaryex implements functions supplement to binary/encoding package.
// It is designed for ease of use before speed.
//
// It supports binary marshaling of all go types, excluding chans, funcs and
// unsafePointers.
//
// Ints and Uints of any size are encoded as VarInts, floats and complex
// numbers using binary encoding in LittleEndian order, and strings, arrays,
// slices and maps are prefixed by a number (varint) specifying number of
// their elements then written as a LittleEndian byte stream.
//
// Write functions take values or pointers to values. If a Pointer value was
// passed to a Write function it is dereferenced up to the value itself then
// written.
//
// Read functions take pointers to output values only as they have to be
// mutable. If the dereferenced output value is a pointer itself, a new
// value is allocated for it.
//
// If a pointer with a a nil value was Written, when Read, the pointer will
// have its' value allocated (not nil) and value set to zero of that type.
//
// All functions can panic if they encounter invalid parameters as most checks
// are ommited for performance reasons.
//
// If a value supports encoding.Binary(un)Marshaler it is preferred. Watch out
// for infinite loops if calling Read, ReadReflect, Write or WriteReflect from
// a BinaryMarshaler or BinaryUnmarshaler implementor.
//
// If an unsupported value is encountered functions will error.
package binaryex

import (
	"encoding"
	"encoding/binary"
	"io"
	"reflect"
	"strings"
)

// BinaryExError is the base error type of binaryex package.
type BinaryExError struct {
	msg string
}

// Error satisfies the Error interface.
func (bxe BinaryExError) Error() string {
	return "binaryex: " + bxe.msg
}

var (
	// ErrUnsupportedValue is returned when an unsupported value is encountered.
	ErrUnsupportedValue = BinaryExError{"unsupported value"}

	// ErrUnadressableValue is returned when a non-pointer value is passed to a Read* function.
	ErrUnadressableValue = BinaryExError{"unadressable value"}

	// ErrUnexpected is returned when an unexpected value is read.
	ErrUnexpected = BinaryExError{"unexpected value"}
)

// readByteWrapper wraps an io.Reader and implements a ReadByte method.
// This is needed for string, slice and array length prefixes stored as VarInts.
type readByteWrapper struct {
	io.Reader
	p [1]byte
}

// ReadByte implements the ReadByte method.
func (rbw *readByteWrapper) ReadByte() (b byte, err error) {

	if _, err = rbw.Read(rbw.p[:]); err != nil {
		return
	}
	return rbw.p[0], nil
}

// wrapReader wraps an io.Reader in a io.ByteReader implementor.
func wrapReader(r io.Reader) *readByteWrapper {
	return &readByteWrapper{r, [1]byte{0}}
}

// WriteReflect writes a reflect value v to writer w or returns an error
// if one occured.
func WriteReflect(w io.Writer, v reflect.Value) (err error) {
	// Dereference pointers down to value.
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// Write 0 for nil values.
	if !v.IsValid() {
		return WriteNumber(w, 0)
	}
	// Try BinaryMarshaler.
	if bm, ok := v.Interface().(encoding.BinaryMarshaler); ok {
		p, e := bm.MarshalBinary()
		if e != nil {
			return err
		}
		return Write(w, p)
	}
	// Write value.
	switch v.Kind() {
	case reflect.Bool:
		err = WriteBoolReflect(w, v)
	case reflect.String:
		err = WriteStringReflect(w, v)
	case reflect.Array:
		err = WriteArrayReflect(w, v)
	case reflect.Slice:
		err = WriteSliceReflect(w, v)
	case reflect.Map:
		err = WriteMapReflect(w, v)
	case reflect.Struct:
		err = WriteStructReflect(w, v)
	default:
		err = WriteNumberReflect(w, v)
	}
	return
}

// Write writes value val to writer w or returns an error if one occured.
func Write(w io.Writer, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return WriteReflect(w, v)
}

// ReadReflect reads a value from reader r and puts it into v or returns an
// error if one occured.
func ReadReflect(r io.Reader, v reflect.Value) (err error) {
	// Must be addressable.
	if !v.CanAddr() {
		return ErrUnadressableValue
	}
	// Alloc a new concrete value,
	// or a pointer to value if v is a pointer.
	ptr := false
	var pv = reflect.Value{}
	if v.Kind() == reflect.Ptr {
		pv = reflect.New(v.Type().Elem())
		ptr = true
	} else {
		pv = reflect.New(v.Type())
	}
	nv := reflect.Indirect(pv)
	// If a pointer, read the pointer.
	if nv.Kind() == reflect.Ptr {
		if err := ReadReflect(r, nv); err != nil {
			return err
		}
		// Set value.
		if ptr {
			v.Set(pv)
		} else {
			v.Set(nv)
		}
		return
	}
	// Try BinaryMarshaler.
	if nv.IsValid() {
		if bu, ok := nv.Interface().(encoding.BinaryUnmarshaler); ok {
			b := []byte{}
			if err = ReadSlice(r, &b); err != nil {
				return
			}
			return bu.UnmarshalBinary(b)
		}
	}
	// Read value to temp.
	switch pv.Elem().Type().Kind() {
	case reflect.Bool:
		err = ReadBoolReflect(r, nv)
	case reflect.String:
		err = ReadStringReflect(r, nv)
	case reflect.Array:
		err = ReadArrayReflect(r, nv)
	case reflect.Slice:
		err = ReadSliceReflect(r, nv)
	case reflect.Map:
		err = ReadMapReflect(r, nv)
	case reflect.Struct:
		err = ReadStructReflect(r, nv)
	default:
		err = ReadNumberReflect(r, nv)
	}
	if err != nil {
		return err
	}
	// Set value.
	if ptr {
		v.Set(pv)
	} else {
		v.Set(nv)
	}
	return
}

// Read reads a value from r and puts it into val or returns an error
// if one occured.
func Read(r io.Reader, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return ReadReflect(r, v)
}

// WriteBoolReflect writes a bool reflect value v to writer w or returns an
// error if one occured.
func WriteBoolReflect(w io.Writer, v reflect.Value) (err error) {
	if v.Bool() {
		_, err = w.Write([]byte{1})
	} else {
		_, err = w.Write([]byte{0})
	}
	return
}

// WriteBool writes bool value val to writer w or returns an error if one
// occured.
func WriteBool(w io.Writer, b bool) error {
	v := reflect.Indirect(reflect.ValueOf(b))
	return WriteBoolReflect(w, v)
}

// ReadBoolReflect reads a bool value from reader r and puts it into v or
// returns an error if one occured.
func ReadBoolReflect(r io.Reader, v reflect.Value) (err error) {

	if !v.CanAddr() {
		return ErrUnadressableValue
	}

	var p [1]byte
	if _, err = r.Read(p[:]); err != nil {
		return
	}

	if p[0] == 0 {
		v.SetBool(false)
		return
	}
	if p[0] == 1 {
		v.SetBool(true)
		return
	}
	return ErrUnexpected
}

// ReadBool reads a bool value from r and puts it into val or returns an error
// if one occured.
func ReadBool(r io.Reader, b interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(b))
	return ReadBoolReflect(r, v)
}

// WriteNumberReflect writes a number reflect value v to writer w or returns an
// error if one occured.
func WriteNumberReflect(w io.Writer, v reflect.Value) (err error) {

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, v.Int())
		_, err = w.Write(buf[:n])
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(buf, v.Uint())
		_, err = w.Write(buf[:n])
	case reflect.Float32, reflect.Float64:
		err = binary.Write(w, binary.LittleEndian, v.Float())
	case reflect.Complex64, reflect.Complex128:
		err = binary.Write(w, binary.LittleEndian, v.Complex())
	default:
		err = ErrUnsupportedValue
	}
	return
}

// WriteNumber writes number value val to writer w or returns an error if one
// occured.
func WriteNumber(w io.Writer, n interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(n))
	return WriteNumberReflect(w, v)
}

// ReadNumberReflect reads a number value from reader r and puts it into v or
// returns an error if one occured.
func ReadNumberReflect(r io.Reader, v reflect.Value) (err error) {

	if !v.CanAddr() {
		return ErrUnadressableValue
	}

	rw := wrapReader(r)

	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, e := binary.ReadVarint(rw)
		if e != nil {
			return e
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		n, e := binary.ReadUvarint(rw)
		if e != nil {
			return e
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		var n float64
		if err = binary.Read(rw, binary.LittleEndian, &n); err != nil {
			return
		}
		v.SetFloat(n)
	case reflect.Complex64, reflect.Complex128:
		var n complex128
		if err = binary.Read(rw, binary.LittleEndian, &n); err != nil {
			return
		}
		v.SetComplex(n)
	default:
		err = ErrUnsupportedValue
	}
	return
}

// ReadNumber reads a number value from r and puts it into val or returns an
// error if one occured.
func ReadNumber(r io.Reader, n interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(n))
	return ReadNumberReflect(r, v)
}

// WriteStringReflect writes a reflect value v to writer w or returns an error
// if one occured.
func WriteStringReflect(w io.Writer, v reflect.Value) (err error) {
	if err = WriteNumber(w, v.Len()); err != nil {
		return err
	}
	_, err = w.Write([]byte(v.String()))
	return
}

// WriteString writes string value val to writer w or returns an error if one
// occured.
func WriteString(w io.Writer, s string) error {
	v := reflect.Indirect(reflect.ValueOf(s))
	return WriteStringReflect(w, v)
}

// ReadSgtringReflect reads a string value from reader r and puts it into v or
// returns an error if one occured.
func ReadStringReflect(r io.Reader, v reflect.Value) (err error) {

	if !v.CanAddr() {
		return ErrUnadressableValue
	}

	l := 0
	if err = ReadNumber(r, &l); err != nil {
		return err
	}
	if l < 0 {
		return ErrUnexpected
	}
	if l == 0 {
		v.Set(reflect.Zero(v.Type()))
		return
	}
	buf := make([]byte, l)
	_, err = r.Read(buf)
	if err != nil {
		return err
	}
	v.SetString(string(buf))
	return
}

// ReadString reads a value from r and puts it into val or returns an error if
// one occured.
func ReadString(r io.Reader, s interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(s))
	return ReadStringReflect(r, v)
}

// WriteArrayReflect writes an array reflect value v to writer w or returns an
// error if one occured.
func WriteArrayReflect(w io.Writer, v reflect.Value) (err error) {
	for i := 0; i < v.Type().Len(); i++ {
		if err = WriteReflect(w, v.Index(i)); err != nil {
			break
		}
	}
	return
}

// WriteArray writes array value val to writer w or returns an error if one
// occured.
func WriteArray(w io.Writer, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return WriteArrayReflect(w, v)
}

// ReadArrayReflect reads an array value from reader r and puts it into v or
// returns an error if one occured.
func ReadArrayReflect(r io.Reader, v reflect.Value) (err error) {

	if !v.CanAddr() {
		return ErrUnadressableValue
	}

	for i := 0; i < v.Type().Len(); i++ {
		if err = ReadReflect(r, v.Index(i)); err != nil {
			break
		}
	}
	return
}

// ReadArray reads an array value from r and puts it into val or returns an
// error if one occured.
func ReadArray(r io.Reader, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return ReadArrayReflect(r, v)
}

// WriteSliceReflect writes a slice reflect value v to writer w or returns an
// error if one occured.
func WriteSliceReflect(w io.Writer, v reflect.Value) (err error) {
	if err = WriteNumber(w, v.Len()); err != nil {
		return
	}
	for i := 0; i < v.Len(); i++ {
		if err = WriteReflect(w, v.Index(i)); err != nil {
			break
		}
	}
	return
}

// WriteSlice writes slice value val to writer w or returns an error if one
// occured.
func WriteSlice(w io.Writer, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return WriteSliceReflect(w, v)
}

// ReadSliceReflect reads a slice value from reader r and puts it into v or
// returns an error if one occured.
func ReadSliceReflect(r io.Reader, v reflect.Value) (err error) {

	if !v.CanAddr() {
		return ErrUnadressableValue
	}

	l := 0
	if err = ReadNumber(r, &l); err != nil {
		return
	}
	if l < 0 {
		return ErrUnexpected
	}
	v.Set(reflect.MakeSlice(v.Type(), l, l))
	for i := 0; i < l; i++ {
		if err = ReadReflect(r, v.Index(i)); err != nil {
			break
		}
	}
	return
}

// ReadSlice reads a slice value from r and puts it into val or returns an error
// if one occured.
func ReadSlice(r io.Reader, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return ReadSliceReflect(r, v)
}

// WriteMapReflect writes a map reflect value v to writer w or returns an error
// if one occured.
func WriteMapReflect(w io.Writer, v reflect.Value) (err error) {

	if err = WriteNumber(w, v.Len()); err != nil {
		return
	}
	for _, mk := range v.MapKeys() {
		mv := v.MapIndex(mk)
		if err = WriteReflect(w, mk); err != nil {
			break
		}
		if err = WriteReflect(w, mv); err != nil {
			break
		}
	}
	return
}

// WriteMap writes map value val to writer w or returns an error if one occured.
func WriteMap(w io.Writer, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return WriteMapReflect(w, v)
}

// ReadMapReflect reads a map value from reader r and puts it into v or returns
// an error if one occured.
func ReadMapReflect(r io.Reader, v reflect.Value) (err error) {

	if !v.CanAddr() {
		return ErrUnadressableValue
	}

	l := 0
	if err = ReadNumber(r, &l); err != nil {
		return
	}
	if l < 0 {
		return ErrUnexpected
	}
	kt := v.Type().Key()
	vt := v.Type().Elem()
	v.Set(reflect.MakeMap(v.Type()))
	for i := 0; i < l; i++ {
		kv := reflect.Indirect(reflect.New(kt))
		if err = ReadReflect(r, kv); err != nil {
			break
		}
		vv := reflect.Indirect(reflect.New(vt))
		if err = ReadReflect(r, vv); err != nil {
			break
		}
		v.SetMapIndex(kv, vv)
	}
	return
}

// ReadMap reads a map value from r and puts it into val or returns an error if
// one occured.
func ReadMap(r io.Reader, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return ReadMapReflect(r, v)
}

// WriteStructReflect writes a struct reflect value v to writer w or returns an
// error if one occured.
func WriteStructReflect(w io.Writer, v reflect.Value) (err error) {

	for i := 0; i < v.NumField(); i++ {
		fname := v.Type().Field(i).Name
		if fname == "_" {
			continue
		}
		if fname[0] == strings.ToLower(fname)[0] {
			continue
		}
		if err = WriteReflect(w, v.Field(i)); err != nil {
			break
		}
	}
	return
}

// WriteStruct writes struct value val to writer w or returns an error if one
// occured.
func WriteStruct(w io.Writer, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return WriteStructReflect(w, v)
}

// ReadStructReflect reads a struct value from reader r and puts it into v or
// returns an error if one occured.
func ReadStructReflect(r io.Reader, v reflect.Value) (err error) {

	if !v.CanAddr() {
		return ErrUnadressableValue
	}

	for i := 0; i < v.NumField(); i++ {
		fname := v.Type().Field(i).Name
		if fname == "_" {
			continue
		}
		if fname[0] == strings.ToLower(fname)[0] {
			continue
		}
		if !v.Field(i).CanSet() {
			continue
		}
		if err = ReadReflect(r, v.Field(i)); err != nil {
			break
		}
	}
	return
}

// ReadStruct reads a struct value from r and puts it into val or returns an
// error if one occured.
func ReadStruct(r io.Reader, val interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(val))
	return ReadStructReflect(r, v)
}

// TODO
func WriteChanReflect(w io.Writer, v reflect.Value) (err error) {
	return nil
}

// TODO
func WriteChan(w io.Writer, val interface{}) error {
	return nil
}

// TODO
func ReadChanReflect(r io.Reader, v reflect.Value) (err error) {
	/*
		typ := v.Type()
		ctyp := reflect.ChanOf(typ.ChanDir(), typ)
		_ = reflect.MakeChan(ctyp, 0)
	*/
	return nil
}

// TODO
func ReadChan(r io.Reader, val interface{}) error {
	return nil
}
