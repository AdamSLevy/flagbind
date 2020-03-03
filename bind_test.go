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

package flagbind

import (
	"bytes"
	"flag"
	"io"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BindTest stores all data for a test of Bind.
type BindTest struct {
	Name     string
	UsePFlag bool

	// F is the *struct{} to bind flags to.
	F       interface{}
	ErrBind string

	// Usage must be contain all strings in UsageContains.
	UsageContains []string

	// Usage must not contain any strings in UsageNotContains.
	UsageNotContains []string

	ParseArgs []string

	// ExpF is what we expect F to be populated to after Parse.
	ExpF interface{}

	// flag and pflag Parse return slightly different errors.
	ErrParse      string
	ErrPFlagParse string
}

// Run launches test with the appropriate test name.
func (test *BindTest) Run(t *testing.T) {
	t.Run(test.Name, test.test)
	test.UsePFlag = true
	t.Run(test.Name+" pflag", test.test)
}

// test runs a single test t.
func (test *BindTest) test(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	var flg interface {
		FlagSet
		SetOutput(io.Writer)
		Usage() string
	}
	args := test.ParseArgs
	if test.UsePFlag {
		flg = pflagSetUsage{pflag.NewFlagSet("", pflag.ContinueOnError)}
		args = append([]string{}, args...)
		for i, arg := range args {
			if arg[0:1] != "-" ||
				len(arg) == 2 {
				continue
			}
			args[i] = "-" + arg
		}
	} else {
		flg = flagSetUsage{flag.NewFlagSet("", flag.ContinueOnError)}
	}
	usageOutput := bytes.NewBuffer(nil)
	flg.SetOutput(usageOutput)

	err := Bind(flg, test.F)

	if test.ErrBind != "" {
		assert.EqualError(err, test.ErrBind, "Bind()")
		return
	}
	require.NoError(err, "Bind()")

	usage := flg.Usage()
	for _, use := range test.UsageContains {
		assert.Contains(usage, use, "flag.FlagSet.Usage()")
	}
	if test.UsePFlag {
		for _, use := range test.UsageNotContains {
			assert.NotContains(usage, use, "flag.FlagSet.Usage()")
		}
	}

	err = flg.Parse(args)

	if test.UsePFlag {
		if test.ErrPFlagParse != "" {
			assert.EqualError(err, test.ErrPFlagParse, "flag.FlagSet.Parse()")
			return
		}

	} else {
		if test.ErrParse != "" {
			assert.EqualError(err, test.ErrParse, "flag.FlagSet.Parse()")
			return
		}
	}
	require.NoError(err, "flag.FlagSet.Parse()")

	assert.Equal(test.ExpF, test.F)
}

// ValidTestFlags is a test struct that exercises all ways to declare flag
// tags.
type ValidTestFlags struct {
	skip           bool
	Ignored        bool `flag:"-"`
	DefaultInherit bool
	Default        bool `flag:";true"`
	Usage          bool `flag:";;Unique usage goes here"`
	CustomName     bool `flag:"different-flag-name"`
	WithDash       bool `flag:"-with-dash"`
	WithTwoDash    bool `flag:"--with-two-dash"`
	AutoKebab      bool
	Short          bool `flag:"s"`
	LongShort      bool `flag:"long,l"`
	ShortLong      bool `flag:"r,-rlong"`

	Hidden      bool   `flag:";;Hidden usage;hidden"`
	HideDefault string `flag:";default value;Hide default;hide-default"`

	Ptr                       *bool
	PtrDefault                *int `flag:";50"`
	DefaultInheritOverride    int  `flag:";41"`
	PtrDefaultInheritOverride *int `flag:";40"`

	Nested     StructA
	NestedFlat StructB `flag:";;;flatten"`

	StructA // embedded
	StructB `flag:"embedded"`

	Bool         bool          `flag:";false"`
	Int          int           `flag:";0"`
	Int64        int64         `flag:";0"`
	Uint         uint          `flag:";0"`
	Uint64       uint64        `flag:";0"`
	Float64      float64       `flag:";0"`
	Duration     time.Duration `flag:";0"`
	String       string
	Value        TestValue
	ValueDefault TestValue `flag:";true;"`
}

type StructA struct {
	StructABool bool
}
type StructB struct {
	StructBBool bool
}

func TestBind(t *testing.T) {
	for _, test := range tests {
		test.Run(t)
	}
}

var tests = []BindTest{
	{
		Name: "invalid type",
		F: struct {
			Bool bool
		}{},
		ErrBind: ErrorInvalidType.Error(),
	}, {
		Name:    "invalid type",
		F:       new(int),
		ErrBind: ErrorInvalidType.Error(),
	}, {
		Name: "valid",
		F: &ValidTestFlags{
			DefaultInherit:            true,
			DefaultInheritOverride:    43,
			PtrDefault:                func() *int { b := 55; return &b }(),
			PtrDefaultInheritOverride: func() *int { i := 44; return &i }(),
		},
		UsageContains:    []string{"Unique usage goes here"},
		UsageNotContains: []string{"Hidden usage", "default value"},
		ParseArgs: []string{
			"-different-flag-name",
			"-with-dash",
			"-with-two-dash",
			"-auto-kebab",
			"-hidden",
			"-bool",
			"-s",
			"-int", "4",
			"-int64", "5",
			"-uint", "6",
			"-uint64", "7",
			"-float64", "0.5",
			"-duration", "1m",
			"-string", "string val",
			"-value", "true",
			"-rlong",
			"-struct-a-bool",
			"-struct-b-bool",
			"-nested-struct-a-bool",
			"-embedded-struct-b-bool",
		},
		ExpF: &ValidTestFlags{
			Default:                   true,
			DefaultInherit:            true,
			CustomName:                true,
			Hidden:                    true,
			WithDash:                  true,
			WithTwoDash:               true,
			AutoKebab:                 true,
			Short:                     true,
			ShortLong:                 true,
			Ptr:                       func() *bool { b := false; return &b }(),
			PtrDefault:                func() *int { b := 55; return &b }(),
			DefaultInheritOverride:    43,
			PtrDefaultInheritOverride: func() *int { i := 44; return &i }(),
			Bool:                      true,
			Int:                       4,
			Int64:                     5,
			Uint:                      6,
			Uint64:                    7,
			Float64:                   0.5,
			Duration:                  time.Minute,
			String:                    "string val",
			HideDefault:               "default value",
			Value:                     true,
			ValueDefault:              true,
			Nested:                    StructA{true},
			NestedFlat:                StructB{true},
			StructA:                   StructA{true},
			StructB:                   StructB{true},
		},
	}, {
		Name: "ignored",
		F:    &ValidTestFlags{},
		ParseArgs: []string{
			"-ignored",
		},
		ExpF: &ValidTestFlags{
			Ignored:    false,
			Default:    true,
			PtrDefault: func() *int { b := 50; return &b }(),
		},
		ErrParse:      "flag provided but not defined: -ignored",
		ErrPFlagParse: "unknown flag: --ignored",
	}, {
		Name: "skip unexported",
		F:    &ValidTestFlags{},
		ParseArgs: []string{
			"-skip",
		},
		ExpF: &ValidTestFlags{
			Ignored:    false,
			Default:    true,
			PtrDefault: func() *int { b := 50; return &b }(),
		},
		ErrParse:      "flag provided but not defined: -skip",
		ErrPFlagParse: "unknown flag: --skip",
	}, {
		Name: "invalid nested struct",
		F: &struct {
			E struct {
				Value TestValue `flag:";asdf;"`
			}
		}{},
		ErrBind: ErrorNestedStruct{"E",
			ErrorDefaultValue{"Value", "asdf", nil}}.Error(),
	}, {
		Name: "invalid default Value",
		F: &struct {
			Value TestValue `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Value", "asdf", nil}.Error(),
	}, {
		Name: "invalid default bool",
		F: &struct {
			Bool bool `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Bool", "asdf", nil}.Error(),
	}, {
		Name: "invalid default int",
		F: &struct {
			Int int `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Int", "asdf", nil}.Error(),
	}, {
		Name: "invalid default uint",
		F: &struct {
			Uint uint `flag:";-1;"`
		}{},
		ErrBind: ErrorDefaultValue{"Uint", "-1", nil}.Error(),
	}, {
		Name: "invalid default uint64",
		F: &struct {
			Uint64 uint64 `flag:";-1;"`
		}{},
		ErrBind: ErrorDefaultValue{"Uint64", "-1", nil}.Error(),
	}, {
		Name: "invalid default int64",
		F: &struct {
			Int64 int64 `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Int64", "asdf", nil}.Error(),
	}, {
		Name: "invalid default float64",
		F: &struct {
			Float64 float64 `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Float64", "asdf", nil}.Error(),
	}, {
		Name: "invalid default time.Duration",
		F: &struct {
			Duration time.Duration `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Duration", "asdf", nil}.Error(),
	},
}
