package context

import (
	"fmt"
	"image/color"
	"log"
	"reflect"
)

// #include "sass_context.h"
import "C"

type SassValue interface{}

// make is needed to create types for use by test
func makevalue(v interface{}) *C.union_Sass_Value {
	f := reflect.ValueOf(v)
	switch f.Kind() {
	default:
		log.Printf("% #v\n", v)
		log.Fatalf("Type unsupported: %v", f.Kind())
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

func unmarshal(arg *C.union_Sass_Value, v interface{}) {
	f := reflect.ValueOf(v).Elem()
	switch {
	default:
		fmt.Printf("Unsupported SASS Value\n")
	case bool(C.sass_value_is_null(arg)):

	case bool(C.sass_value_is_number(arg)):
		i := C.sass_number_get_value(arg)
		// Always cast to float64 and let the passed interface
		// decide the number precision.
		flow := float64(i)
		f.Set(reflect.ValueOf(flow))

		break
		// Is it necessary to check integer precision?
		switch t := f.Kind(); t {
		default:
			log.Printf("fail: %s %v\n", t, t)
			f.Set(reflect.ValueOf(flow))
		case reflect.Int:
			f.SetInt(int64(flow))
		case reflect.Float32:
			f.SetFloat(flow)
		}
	case bool(C.sass_value_is_string(arg)):
		c := C.sass_string_get_value(arg)
		gc := C.GoString(c)
		if !f.CanSet() {
			return
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
			unmarshal(C.sass_list_get_value(arg, C.size_t(i)), &l[i])
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
}

// Decode converts Sass Value to Go compatible data types.
func Unmarshal(arg *C.union_Sass_Value, v interface{}) {
	unmarshal(arg, v)
}
