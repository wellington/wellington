package context

import "image/color"

/*

#include "sass_context.h"
*/
import "C"

type SassValue interface{}

// Decode converts Sass Value to Go compatible data types.
func Decode(arg *C.union_Sass_Value) SassValue {
	switch {
	case bool(C.sass_value_is_null(arg)):
		return nil
	case bool(C.sass_value_is_number(arg)):
		return int(C.sass_number_get_value(arg))
	case bool(C.sass_value_is_string(arg)):
		c := C.sass_string_get_value(arg)
		return C.GoString(c)
	case bool(C.sass_value_is_boolean(arg)):
		return bool(C.sass_boolean_get_value(arg))
	case bool(C.sass_value_is_color(arg)):
		col := color.RGBA{
			R: uint8(C.sass_color_get_r(arg)),
			G: uint8(C.sass_color_get_g(arg)),
			B: uint8(C.sass_color_get_b(arg)),
			A: uint8(C.sass_color_get_a(arg)),
		}
		return col
	case bool(C.sass_value_is_list(arg)):
		l := make([]SassValue, C.sass_list_get_length(arg))
		for i := range l {
			l[i] = Decode(C.sass_list_get_value(arg, C.size_t(i)))
		}
		return l
	case bool(C.sass_value_is_map(arg)):
		len := int(C.sass_map_get_length(arg))
		m := make(map[SassValue]SassValue, len)
		for i := 0; i < len; i++ {
			m[Decode(C.sass_map_get_key(arg, C.size_t(i)))] =
				Decode(C.sass_map_get_value(arg, C.size_t(i)))
		}
		return m
	case bool(C.sass_value_is_error(arg)):
		return C.GoString(C.sass_error_get_message(arg))
	}
	return nil
}
