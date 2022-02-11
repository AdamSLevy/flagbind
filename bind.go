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
// FlagSet, which can be the standard flag package or the popular pflag package
// github.com/spf13/pflag.
//
// By coupling the flag definitions with a type definition, the use of globals
// is discouraged, and flag types may be coupled with related behavior allowing
// for better organized and documented code.
//
// Flag names, usage, defaults and other options can be set using struct tags
// on exported fields. Using struct nesting, flags can be composed and assigned
// a flag name prefix. Exposing the settings of another package as flags is as
// simple as embedding the relevant types.
//
// Alternatively, any type may implement the Binder interface, which allows for
// more direct control over the FlagSet, much like json.Unmarshal passes
// control to types that implement json.Unmarshaler. Also similar to
// json.Unmarshal, Bind will initialize any fields that are nil, and leave any
// fields that are already populated, as defaults.
//
// See Bind documentation for the full details on controling flags.
//
// Bind works seamlessly with both the standard library flag package and the
// popular github.com/spf13/pflag package.
//
// If pflag is used, Bind adapts a flag.Value to a pflag.Value. The underlying
// type name of the flag.Value is used as the return value of the additional
// `Type() string` function required by the pflag.Value interface.
//
//
// Getting Started
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
package flagbind

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Separator is used to separate a prefix from a flag name and as the separator
// passed to FromCamelCase.
var Separator = "-"

// Binder binds itself to a FlagSet.
//
// FlagBind should prepend the `prefix` to any flag names when adding flags to
// `fs` to avoid potential flag name conflicts and allow more portable
// implementations.
//
// Additionally the opt should be passed down to Bind to preserve original opts
// if called again by the implementation.
//
// The underlying type of `fs` is the same as the original FlagSet passed to
// Bind.
type Binder interface {
	FlagBind(fs FlagSet, prefix string, opt Option) error
}

// Bind the exported fields of struct `v` to new flags in the FlagSet `fs`.
//
// Bind returns ErrorInvalidFlagSet if `fs` does not implement STDFlagSet or
// PFlagSet.
//
// Bind returns ErrorInvalidType if `v` is not a pointer to a struct.
//
// Bind recovers from FlagSet panics and instead returns the panic as an error
// if a duplicate flag name occurs.
//
// For each exported field of `v` Bind attempts to define one or more
// corresponding flags in `fs` according to the following rules.
//
// If the field is a nil pointer, it is initialized.
//
// If the field implements Binder, then only FlagBind is called on the field.
//
// If the field implements flag.Value and not Binder, then it is bound as a
// Value on the FlagSet.
//
// Otherwise, if the field is a struct, or struct pointer, then Bind is
// recursively called on a pointer to the struct field.
//
// If the field is any supported type, a new flag is defined in `fs` with the
// settings defined in the field's `flag:"..."` tag. If the field is non-zero,
// its value is used as the default for that flag instead of whatever is
// defined in the `flag:";<default>"` tag. See FlagTag Settings below.
//
// For a complete list of supported types see STDFlagSet and PFlagSet.
// Additionally, a json.RawMessage is also natively supported and is bound as a
// JSONRawMessage flag.
//
//
// Ignoring a Field
//
// Use the tag `flag:"-"` to prevent an exported field from being bound to any
// flag. If the field is a nested or embedded struct then its fields are also
// ignored.
//
//
// Flag Tag Settings
//
// The flag settings for a particular field can be customized using a struct
// field tag of the form:
//
//      `flag:"[<long>][,<short>][;<default>[;<usage>[;<options>]]]"`
//
// The tag and all of its settings are [optional]. Semi-colons are used to
// distinguish subsequent settings.
//
//
// <long>[,<short>] - Explicitly set the long and short names of the flag. All
// leading dashes are trimmed from both names. The two names are sorted for
// length, and the short name must be a single character, else it is ignored.
//
// If `fs` does not implement PFlagSet, then the short name is ignored if a
// long name is defined, otherwise the short name is used as the long name.
//
// If `fs` does implement PFlagSet, and only a short flag is defined, the long
// name defaults to the field name in kebab-case.
//
// If no name is set, the long name defaults to the field name in "kebab-case".
// For example, "ThisFieldName" becomes "this-field-name". See FromCamelCase
// and Separator.
//
// If the field is a nested or embedded struct and the "flatten" option is not
// set (see below), then the name is used as a prefix for all nested field flag
// names.
//
//
// <default> - Bind attempts to parse <default> as the field's default, just
// like it would be parsed as a flag. Non-zero field values override this as
// the default.
//
//
// <usage> - The usage string for the flag. See Extended Usage below for a way
// to break longer usage strings across multiple lines.
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
// tag on a single line. To keep line lengths shorter, use any number of blank
// identifier fields of any type with a `use` field tag to extend the usage of
// a flag. Each `use` tag is joined with a single space.
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
// Nested/Embedded Structs Flag Prefix
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
func Bind(fs FlagSet, v interface{}, opts ...Option) error {
	return newBind(opts...).bind(fs, v)
}

