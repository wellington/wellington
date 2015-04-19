package context

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"
	"strconv"
	"strings"
)

// #include "sass_context.h"
import "C"

type UnionSassValue *C.union_Sass_Value

func unmarshal(arg UnionSassValue, v interface{}) error {

	//Get the underlying value of v and its kind
	f := reflect.ValueOf(v)

	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}

	k := f.Kind()
	t := f.Type()

	if k == reflect.Interface {
		switch {
		default:
			return errors.New("Uncovertable interface value. Specify type desired.")
		case bool(C.sass_value_is_null(arg)):
			f.Set(reflect.ValueOf("<nil>"))
			return nil
		case bool(C.sass_value_is_string(arg)):
			k = reflect.String
		case bool(C.sass_value_is_boolean(arg)):
			k = reflect.Bool
		case bool(C.sass_value_is_color(arg)) ||
			bool(C.sass_value_is_number(arg)):
			k = reflect.Struct
		case bool(C.sass_value_is_list(arg)):
			k = reflect.Slice
			t = reflect.SliceOf(t)
		case bool(C.sass_value_is_error(arg)):
			// This should get implemented as type error
			k = reflect.String
		}
	}

	switch k {
	default:
		return errors.New("Unsupported SassValue")
	case reflect.Invalid:
		return errors.New("Invalid SASS Value - Taylor Swift")
	case reflect.String:
		if C.sass_value_is_string(arg) || C.sass_value_is_error(arg) {
			c := C.sass_string_get_value(arg)
			gc := C.GoString(c)
			//drop quotes
			if t, err := strconv.Unquote(gc); err == nil {
				gc = t
			}
			if strings.HasPrefix(gc, "'") && strings.HasSuffix(gc, "'") {
				gc = gc[1 : len(gc)-1]
			}
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
			return throwMisMatchTypeError(arg, "string")
		}
	case reflect.Bool:
		if C.sass_value_is_boolean(arg) {
			b := bool(C.sass_boolean_get_value(arg))
			f.Set(reflect.ValueOf(b))
		} else {
			return throwMisMatchTypeError(arg, "bool")
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
		} else if C.sass_value_is_number(arg) {
			u, err := getSassNumberUnit(arg)
			if err != nil {
				return err
			}
			sn := SassNumber{
				Value: float64(C.sass_number_get_value(arg)),
				Unit:  u,
			}
			f.Set(reflect.ValueOf(sn))

		} else {
			return throwMisMatchTypeError(arg, "color.RGBA or SassNumber")
		}
	case reflect.Slice:
		if C.sass_value_is_list(arg) {
			newv := reflect.MakeSlice(t, int(C.sass_list_get_length(arg)), int(C.sass_list_get_length(arg)))
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
			return throwMisMatchTypeError(arg, "slice")
		}
	}
	return nil
}

// Decode converts Sass Value to Go compatible data types.
func Unmarshal(arg UnionSassValue, v ...interface{}) error {
	var err error
	if arg == nil {
		return errors.New("I can't work with this. arg UnionSassValue must not be nil. - Unmarshaller")
	} else if len(v) == 0 {
		return errors.New("Cannot Unmarshal an empty value - Michael Scott")
	} else if len(v) > 1 {
		if len(v) < int(C.sass_list_get_length(arg)) { //check for optional arguements that are not passed and pad with nil
			return fmt.Errorf("Arguments mismatch %d C arguments did not match %d",
				int(C.sass_list_get_length(arg)), len(v))
		}
		for i := 0; i < int(C.sass_list_get_length(arg)); i++ {
			err = unmarshal(C.sass_list_get_value(arg, C.size_t(i)), v[i])
			if err != nil {
				return err
			}
		}
		return err
	} else if C.sass_value_is_list(arg) && getKind(v[0]) != reflect.Slice && int(C.sass_list_get_length(arg)) == 1 { //arg is a slice of 1 but we want back a non slice
		return unmarshal(C.sass_list_get_value(arg, C.size_t(0)), v[0])
	} else if C.sass_value_is_list(arg) && getKind(v[0]) == reflect.Slice && C.sass_value_is_list(C.sass_list_get_value(arg, C.size_t(0))) && int(C.sass_list_get_length(arg)) == 1 { //arg is a list of single list and we only want back a list so we need to unwrap
		return unmarshal(C.sass_list_get_value(arg, C.size_t(0)), v[0])
	} else {
		return unmarshal(arg, v[0])
	}
}

func getKind(v interface{}) reflect.Kind {
	f := reflect.ValueOf(v)

	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}

	return f.Kind()
}

func noSassNumberUnit(arg UnionSassValue) bool {
	return C.GoString(C.sass_number_get_unit(arg)) == "" || C.GoString(C.sass_number_get_unit(arg)) == "none"
}

func getSassNumberUnit(arg UnionSassValue) (string, error) {
	u := C.GoString(C.sass_number_get_unit(arg))
	err := error(nil)

	if u == "" || u == "none" {
		err = fmt.Errorf("SassNumber has no units.")
	}

	if _, ok := sassUnitConversions[u]; !ok {
		err = fmt.Errorf("SassNumber units %s are unsupported", u)
	}

	return u, err
}

func Marshal(v interface{}) (UnionSassValue, error) {
	return makevalue(v)
}

// make is needed to create types for use by test
func makevalue(v interface{}) (UnionSassValue, error) {
	f := reflect.ValueOf(v)
	err := error(nil)
	switch f.Kind() {
	default:
		return C.sass_make_null(), err
	case reflect.Bool:
		return C.sass_make_boolean(C.bool(v.(bool))), err
	case reflect.String:
		return C.sass_make_string(C.CString(v.(string))), err
	case reflect.Struct: //only SassNumber and color.RGBA are supported
		if reflect.TypeOf(v).String() == "context.SassNumber" {
			var sn = v.(SassNumber)
			return C.sass_make_number(C.double(sn.Value), C.CString(sn.Unit)), err
		} else if reflect.TypeOf(v).String() == "color.RGBA" {
			var sc = v.(color.RGBA)
			return C.sass_make_color(C.double(sc.R), C.double(sc.G), C.double(sc.B), C.double(sc.A)), err
		} else {
			err = errors.New(fmt.Sprintf("The struct type %s is unsupported for marshalling", reflect.TypeOf(v).String()))
			return C.sass_make_null(), err
		}
	case reflect.Slice:
		// Initialize the list
		l := C.sass_make_list(C.size_t(f.Len()), C.SASS_COMMA)
		for i := 0; i < f.Len(); i++ {
			t, er := makevalue(f.Index(i).Interface())
			if err == nil && er != nil {
				err = er
			}
			C.sass_list_set_value(l, C.size_t(i), t)
		}
		return l, err
	}
}

func throwMisMatchTypeError(arg UnionSassValue, expectedType string) error {
	var intf interface{}
	unmarshal(arg, &intf)
	return fmt.Errorf("Sassvalue is type %s and has value %v but expected %s", reflect.TypeOf(intf), intf, expectedType)
}
