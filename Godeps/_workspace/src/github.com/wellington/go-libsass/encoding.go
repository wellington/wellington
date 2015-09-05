package libsass

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"
	"strconv"
	"strings"

	"github.com/wellington/go-libsass/libs"
)

var (
	ErrSassNumberNoUnit = errors.New("SassNumber has no units")
)

type SassValue struct {
	value libs.UnionSassValue
}

func (sv SassValue) Val() libs.UnionSassValue {
	return sv.value
}

func NewSassValue() SassValue {
	return SassValue{value: libs.NewUnionSassValue()}
}

func unmarshal(arg SassValue, v interface{}) error {
	sv := arg.Val()
	//Get the underlying value of v and its kind
	f := reflect.ValueOf(v)

	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}

	k := f.Kind()
	t := f.Type()

	if k == reflect.Interface {
		switch {
		case libs.IsNil(sv):
			f.Set(reflect.ValueOf("<nil>"))
			return nil
		case libs.IsString(sv):
			k = reflect.String
		case libs.IsBool(sv):
			k = reflect.Bool
		case libs.IsNumber(sv):
			k = reflect.Struct
		case libs.IsList(sv):
			k = reflect.Slice
			t = reflect.SliceOf(t)
		case libs.IsError(sv):
			// This should get implemented as type error
			k = reflect.String
		case libs.IsColor(sv):
			k = reflect.Struct
		default:
			return errors.New("Uncovertable interface value.")
		}
	}

	switch k {
	case reflect.Invalid:
		return errors.New("Invalid SASS Value - Taylor Swift")
	case reflect.String:
		if libs.IsString(sv) || libs.IsError(sv) {
			gc := libs.String(sv)
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
		if libs.IsBool(sv) {
			b := libs.Bool(sv)
			f.Set(reflect.ValueOf(b))
		} else {
			return throwMisMatchTypeError(arg, "bool")
		}
	case reflect.Struct:
		switch {
		case libs.IsColor(sv):
			col := libs.Color(sv)
			f.Set(reflect.ValueOf(col))
		case libs.IsNumber(sv):
			u, err := getSassNumberUnit(arg)
			if err != nil {
				return err
			}
			sn := libs.SassNumber{
				Value: libs.Float(sv),
				Unit:  u,
			}
			f.Set(reflect.ValueOf(sn))
		default:
			return throwMisMatchTypeError(arg, "color.RGBA or SassNumber")
		}
	case reflect.Slice:
		if !libs.IsList(sv) {
			return throwMisMatchTypeError(arg, "slice")
		}
		libs.Slice(arg.Val(), v)
	default:
		return errors.New("Unsupported SassValue")
	}
	return nil
}

// Decode converts Sass Value to Go compatible data types.
func Unmarshal(arg SassValue, v ...interface{}) error {
	var err error
	sv := arg.Val()
	var l int
	if libs.IsList(sv) {
		l = libs.Len(sv)
	}
	if arg.Val() == nil {
		return errors.New("I can't work with this. arg UnionSassValue must not be nil. - Unmarshaller")
	} else if len(v) == 0 {
		return errors.New("Cannot Unmarshal an empty value - Michael Scott")
	} else if len(v) > 1 {
		if len(v) < l { //check for optional arguements that are not passed and pad with nil
			return fmt.Errorf(
				"Arguments mismatch %d C arguments did not match %d",
				l, len(v))
		}
		for i := 0; i < l; i++ {
			val := libs.Index(sv, i)
			err = unmarshal(SassValue{value: val}, v[i])
			if err != nil {
				return err
			}
		}
		return err
	} else if libs.IsList(sv) &&
		getKind(v[0]) != reflect.Slice &&
		l == 1 { //arg is a slice of 1 but we want back a non slice
		val := libs.Index(sv, 0)
		return unmarshal(SassValue{value: val}, v[0])
	} else if libs.IsList(sv) &&
		getKind(v[0]) == reflect.Slice &&
		libs.IsList(libs.Index(sv, 0)) &&
		l == 1 { //arg is a list of single list and we only want back a list so we need to unwrap
		val := libs.Index(sv, 0)
		return unmarshal(SassValue{value: val}, v[0])
		//return unmarshal(C.sass_list_get_value(arg, C.size_t(0)), v[0])
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

func noSassNumberUnit(arg SassValue) bool {
	u := libs.Unit(arg.Val())
	return u == "" || u == "none"
}

func getSassNumberUnit(arg SassValue) (string, error) {
	u := libs.Unit(arg.Val())
	var err error
	if noSassNumberUnit(arg) {
		return u, ErrSassNumberNoUnit
	}
	if _, ok := libs.SassUnitConversions[u]; !ok {
		return u, fmt.Errorf("SassNumber units %s are unsupported", u)
	}
	return u, err
}

func Marshal(v interface{}) (SassValue, error) {
	return makevalue(v)
}

// make is needed to create types for use by test
func makevalue(v interface{}) (SassValue, error) {
	f := reflect.ValueOf(v)
	var err error
	switch f.Kind() {
	default:
		return SassValue{value: libs.MakeNil()}, nil
	case reflect.Bool:
		b := v.(bool)
		return SassValue{value: libs.MakeBool(b)}, nil
	case reflect.String:
		s := v.(string)
		return SassValue{value: libs.MakeString(s)}, nil
	case reflect.Struct: //only SassNumber and color.RGBA are supported
		if sn, ok := v.(libs.SassNumber); ok {
			return SassValue{
				value: libs.MakeNumber(sn.Float(), sn.UnitOf()),
			}, err
		} else if sc, ok := v.(color.RGBA); ok {
			return SassValue{value: libs.MakeColor(sc)}, nil
		} else {
			err = errors.New(fmt.Sprintf("The struct type %s is unsupported for marshalling", reflect.TypeOf(v).String()))
			return SassValue{value: libs.MakeNil()}, err
		}
	case reflect.Slice:
		// Initialize the list
		lst := libs.MakeList(f.Len())
		for i := 0; i < f.Len(); i++ {
			t, er := makevalue(f.Index(i).Interface())
			if err == nil && er != nil {
				err = er
			}
			libs.SetIndex(lst, i, t.Val())
		}
		return SassValue{value: lst}, err
	}
}

func throwMisMatchTypeError(arg SassValue, expectedType string) error {
	var intf interface{}
	unmarshal(arg, &intf)
	svinf := libs.Interface(arg.Val())
	return fmt.Errorf("Invalid Sass type expected: %s got: %T value: %v",
		expectedType, svinf, svinf)
}
