package context

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"reflect"
	"testing"
)

// #include "sass_context.h"
import "C"

type SassValue interface{}

func unmarshal(arg UnionSassValue, v interface{}) error {
	f := reflect.ValueOf(v).Elem()
	switch {
	default:
		return errors.New("Unsupported SASS Value")
	case bool(C.sass_value_is_error(arg)):
		return errors.New("FUCK")
	case bool(C.sass_value_is_null(arg)):
		f.Set(reflect.ValueOf("<nil>"))
		return nil
	case bool(C.sass_value_is_number(arg)):
		i := C.sass_number_get_value(arg)
		// Always cast to float64 and let the passed interface
		// decide the number precision.
		if f.Kind() == reflect.Int {
			sl := fmt.Sprintf("Can not cast %v to type reflect.Float64", f.Kind())
			return errors.New(sl)
		}
		vv := float64(i)
		f.Set(reflect.ValueOf(vv))

		// Is it necessary to check integer precision?
		// switch t := f.Kind(); t {
		// default:
		// 	log.Printf("fail: %s %v\n", t, t)
		// 	f.Set(reflect.ValueOf(vv))
		// case reflect.Int:
		// 	f.SetInt(int64(vv))
		// case reflect.Float32:
		// 	f.SetFloat(vv)
		// }
	case bool(C.sass_value_is_string(arg)):
		c := C.sass_string_get_value(arg)
		gc := C.GoString(c)
		if !f.CanSet() {
			return errors.New("Can not set string")
		}

		switch t := f.Kind(); t {
		default:
			log.Fatalf("unknown type %v", t)
		case reflect.String:
			f.SetString(gc)
		case reflect.Interface:
			f.Set(reflect.ValueOf(gc))
		}
	case bool(C.sass_value_is_boolean(arg)):
		b := bool(C.sass_boolean_get_value(arg))
		f.Set(reflect.ValueOf(b))
	case bool(C.sass_value_is_color(arg)):
		col := color.RGBA{
			R: uint8(C.sass_color_get_r(arg)),
			G: uint8(C.sass_color_get_g(arg)),
			B: uint8(C.sass_color_get_b(arg)),
			A: uint8(C.sass_color_get_a(arg)),
		}
		_ = col
		// return col
		f.Set(reflect.ValueOf(col))
	case bool(C.sass_value_is_list(arg)):
		l := make([]SassValue, C.sass_list_get_length(arg))
		for i := range l {
			err := unmarshal(C.sass_list_get_value(arg, C.size_t(i)), &l[i])
			if err != nil {
				return err
			}
		}
		fl := reflect.ValueOf(l)
		f.Set(fl)
	case bool(C.sass_value_is_map(arg)):
		len := int(C.sass_map_get_length(arg))
		m := make(map[SassValue]SassValue, len)
		for i := 0; i < len; i++ {
			//m[Decode(C.sass_map_get_key(arg, C.size_t(i)))] =
			//	Decode(C.sass_map_get_value(arg, C.size_t(i)))
		}
		_ = m
		// return m
	case bool(C.sass_value_is_error(arg)):
		// return C.GoString(C.sass_error_get_message(arg))
	}
	return nil
}

// Decode converts Sass Value to Go compatible data types.
func Unmarshal(arg UnionSassValue, v interface{}) error {
	return unmarshal(arg, v)
}

func Marshal(v interface{}) UnionSassValue {
	return makevalue(v)
}

// make is needed to create types for use by test
func makevalue(v interface{}) UnionSassValue {
	f := reflect.ValueOf(v)
	switch f.Kind() {
	default:
		return C.sass_make_null()
	case reflect.Float32, reflect.Float64:
		switch f.Kind() {
		default:
			return C.sass_make_number(C.double(0), C.CString("wtfisthis"))
		case reflect.Float32:
			return C.sass_make_number(C.double(v.(float32)), C.CString("wtfisthis"))
		case reflect.Float64:
			return C.sass_make_number(C.double(v.(float64)), C.CString("wtfisthis"))
		}
	case reflect.Int:
		return C.sass_make_number(C.double(v.(int)), C.CString("wtfisthis"))
	case reflect.String:
		return C.sass_make_string(C.CString(v.(string)))
	case reflect.Slice:
		// Initialize the list
		l := C.sass_make_list(C.size_t(f.Len()), C.SASS_COMMA)
		for i := 0; i < f.Len(); i++ {
			t := makevalue(f.Index(i).Interface())
			C.sass_list_set_value(l, C.size_t(i), t)
		}
		return l
	}
}

// Can't import C in the test package, so this is how to test cgo code
func testUnmarshalUnknown(t *testing.T) {
	// Test for nil (no value, pointer, or empty error)
	var unk UnionSassValue
	x := Marshal(unk)
	var v interface{}
	_ = Unmarshal(x, &v)
	if v != "<nil>" {
		t.Error("non-nil returned")
	}

	// Need a test for non-supported type
}

func testMarshalNumber(t *testing.T) {
	num := float64(24)
	e := C.double(num)
	x := Marshal(num)

	if d := C.sass_number_get_value(x); d != e {
		t.Errorf("got: %v wanted: %v", d, e)
	}
}

func testMarshalList(t *testing.T) {

	lst := []SassValue{1, 2, 3, 4}
	e := C.sass_make_list(C.size_t(len(lst)), C.SASS_COMMA)

	for i := range lst {
		C.sass_list_set_value(e, C.size_t(i), Marshal(lst[i]))
	}

	x := Marshal(lst)
	if C.sass_list_get_length(x) != C.sass_list_get_length(e) {
		t.Error("list length mismatch")
	}

	for i := range lst {
		v1, v2 :=
			C.sass_list_get_value(x, C.size_t(i)),
			C.sass_list_get_value(e, C.size_t(i))
		f1, f2 := C.sass_number_get_value(v1),
			C.sass_number_get_value(v2)
		if f1 != f2 {
			t.Errorf("wanted: %d got: %d", v2, v1)
		}
	}
}

func testMarshalListInterface(t *testing.T) {
	lst := []SassValue{"a", "b", 3, 4}

	e := C.sass_make_list(C.size_t(len(lst)), C.SASS_COMMA)

	for i := range lst {
		C.sass_list_set_value(e, C.size_t(i), Marshal(lst[i]))
	}

	x := Marshal(lst)
	if C.sass_list_get_length(x) != C.sass_list_get_length(e) {
		t.Error("list length mismatch")
	}

	for i := range lst {
		v1, v2 :=
			C.sass_list_get_value(x, C.size_t(i)),
			C.sass_list_get_value(e, C.size_t(i))
		var f1, f2 SassValue
		switch {
		case bool(C.sass_value_is_number(v1)):
			f1, f2 = C.sass_number_get_value(v1),
				C.sass_number_get_value(v2)
		case bool(C.sass_value_is_string(v1)):
			f1, f2 = C.GoString(C.sass_string_get_value(v1)),
				C.GoString(C.sass_string_get_value(v2))

		}
		if f1 != f2 {
			t.Errorf("wanted: %v got: %v", f2, f1)
		}
	}
}
