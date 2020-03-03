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
// to Bind to populate a flag.FlagSet or pflag.FlagSet.
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
//		StringFlag: "override default",
//	}
//
//	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
//	flagbind.Bind(fs, &flags)
//	fs.Parse([]string{"--auto-kebab-case"})
//
// Bind works seemlessly with both the standard library flag package and the
// popular pflag package.
//
// For types that only implement flag.Value, Bind wraps them in an adapter so
// that they can be used as a pflag.Value. The return value of the added
// function Type() string is the type name of the struct field.
//
// Additional options may be set for each flag. See Bind for the full
// documentation details.
package flagbind

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// Bind the exported fields of struct v to new flags in flg.
//
// Bind returns ErrorInvalidFlagSet if flg does not implement STDFlagSet or
// PFlagSet. See flag.FlagSet and pflag.FlagSet.
//
// Bind returns ErrorInvalidType if v is not a pointer to a struct.
//
// For each field of v, Bind attempts to bind a new flag in flg if it is a
// supported type, or a pointer to a supported type. If the field is a nil
// pointer, it is initialized. See FlagSet for a list of supported types.
//
// Bind panics if a duplicate flag name occurs.
//
// If v contains nested or embedded structs, their fields are parsed
// recursively. By default the names of nested struct fields are prepended with
// the name(s) of their parent(s) to help avoid flag name collisions. The
// prefix can be omitted for a nested struct with the `flatten` <option>. See
// Flag Settings below.
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
// <short> - If flg does not implement PFlagSet, this is ignored. Otherwise, an
// optional short name may also be provided with the <name>, separated by a
// comma. The order of <name> and <short> does not matter, but <short> may only
// be one character long, excluding leading dashes.
//
//
// <default> - If the current value of the field is zero, Bind attempts to
// parse this as the field's default, just like it would be parsed as a flag.
// Non-zero field values override this as the default.
//
//
// <usage> - The usage string for the flag. By default, the usage for the flag
// is empty unless specified.
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
//      struct to the names of its fields. This overrides the explicit name on
//      an embedded struct which would otherwise unflatten it.
//
//
// Ignoring a Field
//
// Use the tag `flag:"-"` to prevent a field from being bound to any flag. If
// the field is a nested or embedded struct then its fields are also ignored.
//
//
// Adapt flag.Value To pflag.Value When flg Implements PFlagSet
//
// The pflag.Value interface is the flag.Value interface, but with an
// additional Type() string function. This means that flag.Value cannot be used
// directly as a pflag.Value.
//
// In order to work around this when flg implements PFlagSet, Bind wraps any
// fields that implement flag.Value but not pflag.Value in a shim adapter that
// uses the underlying type name as the Type() string. This allows you to only
// need to implement flag.Value. If the field does implement pflag.Value, it is
// used directly.
func Bind(flg FlagSet, v interface{}) error {
	return bind(flg, v, "")
}

func bind(flg FlagSet, v interface{}, prefix string) error {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		return ErrorInvalidType
	}
	val := reflect.Indirect(ptr)
	if val.Kind() != reflect.Struct {
		return ErrorInvalidType
	}

	stdflg, useSTDFlag := flg.(STDFlagSet)
	pflg, usePFlag := flg.(PFlagSet)
	if !useSTDFlag && !usePFlag {
		return ErrorInvalidFlagSet
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
		if !tag.ExplicitName ||
			(usePFlag && tag.Name == tag.ShortName) {
			// No explicit name given
			// OR
			// We are using pflag and the long name is the same as
			// the short name, which is disallowed.
			tag.Name = kebabCase(fieldT.Name)
		}

		var isZero bool
		fieldV := val.Field(i)
		if fieldT.Type.Kind() != reflect.Ptr {
			isZero = fieldV.IsZero()
			// The field is not a ptr, so get a ptr to it.
			fieldV = fieldV.Addr()
		} else if fieldV.IsNil() {
			// We have a nil ptr, so allocate it.
			fieldV.Set(reflect.New(fieldT.Type.Elem()))
		} else {
			// We have a pre-allocated pointer.
			isZero = fieldV.Elem().IsZero()
		}

		if fieldT.Type.Kind() == reflect.Struct {
			prefix := prefix
			if !tag.Flatten &&
				(!fieldT.Anonymous || tag.ExplicitName) {
				prefix = strings.Trim(
					fmt.Sprintf("%v-%v", prefix, tag.Name), "-")
			}
			if err := bind(flg, fieldV.Interface(), prefix); err != nil {
				return ErrorNestedStruct{fieldT.Name, err}
			}
			continue
		}

		if prefix != "" {
			tag.Name = fmt.Sprintf("%v-%v", prefix, tag.Name)
		}

		switch p := fieldV.Interface().(type) {
		case *bool:
			val := *p
			if !usePFlag {
				stdflg.BoolVar(p, tag.Name, val, tag.Usage)
				break
			}
			pflg.BoolVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
		case *time.Duration:
			val := *p
			if !usePFlag {
				stdflg.DurationVar(p, tag.Name, val, tag.Usage)
				break
			}
			pflg.DurationVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
		case *int:
			val := *p
			if !usePFlag {
				stdflg.IntVar(p, tag.Name, val, tag.Usage)
				break
			}
			pflg.IntVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
		case *uint:
			val := *p
			if !usePFlag {
				stdflg.UintVar(p, tag.Name, val, tag.Usage)
				break
			}
			pflg.UintVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
		case *int64:
			val := *p
			if !usePFlag {
				stdflg.Int64Var(p, tag.Name, val, tag.Usage)
				break
			}
			pflg.Int64VarP(p, tag.Name, tag.ShortName, val, tag.Usage)
		case *uint64:
			val := *p
			if !usePFlag {
				stdflg.Uint64Var(p, tag.Name, val, tag.Usage)
				break
			}
			pflg.Uint64VarP(p, tag.Name, tag.ShortName, val, tag.Usage)
		case *float64:
			val := *p
			if !usePFlag {
				stdflg.Float64Var(p, tag.Name, val, tag.Usage)
				break
			}
			pflg.Float64VarP(p, tag.Name, tag.ShortName, val, tag.Usage)
		case *string:
			val := *p
			if !usePFlag {
				stdflg.StringVar(p, tag.Name, val, tag.Usage)
				break
			}
			pflg.StringVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
		case flag.Value:
			if !usePFlag {
				stdflg.Var(p, tag.Name, tag.Usage)
				break
			}
			// Check if p also implements pflag.Value...
			pp, ok := p.(pflag.Value)
			if !ok {
				// If not, use the pflagValue shim...
				pp = pflagValue{p, fieldT.Type.Name()}
			}
			pflg.VarP(pp, tag.Name, tag.ShortName, tag.Usage)
		}

		// Set the tag default value, if field was zero.
		if isZero && tag.Value != "" {
			if err := flg.Set(tag.Name, tag.Value); err != nil {
				return ErrorDefaultValue{fieldT.Name, tag.Value, err}
			}
		}

		// Apply flag options

		if !usePFlag {
			if tag.HideDefault {
				f := stdflg.Lookup(tag.Name)
				f.DefValue = ""
			}
			continue
		}

		f := pflg.Lookup(tag.Name)
		f.Hidden = tag.Hidden
		if tag.HideDefault {
			f.DefValue = ""
		}
	}

	return nil
}
