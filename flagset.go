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
	"flag"
	"time"

	"github.com/spf13/pflag"
)

// FlagSet is an interface satisfied by both *flag.FlagSet and *pflag.FlagSet.
type FlagSet interface {
	Arg(i int) string
	Args() []string
	NArg() int
	NFlag() int
	Set(name, value string) error

	Parse([]string) error

	BoolVar(p *bool, name string, value bool, usage string)
	DurationVar(p *time.Duration, name string, value time.Duration, usage string)
	Float64Var(p *float64, name string, value float64, usage string)
	Int64Var(p *int64, name string, value int64, usage string)
	IntVar(p *int, name string, value int, usage string)
	StringVar(p *string, name string, value string, usage string)
	Uint64Var(p *uint64, name string, value uint64, usage string)
	UintVar(p *uint, name string, value uint, usage string)
}

// STDFlagSet is an interface satisfied by *flag.FlagSet.
type STDFlagSet interface {
	FlagSet
	Lookup(name string) *flag.Flag
	Var(value flag.Value, name string, usage string)
}

// PFlagSet is an interface satisfied by *pflag.FlagSet.
type PFlagSet interface {
	Lookup(name string) *pflag.Flag
	BoolVarP(p *bool, name, short string, value bool, usage string)
	DurationVarP(p *time.Duration, name, short string, value time.Duration, usage string)
	Float64VarP(p *float64, name, short string, value float64, usage string)
	Int64VarP(p *int64, name, short string, value int64, usage string)
	IntVarP(p *int, name, short string, value int, usage string)
	StringVarP(p *string, name, short string, value string, usage string)
	Uint64VarP(p *uint64, name, short string, value uint64, usage string)
	UintVarP(p *uint, name, short string, value uint, usage string)
	VarP(value pflag.Value, name, short string, usage string)
}

// Ensure we are interface compatible with flag and pflag.
var _ FlagSet = &flag.FlagSet{}
var _ STDFlagSet = &flag.FlagSet{}

var _ FlagSet = &pflag.FlagSet{}
var _ PFlagSet = &pflag.FlagSet{}

// pflagValue is a flag.Value -> pflag.Value adapter with a constant Type()
// string.
type pflagValue struct {
	flag.Value
	typeStr string
}

func (val pflagValue) Type() string {
	return val.typeStr
}
