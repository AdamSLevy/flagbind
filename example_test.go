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

package flagbinder_test

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/AdamSLevy/flagbinder"
	"github.com/spf13/pflag"
)

// Flags is a struct that exercises all ways to declare flag tags.
type Flags struct {
	// Unexported fields are ignored.
	skip bool

	// Explicitly ignored field.
	Ignored bool `flag:"-"`

	// Explicitly set <default>.
	Default bool `flag:";true"`

	// Explicitly set <usage>, but not <name> or <default>
	Usage bool `flag:";;Unique usage goes here"`

	// Custom <name>
	CustomName bool `flag:"different-flag-name"`

	// Dashes are trimmed and have no effect.
	WithDash    bool `flag:"-with-dash"`
	WithTwoDash bool `flag:"--with-two-dash"`

	// The default name for this flag is `auto-kebab`.
	AutoKebab bool

	// When using pflag this will be named --short and -s.
	// When using flag, this will just be named -s.
	Short bool `flag:"s"`

	// Order doesn't matter for short flags, only length.
	// Short flags are ignored when using flag, and not pflag.
	LongShort bool `flag:"long,l"`
	ShortLong bool `flag:"r,-rlong"`

	// When using pflag, this flag will be hidden from the usage.
	Hidden bool `flag:";;Hidden usage;hidden"`

	// The default value of this flag will be hidden from the usage.
	HideDefault string `flag:";default value;Hide default;hide-default"`

	// These pointers will be allocated by Bind if they are nil.
	Ptr *bool
	// This will be allocated and set to true.
	PtrDefault *bool `flag:";true"`

	// Set <name> and <usage> but not <default>
	Bool    bool    `flag:"bool;;Set the Bool true."`
	Int     int     `flag:";0"`
	Int64   int64   `flag:";0"`
	Uint    uint    `flag:";0"`
	Uint64  uint64  `flag:";0"`
	Float64 float64 `flag:";0"`

	// <default> will be 1*time.Hour
	Duration     time.Duration `flag:";1h"`
	String       string
	Value        TestValue
	ValueDefault TestValue `flag:";true"`

	// Nested and embedded structs are also parsed
	Nested     StructA
	NestedFlat StructB `flag:";;;flatten"`

	StructA // embedded
	StructB `flag:"embedded"`
}

type StructA struct {
	StructABool bool
}
type StructB struct {
	StructBBool bool
}

type TestValue bool

func (v *TestValue) Set(text string) error {
	switch strings.ToLower(text) {
	case "true":
		*v = true
	case "false":
		*v = false
	default:
		return fmt.Errorf("could not parse %q as TestValue", text)
	}
	return nil
}
func (v TestValue) String() string {
	return fmt.Sprint(bool(v))
}

func ExampleBind() {
	// Set some defaults
	f := Flags{String: "inherit this default"}
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	if err := flagbinder.Bind(fs, &f); err != nil {
		log.Fatal(err)
	}

	if err := fs.Parse([]string{
		"--bool",
		"--hidden",
		"-s",
		"--duration", "1m",
		"--auto-kebab",
		"--nested-struct-a-bool",
		"--struct-b-bool",
		"--struct-a-bool",
		"--embedded-struct-b-bool",
	}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("--bool", f.Bool)
	fmt.Println("--hidden", f.Hidden)
	fmt.Println("--short", f.Short)
	fmt.Println("--duration", f.Duration)
	// Output:
	// --bool true
	// --hidden true
	// --short true
	// --duration 1m0s
}
