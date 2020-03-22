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

// Package flagbind makes defining flags as simple as defining a struct type.
//
// flagbind.Bind parses the exported fields of a struct and binds them to a
// FlagSet. This works with the standard flag package as well as
// github.com/spf13/pflag.
//
// Start by declaring a struct type for your flags.
//
//      var flags := struct {
//              StringFlag string `flag:"flag-name;default value;Usage for string-flag"`
//              Int        int    `flag:"integer;5"`
//
//              // Flag names default to `auto-kebab-case`
//              AutoKebabCase int
//
//              // If pflag is used, -s is be used as the shorthand flag name,
//              // otherwise it is ignored for use with the standard flag package.
//              ShortName bool `flag:"short,s"`
//
//              // Optionally extende the usage tag with subsequent `use` tags
//              // on _ fields.
//              URL string `flag:"url,u;http://www.example.com/;Start usage here"
//              _   struct{} `use:"continue longer usage string for --url below it",
//
//              // Nested and Embedded structs can add a flag name prefix, or not.
//              Nested     StructA
//              NestedFlat StructB           `flag:";;;flatten"`
//              StructA                      // Flat by default
//              StructB    `flag:"embedded"` // Add prefix to nested field flag names.
//
//              // Ignored
//              ExplicitlyIgnored bool `flag:"-"`
//              unexported        bool
//      }{
//              // Default values may also be set directly to override the tag.
//              StringFlag: "override tag default",
//      }
//
//      fs := pflag.NewFlagSet("", pflag.ContinueOnError)
//      flagbind.Bind(fs, &flags)
//      fs.Parse([]string{"--auto-kebab-case"})
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

// Separator is used to separate a prefix from a flag name and as the separator
// passed to FromCamelCase.
var Separator = "-"

// Bind the exported fields of struct v to new flags in fs.
//
// Bind returns ErrorInvalidFlagSet if fs does not implement STDFlagSet or
// PFlagSet.
//
// Bind returns ErrorInvalidType if v is not a pointer to a struct.
//
// For each exported field of v that is a supported type (or a pointer to a
// supported type), Bind defines a corresponding flag in fs.
//
// For a complete list of supported types see STDFlagSet and PFlagSet.
// Additionally, a json.RawMessage is also supported and is parsed as a
// JSONRawMessage flag.
//
// If the field is a nil pointer, it is initialized.
//
// Bind returns an error if a duplicate flag name occurs.
//
//
// Ignoring a Field
//
// Use the tag `flag:"-"` to prevent a field from being bound to any flag. If
// the field is a nested or embedded struct then its fields are also ignored.
//
//
// Flag Tag Settings
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
// is empty unless specified. See Extended Usage below for a way to break
// longer usage strings across multiple lines.
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
// Extended Usage
//
// Usage lines can frequently be longer than what comfortably fits in a flag
// tag on a single line. To keep line lengths shorter, any number of blank
// identifier fields (`_`), each with a `use` tag, may be defined immediately
// following a `flag` tag. Each `use` tag is joined with the existing usage
// with a single space inserted where needed.
//
//      type Flags struct {
//              URL string   `flag:"url;;Usage starts here"`
//              _   struct{} `use:"and continues here"`
//              _   struct{} `use:"and ends here."`
//      }
//
//
// Auto-Adapt flag.Value To pflag.Value
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
//
//
// Nested/Embedded Structs
//
// If the field is a nested or embedded struct, its fields are also recursively
// parsed.
//
// In order to help avoid flag name collisions, child flag names may be
// prepended with a prefix. The prefix defaults to the parent's field name
// passed through FromCamelCase, with a trailing Separator.
//
// The prefix may be set explicitly using the <name> on the parent's Flag Tag.
//
// To allow for a distinct separator symbol to be used just for a prefix, an
// explicitly set prefix that ends in "-", "_", or "." will not have Separator
// appended.
//
// By default, flags in nested structs always have a prefix, but this can be
// omitted with Flag Tag `flatten` <option>.
//
// By default, flags in embedded structs do not given a prefix, but one can be
// added by setting an explicit Flag Tag <name>.
//
//
// Overriding Flag Settings
//
// It is not always possible to set a Flag Tag on the fields of a nested struct
// type, such as when the type is from an external package. To allow for
// setting the default value, usage, or other options, an Overriding Flag Tag
// may be specified on a blank identifier field (`_`) that occurs anywhere
// after the field that defined the overridden flag.
//
// The name specified in the Overriding Flag Tag must exactly match the flag
// name of the overridden flag, including any prefixes that were prepended due
// to nesting. Bind returns ErrorFlagOverrideUndefined if the flag name cannot
// be found.
//
// Extended Usage may also defined immediately after an Overriding Flag Tag
// field.
//
// For example, this sets the default value and usage on the flag for Timeout
// on an embedded http.Client.
//
//      type Flags struct {
//              http.Client // Defines the -timeout flag
//              _ struct{} `flag:"timeout;5s;HTTP request timeout"`
//              _ struct{} `use:"... continued usage"`
//      }
func Bind(fs FlagSet, v interface{}) error {
	return BindWithPrefix(fs, v, "")
}