func BindCobra(cmd *cobra.Command, v interface{}, opts ...Option) (err error) {
	if err = newBind(append(opts, CobraFilter("persistent"))...).bind(cmd.PersistentFlags(), v); err != nil {
		return err
	}
	if err = newBind(append(opts, CobraFilter("local"))...).bind(cmd.LocalFlags(), v); err != nil {
		return err
	}
	if err = newBind(append(opts, CobraFilter("flags"))...).bind(cmd.Flags(), v); err != nil {
		return err
	}
	return nil
}

func (b bind) bind(fs FlagSet, v interface{}) (err error) {

	// Hand control over to the Binder implementation.
	if binder, ok := v.(Binder); ok {
		return binder.FlagBind(fs, b.Prefix, b.Option())
	}

	// Ensure we have a non-nil pointer.
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		return ErrorInvalidType{v, false}
	}
	if ptr.IsNil() {
		return ErrorInvalidType{v, true}
	}

	// We must operate on the addressable value, not the pointer.
	val := reflect.Indirect(ptr)

	// We can only inspect structs.
	if val.Kind() != reflect.Struct {
		return ErrorInvalidType{v, false}
	}

	// The flag and pflag packages panic when a flag with a duplicate name
	// is defined. This works well for identifying the offending line of
	// code where the flag name is redefined, but that is just noise to
	// users of this package. The only useful information from such a panic
	// is the duplicate flagname included in the panic message.
	defer func() {
		if r := recover(); r != nil {
			// Clean up the inconsistent leading space that pflag
			// leaves behind if no FlagSet name was set.
			r = strings.TrimSpace(fmt.Sprintf("%v", r))
			err = fmt.Errorf("%v", r)
		}
	}()

	_, usePFlag := fs.(PFlagSet)

	valT := val.Type()

	defaults := make(map[string]string)

	// loop through all fields
	for i := 0; i < val.NumField(); i++ {

		structField := valT.Field(i)

		// Special flag metadata may be set using the blank identifier.
		isMetadata := structField.Name == "_"

		// See reflect.StructField for details.
		isExported := structField.PkgPath == ""

		// Ignore unexported, non-metadata fields.
		if !isExported && !isMetadata {
			continue
		}

		// Parse the flagTag.
		tagStr, hasTag := structField.Tag.Lookup("flag")
		tag := newFlagTag(tagStr)

		if b.IsIgnored(tag) {
			continue
		}

		// Auto populate name if it has no explicit name, or only has a
		// short name.
		if !tag.HasExplicitName ||
			(usePFlag && tag.Name == tag.ShortName) {
			tag.Name = FromCamelCase(structField.Name, Separator)
		}

		fieldV := val.Field(i)

		i = loadExtendedUsage(i, valT, &tag)

		// Update Flag with Metadata tag.
		if isMetadata {
			if hasTag {
				if err := overrideFlag(fs, tag); err != nil {
					return err
				}
			}
			continue
		}

		// Ensure we are dealing with a pointer.
		if structField.Type.Kind() != reflect.Ptr {
			fieldV = fieldV.Addr()
		}

		// Obtain the underlying type of the field.
		fieldT := fieldV.Type().Elem()

		// Allocate the field pointer if nil.
		if fieldV.IsNil() {
			fieldV.Set(reflect.New(fieldT))
		}

		fieldI := fieldV.Interface()

		_, isBinder := fieldI.(Binder)

		_, isFlagValue := fieldI.(flag.Value)
		_, isJSONRawMessage := fieldI.(*json.RawMessage)
		_, isURL := fieldI.(*url.URL)
		noDive := isFlagValue || isJSONRawMessage || isURL

		isStruct := fieldT.Kind() == reflect.Struct

		// If the field implements Binder, we call Bind on the field,
		// which will call its Binder implementation.
		//
		// If the field is a struct, and does not implement flag.Value,
		// we will recursively call BindWithPrefix.
		//
		// Otherwise, if the field implements flag.Value or any other
		// type supported, we will bind the field directly below.
		if isBinder || (!noDive && isStruct) {

			// Set prefix up to this point.
			b := b

			// If the nested field is not explicitly flattened AND
			// ( auto flattening is disabled OR the field is not
			// anonymous (embedded) OR has an explicit name ),
			// then grow the prefix.
			if !tag.Flatten &&
				(b.NoAutoFlatten ||
					!structField.Anonymous || tag.HasExplicitName) {
				b.Prefix += tag.Name
			}

			b.Prefix = appendSeparator(b.Prefix)

			if err := b.bind(fs, fieldI); err != nil {
				return newErrorNestedStruct(structField.Name, err)
			}
			continue
		}

		tag.Name = fmt.Sprintf("%v%v", b.Prefix, tag.Name)

		newFlag, err := bindField(fs, tag, fieldI, fieldT.Name())
		if err != nil {
			return err
		}
		if !newFlag {
			continue
		}

		// If field value was zero, then set the tag default, if
		// specified.
		if fieldV.Elem().IsZero() && tag.DefValue != "" {
			defaults[tag.Name] = tag.DefValue
			if err := fs.Set(tag.Name, tag.DefValue); err != nil {
				return ErrorDefaultValue{structField.Name, tag.DefValue, err}
			}
		}
	}

	return setDefaults(fs, defaults)
}

