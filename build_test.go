package flagbuilder

import (
	"bytes"
	"flag"
	"fmt"
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
	// This is the *struct{} to bind flags to.
	F             interface{}
	ErrBind       string
	Usage         string
	Args          []string
	ExpF          interface{}
	ErrParse      string
	ErrPFlagParse string
}

func (test *BindTest) Run(t *testing.T) {
	t.Run(test.Name, test.test)
	test.UsePFlag = true
	t.Run(test.Name+" pflag", test.test)
}

func (test *BindTest) test(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	var flg interface {
		FlagSet
		SetOutput(io.Writer)
		Usage()
		Parse([]string) error
	}
	args := test.Args
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
		assert.EqualError(err, test.ErrBind, "flagbuilder.Bind()")
		return
	}
	require.NoError(err, "flagbuilder.Bind()")

	if test.Usage != "" {
		flg.Usage()
		assert.Contains(string(usageOutput.Bytes()), test.Usage,
			"flag.FlagSet.Usage()")
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
		ErrParse:      "flag provided but not defined: -ignored",
		ErrPFlagParse: "unknown flag: --ignored",
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
		ErrParse:      "flag provided but not defined: -skip",
		ErrPFlagParse: "unknown flag: --skip",
	}, {
		Name: "invalid default Value",
		F: &struct {
			Value TestValue `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Value", "asdf",
			fmt.Errorf(`could not parse "asdf" as TestValue`)}.Error(),
	}, {
		Name: "invalid default bool",
		F: &struct {
			Bool bool `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Bool", "asdf",
			fmt.Errorf(`strconv.ParseBool: parsing "asdf": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default int",
		F: &struct {
			Int int `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Int", "asdf",
			fmt.Errorf(`strconv.ParseInt: parsing "asdf": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default uint",
		F: &struct {
			Uint uint `flag:";-1;"`
		}{},
		ErrBind: ErrorDefaultValue{"Uint", "-1",
			fmt.Errorf(`strconv.ParseUint: parsing "-1": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default uint64",
		F: &struct {
			Uint64 uint64 `flag:";-1;"`
		}{},
		ErrBind: ErrorDefaultValue{"Uint64", "-1",
			fmt.Errorf(`strconv.ParseUint: parsing "-1": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default int64",
		F: &struct {
			Int64 int64 `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Int64", "asdf",
			fmt.Errorf(`strconv.ParseInt: parsing "asdf": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default float64",
		F: &struct {
			Float64 float64 `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Float64", "asdf",
			fmt.Errorf(`strconv.ParseFloat: parsing "asdf": invalid syntax`),
		}.Error(),
	}, {
		Name: "invalid default time.Duration",
		F: &struct {
			Duration time.Duration `flag:";asdf;"`
		}{},
		ErrBind: ErrorDefaultValue{"Duration", "asdf",
			fmt.Errorf(`time: invalid duration asdf`),
		}.Error(),
	},
}
