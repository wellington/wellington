package libs

// #include "sass_context.h"
import "C"
import (
	"image/color"
	"reflect"
)

type UnionSassValue *C.union_Sass_Value

func NewUnionSassValue() UnionSassValue {
	return &C.union_Sass_Value{}
}

// types
func MakeNil() UnionSassValue {
	return C.sass_make_null()
}

func MakeBool(b bool) UnionSassValue {
	return C.sass_make_boolean(C.bool(b))
}

func MakeError(s string) UnionSassValue {
	return C.sass_make_error(C.CString(s))
}

func MakeWarning(s string) UnionSassValue {
	return C.sass_make_warning(C.CString(s))
}

func MakeString(s string) UnionSassValue {
	return C.sass_make_string(C.CString(s))
}

// TODO: validate unit
func MakeNumber(f float64, unit string) UnionSassValue {
	return C.sass_make_number(C.double(f), C.CString(unit))
}

func MakeColor(c color.RGBA) UnionSassValue {
	return C.sass_make_color(C.double(c.R), C.double(c.G),
		C.double(c.B), C.double(c.A))
}

func MakeList(len int) UnionSassValue {
	return C.sass_make_list(C.size_t(len), C.SASS_COMMA)
}

func MakeMap(len int) UnionSassValue {
	return C.sass_make_map(C.size_t(len))
}

func Slice(usv UnionSassValue, inf interface{}) {
	if !IsList(usv) {
		panic("sass value is not a list")
	}
	l := Len(usv)
	r := reflect.ValueOf(inf)
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	if !r.CanSet() {
		panic("value is not addressable")
	}
	// if a non-slice is passed, make a slice
	t := r.Type()
	if r.Kind() != reflect.Slice {
		t = reflect.SliceOf(t)
	}
	d := reflect.MakeSlice(t, l, l)
	for i := 0; i < l; i++ {
		sv := Index(usv, i)
		inf := Interface(sv)
		rf := reflect.ValueOf(inf)
		// Special case for nil
		if inf == nil {
			d.Index(i).Set(reflect.ValueOf("<nil>"))
			continue
		}
		d.Index(i).Set(rf)
	}
	r.Set(d)
}

func IsNil(usv UnionSassValue) bool {
	return bool(C.sass_value_is_null(usv))
}

func IsBool(usv UnionSassValue) bool {
	return bool(C.sass_value_is_boolean(usv))
}

func IsString(usv UnionSassValue) bool {
	return bool(C.sass_value_is_string(usv))
}

func IsColor(usv UnionSassValue) bool {
	return bool(C.sass_value_is_color(usv))
}

func IsNumber(usv UnionSassValue) bool {
	return bool(C.sass_value_is_number(usv))
}

func IsList(usv UnionSassValue) bool {
	return bool(C.sass_value_is_list(usv))
}

func IsMap(usv UnionSassValue) bool {
	return bool(C.sass_value_is_map(usv))
}

func IsError(usv UnionSassValue) bool {
	return bool(C.sass_value_is_error(usv))
}

// Interface creates Go types from union sass_value
func Interface(usv UnionSassValue) interface{} {
	switch {
	case IsNil(usv):
		return nil
	case IsBool(usv):
		return Bool(usv)
	case IsString(usv):
		return String(usv)
	case IsColor(usv):
		return Color(usv)
	case IsNumber(usv):
		return Number(usv)
	case IsList(usv):
		fallthrough
		//return List(usv)
	case IsMap(usv):
		fallthrough
		//return Map(usv)
	default:
		return nil
	}
	panic("call of interface not supported on type")
}

func Len(usv UnionSassValue) int {
	switch {
	case IsList(usv):
		return int(C.sass_list_get_length(usv))
	case IsMap(usv):
		return int(C.sass_map_get_length(usv))
	}
	panic("call of len on unknown type")
}

func String(usv UnionSassValue) string {
	c := C.sass_string_get_value(usv)
	gc := C.GoString(c)
	return gc
}

type SassNumber struct {
	Value float64
	Unit  string
}

func (n SassNumber) Float() float64 {
	return n.Value
}

func (n SassNumber) UnitOf() string {
	return n.Unit
}

func Number(usv UnionSassValue) SassNumber {
	return SassNumber{
		Value: Float(usv),
		Unit:  Unit(usv),
	}
}

func Float(usv UnionSassValue) float64 {
	f := C.sass_number_get_value(usv)
	return float64(f)
}

func Unit(usv UnionSassValue) string {
	return C.GoString(C.sass_number_get_unit(usv))
}

func Bool(usv UnionSassValue) bool {
	b := C.sass_boolean_get_value(usv)
	return bool(b)
}

func Color(usv UnionSassValue) color.Color {
	return color.RGBA{
		R: uint8(C.sass_color_get_r(usv)),
		G: uint8(C.sass_color_get_g(usv)),
		B: uint8(C.sass_color_get_b(usv)),
		A: uint8(C.sass_color_get_a(usv)),
	}
}

func Index(usv UnionSassValue, i int) UnionSassValue {
	switch {
	case IsList(usv):
		return C.sass_list_get_value(usv, C.size_t(i))
	default:
		panic("call of index on unknown type")
	}
	return NewUnionSassValue()
}

func SetIndex(usv UnionSassValue, i int, item UnionSassValue) {
	switch {
	case IsList(usv):
		C.sass_list_set_value(usv, C.size_t(i), item)
		return
	default:
	}
	panic("call of setindex on unknown type")
}

func MapKeys(m UnionSassValue) []UnionSassValue {
	res := make([]UnionSassValue, Len(m))
	for i := range res {
		res[i] = C.sass_map_get_key(m, C.size_t(i))
	}
	return res
}

func mapVals(m UnionSassValue) []UnionSassValue {
	res := make([]UnionSassValue, Len(m))
	for i := range res {
		res[i] = C.sass_map_get_value(m, C.size_t(i))
	}
	return res
}

// TODO: support maps
// func MapIndex(m UnionSassValue, key UnionSassValue) UnionSassValue {
// 	keys := MapKeys(m)
// }