// BindWithPrefix is the same as Bind, but all flag names are prefixed with
// `prefix`. Separator is NOT appended to prefix.
func BindWithPrefix(fs FlagSet, v interface{}, prefix string) (err error) {
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

	// The flag and pflag packages panic when a flag with a duplicate name
	// is defined. This works well for identifying the offending line of
	// code where the flag name is redefined, but that is just noise to
	// users of this package. The only useful information from such a panic
	// is the redefined flagname included in the panic message.
	defer func() {
		if r := recover(); r != nil {
			// Clean up the leading space that pflag leaves behind
			// if no FlagSet name was set.
			r = strings.TrimSpace(fmt.Sprintf("%v", r))
			err = fmt.Errorf("%v", r)
		}
	}()

	_, usePFlag := fs.(PFlagSet)

	valT := val.Type()
	// loop through all fields
	for i := 0; i < val.NumField(); i++ {
		fieldT := valT.Field(i)
		isOverride := fieldT.Name == "_"
		if fieldT.PkgPath != "" && !isOverride {
			// unexported field
			continue
		}
		tagStr, hasTag := fieldT.Tag.Lookup("flag")
		tag := newFlagTag(tagStr)
		if tag.isIgnored {
			continue
		}

		fieldV := val.Field(i)

		// Auto populate name if needed...
		if !tag.hasExplicitName ||
			(usePFlag && tag.Name == tag.ShortName) {
			// No explicit name given OR
			// We are using pflag and the long name is the same as
			// the short name, which is disallowed.
			tag.Name = FromCamelCase(fieldT.Name, Separator)
		}
		// Check for extended usage tags.
		for i++; i < val.NumField(); i++ {
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
		i--

		// Flag override...
		if isOverride {
			if hasTag {
				// Update flag if it exists.
				var err error
				switch fs := fs.(type) {
				case STDFlagSet:
					err = overrideSTDFlag(fs, tag)
				case PFlagSet:
					err = overridePFlag(fs, tag)
				default:
					return ErrorInvalidFlagSet
				}
				if err != nil {
					return err
				}
			}
			continue
		}

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
				(!fieldT.Anonymous || tag.hasExplicitName) {
				prefix += tag.Name
			}
			if prefix != "" && !strings.HasSuffix(prefix, "-") &&
				!strings.HasSuffix(prefix, ".") &&
				!strings.HasSuffix(prefix, "_") {
				prefix += Separator
			}
			if err := BindWithPrefix(fs, fieldV.Interface(),
				prefix); err != nil {
				return newErrorNestedStruct(fieldT.Name, err)
			}
			continue
		}

		tag.Name = fmt.Sprintf("%v%v", prefix, tag.Name)

		var newFlag bool
		switch fs := fs.(type) {
		case STDFlagSet:
			newFlag = bindSTDFlag(fs, tag, fieldV.Interface())
		case PFlagSet:
			newFlag = bindPFlag(
				fs, tag, fieldV.Interface(), fieldT.Type.Name())
		default:
			return ErrorInvalidFlagSet
		}
		if !newFlag {
			continue
		}

		// Set the tag default value, if field was zero.
		if isZero && tag.DefValue != "" {
			if err := fs.Set(tag.Name, tag.DefValue); err != nil {
				return ErrorDefaultValue{fieldT.Name, tag.DefValue, err}
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
	case flag.Value:
		fs.Var(p, tag.Name, tag.Usage)
	case *json.RawMessage:
		fs.Var((*JSONRawMessage)(p), tag.Name, tag.Usage)
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
	default:
		return false
	}

	if tag.HideDefault {
		f := fs.Lookup(tag.Name)
		f.DefValue = ""
	}

	return true
}
func overrideSTDFlag(fs STDFlagSet, tag flagTag) error {

	f := fs.Lookup(tag.Name)
	if f == nil {
		return ErrorFlagOverrideUndefined{tag.Name}
	}

	if tag.DefValue != "" {
		f.Value.Set(tag.DefValue)
		f.DefValue = tag.DefValue
	}
	if tag.Usage != "" {
		f.Usage = tag.Usage
	}
	if tag.HideDefault {
		f.DefValue = ""
	}

	return nil
}

func bindPFlag(fs PFlagSet, tag flagTag, p interface{}, typeName string) bool {

	var f *pflag.Flag
	switch p := p.(type) {
	case flag.Value:
		// Check if p also implements pflag.Value...
		pp, ok := p.(pflag.Value)
		if !ok {
			// If not, use the pflagValue shim...
			pp = pflagValue{p, typeName}
		}
		f = fs.VarPF(pp, tag.Name, tag.ShortName, tag.Usage)
	case *json.RawMessage:
		f = fs.VarPF((*JSONRawMessage)(p), tag.Name, tag.ShortName, tag.Usage)
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
	default:
		return false
	}

	if !(tag.HideDefault || tag.Hidden) {
		return true
	}

	if f == nil {
		f = fs.Lookup(tag.Name)
	}

	if tag.HideDefault {
		f.DefValue = ""
	}
	f.Hidden = tag.Hidden

	return true
}
func overridePFlag(fs PFlagSet, tag flagTag) error {

	f := fs.Lookup(tag.Name)
	if f == nil {
		return ErrorFlagOverrideUndefined{tag.Name}
	}

	if tag.DefValue != "" {
		f.Value.Set(tag.DefValue)
		f.DefValue = tag.DefValue
	}
	if tag.Usage != "" {
		f.Usage = tag.Usage
	}
	if tag.HideDefault {
		f.DefValue = ""
	}
	f.Hidden = tag.Hidden

	return nil
}
