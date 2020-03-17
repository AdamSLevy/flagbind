// Copyright (c) 2020 Adam S Levy
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

// Package flagbind parses the exported fields of a struct and binds them to
// flags in a flag.FlagSet or pflag.FlagSet.
//
// Bind allows for creating flags declaratively right alongside the definition
// of their containing struct. For example, the following stuct could be passed
// to Bind to populate a flag.FlagSet or github.com/spf13/pflag.FlagSet.
//
//	flags := struct {
//		StringFlag string `flag:"flag-name;default value;Usage for string-flag"`
//		Int        int    `flag:"integer;5"`
//
//		// Flag names default to `auto-kebab-case`
//		AutoKebabCase int
//
//              // If pflag is used, -s is be used as the shorthand flag name,
//              // otherwise it is ignored for use with the standard flag package.
//		ShortName bool `flag:"short,s"`
//
//              // Optionally extende the usage tag with subsequent `use` tags
//              // on _ fields.
//              URL string `flag:"url,u;http://www.example.com/;Start usage here"
//              _   struct{} `use:"continue longer usage string for --url below it",
//
//		// Nested and Embedded structs can add a flag name prefix, or not.
//		Nested     StructA
//		NestedFlat StructB           `flag:";;;flatten"`
//		StructA                      // Flat by default
//		StructB    `flag:"embedded"` // Add prefix to nested field flag names.
//
//		// Ignored
//		ExplicitlyIgnored bool `flag:"-"`
//		unexported        bool
//	}{
//		// Default values may also be set directly to override the tag.
//		StringFlag: "override tag default",
//	}
//
//	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
//	flagbind.Bind(fs, &flags)
//	fs.Parse([]string{"--auto-kebab-case"})
//
// Bind works seemlessly with both the standard library flag package and the
// popular github.com/spf13/pflag package.
//
// If pflag is used, for types that implement flag.Value but not pflag.Value,
// Bind wraps them in an adapter so that they can still be used as a
// pflag.Value. The return value of the additional function `Type() string` is
// the type name of the struct field.
//
// Additional options may be set for each flag. See Bind for the full
// documentation details.
package flagbind

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// PrefixSeparator is used to separate a prefix from a flag name.
var PrefixSeparator = "-"

