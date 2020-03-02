package flagbuilder

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var ErrorInvalidType = fmt.Errorf("v must be a pointer to a struct")

type ErrorDefaultValue struct {
	FieldName string
	Value     string
	Err       error
}

func (err ErrorDefaultValue) Error() string {
	return fmt.Sprintf("%v: cannot assign default value %q: %v",
		err.FieldName, err.Value, err.Err)
}
func (err ErrorDefaultValue) Unwrap() error {
	return err.Err
}

func Build(flg FlagSet, v interface{}) error {
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
			flg.Var(p, tag.Name, tag.Usage)
		}
	}

	return nil
}

type FlagSet interface {
	BoolVar(p *bool, name string, value bool, usage string)
	DurationVar(p *time.Duration, name string, value time.Duration, usage string)
	Float64Var(p *float64, name string, value float64, usage string)
	Int64Var(p *int64, name string, value int64, usage string)
	IntVar(p *int, name string, value int, usage string)
	StringVar(p *string, name string, value string, usage string)
	Uint64Var(p *uint64, name string, value uint64, usage string)
	UintVar(p *uint, name string, value uint, usage string)
	Var(value flag.Value, name string, usage string)
}

type _ struct {
	X int `flag:"-X;"`
}
type flagTag struct {
	Name  string
	Value string
	Usage string

	Ignored bool
}

func newFlagTag(tag string) (fTag flagTag) {
	if len(tag) == 0 {
		return
	}
	args := strings.Split(tag, ";")
	fTag.Ignored = args[0] == "-" // Ignore this field
	if fTag.Ignored {
		return
	}
	fTag.Name = strings.TrimLeft(args[0], "-")
	if len(args) == 1 {
		return
	}
	fTag.Value = args[1]
	if len(args) == 2 {
		return
	}
	fTag.Usage = args[2]
	return
}

func kebabCase(name string) string {
	var kebab string
	for _, r := range name {
		if unicode.IsUpper(r) {
			if len(kebab) > 0 {
				kebab += "-"
			}
			r = unicode.ToLower(r)
		}
		kebab += string(r)
	}
	return kebab
}