func setDefaults(fs FlagSet, defaults map[string]string) error {
	switch fs := fs.(type) {
	case STDFlagSet:
		fs.VisitAll(func(f *flag.Flag) {
			defVal, ok := defaults[f.Name]
			if !ok {
				return
			}
			f.DefValue = defVal
		})
	case PFlagSet:
		fs.VisitAll(func(f *pflag.Flag) {
			defVal, ok := defaults[f.Name]
			if !ok {
				return
			}
			f.DefValue = defVal
		})
	default:
		return ErrorInvalidFlagSet
	}
	return nil

}

func loadExtendedUsage(i int, valT reflect.Type, tag *flagTag) int {
	// Check for extended usage tags.
	for i++; i < valT.NumField(); i++ {
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
	return i
}

func appendSeparator(prefix string) string {
	// Do not append separator to an empty prefix.
	if prefix == "" {
		return prefix
	}

	// Do not append separator when other common separators are being used.
	for _, sep := range []string{"-", ".", "_"} {
		if strings.HasSuffix(prefix, sep) {
			return prefix
		}
	}

	return prefix + Separator
}

func bindField(fs FlagSet, tag flagTag, p interface{}, typeName string) (bool, error) {
	switch fs := fs.(type) {
	case STDFlagSet:
		return bindSTDFlag(fs, tag, p), nil
	case PFlagSet:
		return bindPFlag(fs, tag, p, typeName), nil
	default:
		return false, ErrorInvalidFlagSet
	}
}
func bindSTDFlag(fs STDFlagSet, tag flagTag, p interface{}) bool {

	switch p := p.(type) {
	case flag.Value:
		fs.Var(p, tag.Name, tag.Usage)
	case *json.RawMessage:
		fs.Var((*JSONRawMessage)(p), tag.Name, tag.Usage)
	case *url.URL:
		fs.Var((*URL)(p), tag.Name, tag.Usage)
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
	case *url.URL:
		f = fs.VarPF((*URL)(p), tag.Name, tag.ShortName, tag.Usage)
	case *net.IP:
		val := *p
		fs.IPVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
	case *[]net.IP:
		val := *p
		fs.IPSliceVarP(p, tag.Name, tag.ShortName, val, tag.Usage)
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

func overrideFlag(fs FlagSet, tag flagTag) error {
	// Update flag if it exists.
	switch fs := fs.(type) {
	case STDFlagSet:
		return overrideSTDFlag(fs, tag)
	case PFlagSet:
		return overridePFlag(fs, tag)
	default:
		return ErrorInvalidFlagSet
	}
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