// Bind the exported fields of struct v to new flags in fs.
//
// Bind returns ErrorInvalidFlagSet if fs does not implement STDFlagSet or
// PFlagSet. See flag.FlagSet and pflag.FlagSet.
//
// Bind returns ErrorInvalidType if v is not a pointer to a struct.
//
// For each field of v, Bind attempts to bind a new flag in fs if it is a
// supported type, or a pointer to a supported type. If the field is a nil
// pointer, it is initialized. See FlagSet for a list of supported types.
//
// Bind panics if a duplicate flag name occurs.
//
// If v contains nested or embedded structs, their fields are parsed
// recursively. By default the names of nested struct fields are prepended with
// the name(s) of their parent(s) separated by PrefixSeparator to help avoid
// flag name collisions. Explicit names of nested or embedded structs with a
// trailing "." or "-" will not have the PrefixSeparator appended.
//
// The prefix can be omitted for a nested struct with the `flatten` <option>.
// See Flag Settings below.
//
// By default, the flag names of embedded embedded struct fields are treated as
// if they are part of the top level struct. However, an explicit flag name may
// be given to an embedded struct to unflatten its fields like a nested struct.
//
//
// Flag Settings
//
// The settings for a particular flag can be customized using a struct field
// tag of the form:
//
//      `flag:"<name>[,<short>][;<default>[;<usage>[;<options>]]]"`
//
// The tag is optional and not all values need to be provided. Semi-colons only
// must be added to distinguish subsequent values if earlier ones are omitted.
//
//
// <name> - The name of the flag. All leading dashes are trimmed. If empty, the
// flag name defaults to the "kebab case" of the field name. For example,
// `ThisFieldName` would have the default flag name `this-field-name`. If the
// field is a nested or embedded struct, this overrides the prefix on its
// fields.
//
//
// <short> - If fs does not implement PFlagSet, this is ignored. Otherwise, an
// optional short name may also be provided with the <name>, separated by a
// comma. The order of <name> and <short> does not matter, their lengths will
// be used to sort them out. If <short> is longer than one character, excluding
// leading dashes, then it is ignored.
//
//
// <default> - If the current value of the field is zero, Bind attempts to
// parse this as the field's default, just like it would be parsed as a flag.
// Non-zero field values override this as the default.
//
//
// <usage> - The usage string for the flag. By default, the usage for the flag
// is empty unless specified. For longer usage strings that don't fit nicely in
// a single tag, you may define subsequent fields named _ with a `use:"..."`
// tag. These are joined with a single space into the full usage. For example,
//
//       flags := struct {
//              URL string `flag:"url;;Start usage here..."`
//              _ struct{} `use:"... contiued usage goes here"`
//              _ struct{} `use:"... and more here"`
//      }{"http://www.example.com", "Query this URL"}
//      err := Bind(fs, &flags)
//
//
// <options> - A comma separated list of additional options for the flag.
//
//      hide-default - Do not print the default value of this flag in the usage
//      output.
//
//      hidden - (PFlagSet only) Do not show this flag in the usage output.
//
//      flatten - (Nested/embedded structs only) Do not prefix the name of the
//      struct to the names of its fields. This overrides any explicit name on
//      an embedded struct which would otherwise unflatten it.
//
//
// Ignoring a Field
//
// Use the tag `flag:"-"` to prevent a field from being bound to any flag. If
// the field is a nested or embedded struct then its fields are also ignored.
//
//
// Adapt flag.Value To pflag.Value When fs Implements PFlagSet
//
// The pflag.Value interface is the flag.Value interface, but with an
// additional Type() string function. This means that flag.Value cannot be used
// directly as a pflag.Value.
//
// In order to work around this when fs implements PFlagSet, Bind wraps any
// fields that implement flag.Value but not pflag.Value in a shim adapter that
// uses the underlying type name as the Type() string. This allows you to only
// need to implement flag.Value. If the field does implement pflag.Value, it is
// used directly.
func Bind(fs FlagSet, v interface{}) error {
	return BindWithPrefix(fs, v, "")
}

// TODO: update docs
// TODO: option for flag case conversion

func BindWithPrefix(fs FlagSet, v interface{}, prefix string) error {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		return ErrorInvalidType{v, false}
	}
	if ptr.IsNil() {
		return ErrorInvalidType{v, true}
	}
	val := reflect.Indirect(ptr)
	if val.Kind() != reflect.Struct {
		return ErrorInvalidType{v, false}
	}

	_, usePFlag := fs.(PFlagSet)

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
		if !tag.ExplicitName ||
			(usePFlag && tag.Name == tag.ShortName) {
			// No explicit name given
			// OR
			// We are using pflag and the long name is the same as
			// the short name, which is disallowed.
			tag.Name = CamelToKebabCase(fieldT.Name)
		}

		fieldV := val.Field(i)
		if fieldT.Type.Kind() != reflect.Ptr {
			// The field is not a ptr, so get a ptr to it.
			fieldV = fieldV.Addr()
		}
		allocateIfNil(fieldV)
		isZero := fieldV.Elem().IsZero()
		// Correct the fieldT.Type to refer to the underlying type.
		fieldT.Type = fieldV.Elem().Type()

		if fieldT.Type.Kind() == reflect.Struct {
			prefix := prefix
			if !tag.Flatten &&
				(!fieldT.Anonymous || tag.ExplicitName) {
				prefix += tag.Name
			}
			if prefix != "" && !strings.HasSuffix(prefix, "-") &&
				!strings.HasSuffix(prefix, ".") {
				prefix += PrefixSeparator
			}
			if err := BindWithPrefix(fs, fieldV.Interface(),
				prefix); err != nil {
				return ErrorNestedStruct{fieldT.Name, err}
			}
			continue
		}

		tag.Name = fmt.Sprintf("%v%v", prefix, tag.Name)

		// Check for extended usage tags.
		for i := i + 1; i < val.NumField(); i++ {
			// Check if next field is named "_" and has a use tag.
			usageT := valT.Field(i)
			if usageT.Name != "_" {
				break
			}
			usage, ok := usageT.Tag.Lookup("use")
			if !ok {
				break
			}
			if tag.Usage != "" {
				tag.Usage += " "
			}
			tag.Usage += usage
		}

		var newFlag bool
		switch fs := fs.(type) {
		case STDFlagSet:
			newFlag = bindSTDFlag(fs, tag, fieldV.Interface())
		case PFlagSet:
			newFlag = bindPFlag(
				fs, tag, fieldV.Interface(), fieldT.Type.Name())
		default:
			panic("unsupported FlagSet")
		}
		if !newFlag {
			continue
		}

		// Set the tag default value, if field was zero.
		if isZero && tag.Value != "" {
			if err := fs.Set(tag.Name, tag.Value); err != nil {
				return ErrorDefaultValue{fieldT.Name, tag.Value, err}
			}
		}

	}

	return nil
}

