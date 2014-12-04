package context

import (
	"errors"
	"fmt"
	//"fmt"
	"image/color"
	"reflect"
)

// #include "sass_context.h"
import "C"

type UnionSassValue *C.union_Sass_Value
type IntSassComma C.int

func unmarshal(arg UnionSassValue, v interface{}) error {
	//Get the underlying value of v and its kind
	f := reflect.ValueOf(v)
	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}

	t := f.Kind()

	//If passed an interface allow the SassValue to dictate the resulting type
	if t == reflect.Interface {
		switch {
		default:
			return errors.New("Uncovertable interface value. Specify type desired.")
		case bool(C.sass_value_is_null(arg)):
			f.Set(reflect.ValueOf("<nil>"))
			return nil
		case bool(C.sass_value_is_number(arg)):
			t = reflect.Float64
		case bool(C.sass_value_is_string(arg)):
			t = reflect.String
		case bool(C.sass_value_is_boolean(arg)):
			t = reflect.Bool
		}
	}

	switch t {
	default:
		return errors.New("Unsupported SassValue")
	case reflect.Invalid:
		return errors.New("Invalid SASS Value - Taylor Swift")
	case reflect.Interface:
		//Wild card.  Check nil first and then build the value based on the SassValue

	case reflect.Float64:
		if C.sass_value_is_number(arg) {
			i := C.sass_number_get_value(arg)
			vv := float64(i)
			f.Set(reflect.ValueOf(vv))
		} else {
			return errors.New("Matching SassValue is not a float64")
		}
	case reflect.String:
		if C.sass_value_is_string(arg) {
			c := C.sass_string_get_value(arg)
			gc := C.GoString(c)
			if !f.CanSet() {
				return errors.New("Can not set string")
			}

			switch t := f.Kind(); t {
			case reflect.String:
				f.SetString(gc)
			case reflect.Interface:
				f.Set(reflect.ValueOf(gc))
			}
		} else {
			return errors.New("Matching SassValue is not a string")
		}
	case reflect.Bool:
		if C.sass_value_is_boolean(arg) {
			b := bool(C.sass_boolean_get_value(arg))
			f.Set(reflect.ValueOf(b))
		} else {
			return errors.New("Matching SassValue is not a bool")
		}
	case reflect.Struct:
		//Check for color
		if C.sass_value_is_color(arg) {
			col := color.RGBA{
				R: uint8(C.sass_color_get_r(arg)),
				G: uint8(C.sass_color_get_g(arg)),
				B: uint8(C.sass_color_get_b(arg)),
				A: uint8(C.sass_color_get_a(arg)),
			}
			f.Set(reflect.ValueOf(col))
		} else {
			return errors.New("Matching SassValue is not a color.RGBA")
		}
	case reflect.Slice:
		if C.sass_value_is_list(arg) {
			newv := reflect.MakeSlice(f.Type(), int(C.sass_list_get_length(arg)), int(C.sass_list_get_length(arg)))
			l := make([]interface{}, C.sass_list_get_length(arg))
			for i := range l {
				err := unmarshal(C.sass_list_get_value(arg, C.size_t(i)), &l[i])
				if err != nil {
					return err
				}
				newv.Index(i).Set(reflect.ValueOf(l[i]))
			}
			f.Set(newv)
		} else {
			return errors.New("Matching SassValue is not a list")
		}
	}
	return nil
}

// Decode converts Sass Value to Go compatible data types.
func Unmarshal(arg UnionSassValue, v ...interface{}) error {
	var err error
	if len(v) == 0 {
		return errors.New("Cannot Unmarshal an empty value - Michael Scott")
	}
	if len(v) > 1 {
		if len(v) != int(C.sass_list_get_length(arg)) {
			return errors.New(fmt.Sprintf("Arguments mismatch %d C arguments did not match %d",
				int(C.sass_list_get_length(arg)), len(v)))
		}
		for i := range v {
			err = unmarshal(C.sass_list_get_value(arg, C.size_t(i)), v[i])
			if err != nil {
				return err
			}
		}
		return err
	}
	return unmarshal(arg, v[0])
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
