package flagbuilder

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FlagTest struct {
	Name     string
	F        interface{}
	ErrBuild string
	Usage    string
	Args     []string
	ExpF     interface{}
	ErrParse string
}

func (test *FlagTest) Run(t *testing.T) {
	t.Run(test.Name, test.test)
}

func (test *FlagTest) test(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	flg := flag.NewFlagSet("", flag.ContinueOnError)
	usageOutput := bytes.NewBuffer(nil)
	flg.SetOutput(usageOutput)

	err := Build(flg, test.F)

	if test.ErrBuild != "" {
		assert.EqualError(err, test.ErrBuild, "flagbuilder.Build()")
		return
	}
	require.NoError(err, "flagbuilder.Build()")

	if test.Usage != "" {
		flg.Usage()
		assert.Contains(string(usageOutput.Bytes()), test.Usage,
			"flag.FlagSet.Usage()")
	}

	err = flg.Parse(test.Args)

	if test.ErrParse != "" {
		assert.EqualError(err, test.ErrParse, "flag.FlagSet.Parse()")
		return
	}
	require.NoError(err, "flag.FlagSet.Parse()")

	assert.Equal(test.ExpF, test.F)
}

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

	Ptr               *bool
	PtrDefault        *bool `flag:";true"`
	PtrDefaultInherit *bool

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

var tests = []FlagTest{
	{
		Name: "invalid type",
		F: struct {
			Bool bool
		}{},
		ErrBuild: ErrorInvalidType.Error(),
	}, {
		Name:     "invalid type",
		F:        new(int),
		ErrBuild: ErrorInvalidType.Error(),
	}, {
		Name: "valid",
		F: &ValidTestFlags{
			DefaultInherit:    true,
			PtrDefaultInherit: func() *bool { b := true; return &b }(),
		},
		Usage: "Unique usage goes here",
		Args: []string{
			"-different-flag-name",
			"-with-dash",
			"-with-two-dash",
			"-auto-kebab",
			"-bool",
			"-int", "4",
			"-int64", "5",
			"-uint", "6",
			"-uint64", "7",
			"-float64", "0.5",
			"-duration", "1m",
			"-string", "string val",
			"-value", "true",
		},
		ExpF: &ValidTestFlags{
			Default:           true,
			DefaultInherit:    true,
			CustomName:        true,
			WithDash:          true,
			WithTwoDash:       true,
			AutoKebab:         true,
			Ptr:               func() *bool { b := false; return &b }(),
			PtrDefault:        func() *bool { b := true; return &b }(),
			PtrDefaultInherit: func() *bool { b := true; return &b }(),
			Bool:              true,
			Int:               4,
			Int64:             5,
			Uint:              6,
			Uint64:            7,
			Float64:           0.5,
			Duration:          time.Minute,
			String:            "string val",
			Value:             true,
			ValueDefault:      true,
		},
	}, {
		Name: "ignored",
		F:    &ValidTestFlags{},
		Args: []string{
			"-ignored",
		},
		ExpF: &ValidTestFlags{
			Ignored:           false,
			Default:           true,
			PtrDefaultInherit: func() *bool { b := true; return &b }(),
			PtrDefault:        func() *bool { b := true; return &b }(),
		},
		ErrParse: "flag provided but not defined: -ignored",
	}, {
		Name: "skip unexported",
		F:    &ValidTestFlags{},
		Args: []string{
			"-skip",
		},
		ExpF: &ValidTestFlags{
			Ignored:           false,
			Default:           true,
			PtrDefaultInherit: func() *bool { b := true; return &b }(),
			PtrDefault:        func() *bool { b := true; return &b }(),
		},
		ErrParse: "flag provided but not defined: -skip",
	}, {
		Name: "invalid default Value",
		F: &struct {
			Value TestValue `flag:";asdf;"`
		}{},
		ErrBuild: ErrorDefaultValue{"Value", "asdf",
			fmt.Errorf(`could not parse "asdf" as TestValue`)}.Error(),
	}, {
		Name: "invalid default bool",
		F: &struct {
			Bool bool `flag:";asdf;"`
		}{},
		ErrBuild: ErrorDefaultValue{"Bool", "asdf",
			fmt.Errorf(`strconv.ParseBool: parsing "asdf": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default int",
		F: &struct {
			Int int `flag:";asdf;"`
		}{},
		ErrBuild: ErrorDefaultValue{"Int", "asdf",
			fmt.Errorf(`strconv.ParseInt: parsing "asdf": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default uint",
		F: &struct {
			Uint uint `flag:";-1;"`
		}{},
		ErrBuild: ErrorDefaultValue{"Uint", "-1",
			fmt.Errorf(`strconv.ParseUint: parsing "-1": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default uint64",
		F: &struct {
			Uint64 uint64 `flag:";-1;"`
		}{},
		ErrBuild: ErrorDefaultValue{"Uint64", "-1",
			fmt.Errorf(`strconv.ParseUint: parsing "-1": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default int64",
		F: &struct {
			Int64 int64 `flag:";asdf;"`
		}{},
		ErrBuild: ErrorDefaultValue{"Int64", "asdf",
			fmt.Errorf(`strconv.ParseInt: parsing "asdf": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default float64",
		F: &struct {
			Float64 float64 `flag:";asdf;"`
		}{},
		ErrBuild: ErrorDefaultValue{"Float64", "asdf",
			fmt.Errorf(`strconv.ParseFloat: parsing "asdf": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default time.Duration",
		F: &struct {
			Duration time.Duration `flag:";asdf;"`
		}{},
		ErrBuild: ErrorDefaultValue{"Duration", "asdf",
			fmt.Errorf(`time: invalid duration asdf`),
		}.Error(),
	},
}

func TestBuild(t *testing.T) {
	for _, test := range tests {
		test.Run(t)
	}
}