func allocateIfNil(val reflect.Value) {
	if val.IsNil() {
		val.Set(reflect.New(val.Type().Elem()))
	}
}

func bindSTDFlag(fs STDFlagSet, tag flagTag, p interface{}) bool {
	switch p := p.(type) {
	case *bool:
		val := *p
		fs.BoolVar(p, tag.Name, val, tag.Usage)
	case *time.Duration:
		val := *p
		fs.DurationVar(p, tag.Name, val, tag.Usage)
	case *int:
		val := *p
		fs.IntVar(p, tag.Name, val, tag.Usage)
	case *uint:
		val := *p
		fs.UintVar(p, tag.Name, val, tag.Usage)
	case *int64:
		val := *p
		fs.Int64Var(p, tag.Name, val, tag.Usage)
	case *uint64:
		val := *p
		fs.Uint64Var(p, tag.Name, val, tag.Usage)
	case *float64:
		val := *p
		fs.Float64Var(p, tag.Name, val, tag.Usage)
	case *string:
		val := *p
		fs.StringVar(p, tag.Name, val, tag.Usage)
	case *json.RawMessage:
		fs.Var((*JSONRawMessage)(p), tag.Name, tag.Usage)
	case flag.Value:
		fs.Var(p, tag.Name, tag.Usage)
	default:
		return false
	}
	if tag.HideDefault {
		f := fs.Lookup(tag.Name)
		f.DefValue = ""
	}
	return true
}

func bindPFlag(fs PFlagSet, tag flagTag, p interface{}, typeName string) bool {
	switch p := p.(type) {
	case *bool:
		val := *p
		fs.BoolVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]bool:
		val := *p
		fs.BoolSliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *time.Duration:
		val := *p
		fs.DurationVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]time.Duration:
		val := *p
		fs.DurationSliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *int:
		val := *p
		fs.IntVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]int:
		val := *p
		fs.IntSliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *uint:
		val := *p
		fs.UintVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]uint:
		val := *p
		fs.UintSliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *int64:
		val := *p
		fs.Int64VarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]int64:
		val := *p
		fs.Int64SliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *uint64:
		val := *p
		fs.Uint64VarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *float32:
		val := *p
		fs.Float32VarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]float32:
		val := *p
		fs.Float32SliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *float64:
		val := *p
		fs.Float64VarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]float64:
		val := *p
		fs.Float64SliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *string:
		val := *p
		fs.StringVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]string:
		val := *p
		fs.StringSliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *json.RawMessage:
		fs.VarP((*JSONRawMessage)(p), tag.Name, tag.ShortName, tag.Usage)
	case flag.Value:
		// Check if p also implements pflag.Value...
		pp, ok := p.(pflag.Value)
		if !ok {
			// If not, use the pflagValue shim...
			pp = pflagValue{p, typeName}
		}
		fs.VarP(pp, tag.Name, tag.ShortName, tag.Usage)
	default:
		return false
	}

	f := fs.Lookup(tag.Name)
	f.Hidden = tag.Hidden
	if tag.HideDefault {
		f.DefValue = ""
	}
	return true
}
