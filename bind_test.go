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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
	Opts     []Option

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

	err := Bind(flg, test.F, test.Opts...)

	if test.ErrBind != "" {
		assert.EqualError(err, test.ErrBind, "Bind()")
		return
	}
	require.NoError(err, "Bind()")

	usage := flg.Usage()
	for _, use := range test.UsageContains {
		assert.Contains(usage, use, "flag.FlagSet.Usage()")
	}
	//fmt.Println(usage)
	if test.UsePFlag {
		for _, use := range test.UsageNotContains {
			assert.NotContains(usage, use, "flag.FlagSet.Usage()")
		}
	}

	err = flg.Parse(args)

	if test.UsePFlag {
		if test.ErrPFlagParse != "" {
			assert.EqualError(err, test.ErrPFlagParse,
				"flag.FlagSet.Parse()")
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

	ExtendedUsage bool     `flag:";;"`
	_             struct{} `use:"Extended usage, "`
	_             struct{} `use:"continue usage"`

	Hidden      bool     `flag:";;Hidden usage;hidden"`
	HideDefault string   `flag:";default value;Hide default;hide-default"`
	_           struct{} // no use tag

	Ptr                       *bool
	PtrDefault                *int `flag:";50"`
	DefaultInheritOverride    int  `flag:";41"`
	PtrDefaultInheritOverride *int `flag:";40"`

	Nested     *StructA
	NestedFlat StructB  `flag:";;;flatten"`
	_          struct{} `flag:"struct-b-bool;;StructBBool"`
	_          struct{} `use:"continued"`
	_          struct{} `use:"twice"`
	_          struct{} `flag:"struct-b-bool;true;;hide-default"`
	_          struct{} `flag:"nested-struct-a-bool;true;;hidden,hide-default"`

	StructA // embedded
	StructB `flag:"embedded"`

	Bool         bool          `flag:";false"`
	Int          int           `flag:";0"`
	Int64        int64         `flag:";0"`
	Uint         uint          `flag:";0"`
	Uint64       uint64        `flag:";0"`
	Float32      float32       `flag:";0"`
	Float64      float64       `flag:";0"`
	Duration     time.Duration `flag:";0"`
	JSON         json.RawMessage
	String       string
	Value        TestValue
	ValueDefault TestValue `flag:";true;"`
	Marshaler    TestTextMarshaler

	IP  net.IP
	IPs []net.IP

	BoolS     []bool
	IntS      []int
	Int64S    []int64
	UintS     []uint
	Float32S  []float32
	Float64S  []float64
	DurationS []time.Duration
	StringS   []string

	Unsupported UnsupportedType

	ExportedInterface interface{}

	CustomURLPtr *url.URL
	CustomURL    url.URL

	custom bool
}

func (v *ValidTestFlags) FlagBind(fs FlagSet, prefix string, opt Option) error {
	fs.BoolVar(&v.custom, prefix+"custom", false, "")
	type _ValidTestFlags ValidTestFlags
	return Bind(fs, (*_ValidTestFlags)(v), opt)
}

var _ Binder = &ValidTestFlags{}

type UnsupportedType int

type StructA struct {
	StructABool bool
	custom      bool
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
		Name:    "ErrorInvalidType_bool",
		F:       true,
		ErrBind: ErrorInvalidType{bool(true), false}.Error(),
	}, {
		Name:    "ErrorInvalidType_int_ptr",
		F:       new(int),
		ErrBind: ErrorInvalidType{new(int), false}.Error(),
	}, {
		Name:    "ErrorInvalidType_int_ptr",
		F:       new(int),
		ErrBind: ErrorInvalidType{new(int), false}.Error(),
	}, {
		Name:    "ErrorInvalidType_nil",
		ErrBind: ErrorInvalidType{nil, false}.Error(),
	}, {
		Name:    "ErrorInvalidType_*struct{}(nil)",
		F:       (*struct{})(nil),
		ErrBind: ErrorInvalidType{(*struct{})(nil), true}.Error(),
	}, {
		Name: "valid",
		F: &ValidTestFlags{
			DefaultInherit:         true,
			DefaultInheritOverride: 43,
			PtrDefault: func() *int {
				i := 55
				return &i
			}(),
			PtrDefaultInheritOverride: func() *int {
				i := 44
				return &i
			}(),
		},
		UsageContains: []string{"Unique usage goes here", "Extended usage",
			"StructBBool continued twice"},
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
			"-custom-url",
			"http://example.com",
			"-custom-url-ptr",
			"http://example.com",
			"-custom",
		},
		ExpF: &ValidTestFlags{
			Default:        true,
			DefaultInherit: true,
			CustomName:     true,
			Hidden:         true,
			WithDash:       true,
			WithTwoDash:    true,
			AutoKebab:      true,
			Short:          true,
			ShortLong:      true,
			Ptr: func() *bool {
				var b bool
				return &b
			}(),
			PtrDefault: func() *int {
				i := 55
				return &i
			}(),
			DefaultInheritOverride: 43,
			PtrDefaultInheritOverride: func() *int {
				i := 44
				return &i
			}(),
			Bool:         true,
			Int:          4,
			Int64:        5,
			Uint:         6,
			Uint64:       7,
			Float64:      0.5,
			Duration:     time.Minute,
			String:       "string val",
			HideDefault:  "default value",
			Value:        true,
			ValueDefault: true,
			Nested:       &StructA{true, false},
			NestedFlat:   StructB{true},
			StructA:      StructA{true, false},
			StructB:      StructB{true},
			CustomURL: func() url.URL {
				u := mustParseURL("http://example.com")
				return *u
			}(),
			CustomURLPtr: mustParseURL("http://example.com"),
			custom:       true,
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
		ErrParse:      "flag provided but not defined: -skip",
		ErrPFlagParse: "unknown flag: --skip",
	}, {
		Name: "invalid short name ignored",
		F: &struct {
			E bool `flag:"lg,long"`
		}{},
		ParseArgs: []string{
			"-lg",
		},
		ErrParse:      "flag provided but not defined: -lg",
		ErrPFlagParse: "unknown flag: --lg",
	}, {
		Name: "valid JSON",
		F: &struct {
			E json.RawMessage `flag:"json"`
		}{},
		ExpF: &struct {
			E json.RawMessage `flag:"json"`
		}{E: json.RawMessage(`{"hello":"world"}`)},
		ParseArgs: []string{
			`-json`, `{"hello":"world"}`,
		},
	}, {
		Name: "invalid JSON",
		F: &struct {
			E json.RawMessage `flag:"json"`
		}{},
		ParseArgs: []string{
			`-json`, `asdf{"hello":"world"}`,
		},
		ErrParse:      "invalid value \"asdf{\\\"hello\\\":\\\"world\\\"}\" for flag -json: invalid character 'a' looking for beginning of value",
		ErrPFlagParse: "invalid argument \"asdf{\\\"hello\\\":\\\"world\\\"}\" for \"--json\" flag: invalid character 'a' looking for beginning of value",
	}, {
		Name: "ErrorNestedStruct",
		F: &struct {
			E struct {
				Value TestValue `flag:";asdf;"`
			}
		}{},
		ErrBind: ErrorNestedStruct{"E",
			ErrorDefaultValue{"Value", "asdf", nil}}.Error(),
	}, {
		Name: "ErrorDefaultValue",
		F: &struct {
			Value TestValue `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Value", "asdf", nil}.Error(),
	}, {
		Name: "ErrorFlagOverrideUndefined",
		F: &struct {
			_         struct{} `flag:"undefined;true"`
			Undefined bool
		}{},
		ErrBind: ErrorFlagOverrideUndefined{"undefined"}.Error(),
	}, {
		Name: "Duplicate Flag name",
		F: &struct {
			Duplicate  bool
			Duplicate_ bool `flag:"duplicate"`
		}{},
		ErrBind: fmt.Errorf("flag redefined: %v", "duplicate").Error(),
	}, {
		Name: "NoAutoFlatten",
		Opts: []Option{NoAutoFlatten()},
		F: &struct {
			http.Client
		}{},
		ParseArgs: []string{
			"-client-timeout=5s",
		},
		ExpF: &struct {
			http.Client
		}{http.Client{Timeout: 5 * time.Second}},
	}, {
		Name: "Prefix",
		Opts: []Option{Prefix("http-")},
		F: &struct {
			http.Client
		}{},
		ParseArgs: []string{
			"-http-timeout=5s",
		},
		ExpF: &struct {
			http.Client
		}{http.Client{Timeout: 5 * time.Second}},
	}, {
		Name: "Marshaler",
		F: &struct {
			E TestTextMarshaler `flag:"marshaler"`
		}{},
		ExpF: &struct {
			E TestTextMarshaler `flag:"marshaler"`
		}{E: TestTextMarshaler{v: "value"}},
		ParseArgs: []string{
			`-marshaler`, `value`,
		},
	}, {
		Name: "Marshaler error",
		F: &struct {
			E TestTextMarshaler `flag:"marshaler"`
		}{E: TestTextMarshaler{err: errors.New("bad")}},
		ParseArgs: []string{
			`-marshaler`, `value`,
		},
		ErrParse:      `invalid value "value" for flag -marshaler: bad`,
		ErrPFlagParse: `invalid argument "value" for "--marshaler" flag: bad`,
	},
}

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}
