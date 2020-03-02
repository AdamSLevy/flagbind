package flagbuilder

import (
	"flag"
	"reflect"
	"strconv"
	"time"

	"github.com/spf13/pflag"
)

func Bind(flg FlagSet, v interface{}) error {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		return ErrorInvalidType
	}
	val := reflect.Indirect(ptr)
	if val.Kind() != reflect.Struct {
		return ErrorInvalidType
	}

	valT := val.Type()
	// loop through all fields
	for i := 0; i < val.NumField(); i++ {
		fieldT := valT.Field(i)
		if fieldT.PkgPath != "" {
			// unexported field
			continue
		}
		tag := newFlagTag(fieldT.Tag.Get("flag"))
		if tag.Ignored {
			continue
		}
		if tag.Name == "" {
			// no explicit name given
			tag.Name = kebabCase(fieldT.Name)
		}
		fieldV := val.Field(i)
		if fieldT.Type.Kind() != reflect.Ptr {
			// The field is not a ptr, so get a ptr to it.
			fieldV = fieldV.Addr()
		} else if fieldV.IsNil() {
			// We have a nil ptr, so allocate it.
			fieldV.Set(reflect.New(fieldT.Type.Elem()))
		}
		var err error
		switch p := fieldV.Interface().(type) {
		case *bool:
			val := *p
			if tag.Value != "" {
				val, err = strconv.ParseBool(tag.Value)
				if err != nil {
					return ErrorDefaultValue{
						fieldT.Name, tag.Value, err}
				}
			}
			flg.BoolVar(p, tag.Name, val, tag.Usage)
		case *time.Duration:
			val := *p
			if tag.Value != "" {
				val, err = time.ParseDuration(tag.Value)
				if err != nil {
					return ErrorDefaultValue{
						fieldT.Name, tag.Value, err}
				}
			}
			flg.DurationVar(p, tag.Name, val, tag.Usage)
		case *int:
			val := *p
			if tag.Value != "" {
				val64, err := strconv.ParseInt(tag.Value, 10,
					strconv.IntSize)
				if err != nil {
					return ErrorDefaultValue{
						fieldT.Name, tag.Value, err}
				}
				val = int(val64)
			}
			flg.IntVar(p, tag.Name, val, tag.Usage)
		case *uint:
			val := *p
			if tag.Value != "" {
				val64, err := strconv.ParseUint(tag.Value, 10,
					strconv.IntSize)
				if err != nil {
					return ErrorDefaultValue{
						fieldT.Name, tag.Value, err}
				}
				val = uint(val64)
			}
			flg.UintVar(p, tag.Name, val, tag.Usage)
		case *int64:
			val := *p
			if tag.Value != "" {
				val, err = strconv.ParseInt(tag.Value, 10, 64)
				if err != nil {
					return ErrorDefaultValue{
						fieldT.Name, tag.Value, err}
				}
			}
			flg.Int64Var(p, tag.Name, val, tag.Usage)
		case *uint64:
			val := *p
			if tag.Value != "" {
				val, err = strconv.ParseUint(tag.Value, 10, 64)
				if err != nil {
					return ErrorDefaultValue{
						fieldT.Name, tag.Value, err}
				}
			}
			flg.Uint64Var(p, tag.Name, val, tag.Usage)
		case *float64:
			val := *p
			if tag.Value != "" {
				val, err = strconv.ParseFloat(tag.Value, 64)
				if err != nil {
					return ErrorDefaultValue{
						fieldT.Name, tag.Value, err}
				}
			}
			flg.Float64Var(p, tag.Name, val, tag.Usage)
		case *string:
			val := *p
			flg.StringVar(p, tag.Name, val, tag.Usage)
		case flag.Value:
			if tag.Value != "" {
				if err := p.Set(tag.Value); err != nil {
					return ErrorDefaultValue{
						fieldT.Name, tag.Value, err}
				}
			}
			switch flg := flg.(type) {
			case STDFlagSet:
				flg.Var(p, tag.Name, tag.Usage)
			case PFlagSet:
				pp, ok := p.(pflag.Value)
				if !ok {
					pp = pflagValue{p, fieldT.Type.Name()}
				}
				flg.Var(pp, tag.Name, tag.Usage)
			}
		}
	}

	return nil
}
