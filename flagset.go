package flagbuilder

import (
	"flag"
	"time"

	"github.com/spf13/pflag"
)

// FlagSet is an interface satisfied by both *flag.FlagSet and *pflag.FlagSet.
type FlagSet interface {
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
	Var(value flag.Value, name string, usage string)
}

// PFlagSet is an interface satisfied by *pflag.FlagSet.
type PFlagSet interface {
	FlagSet
	Var(value pflag.Value, name string, usage string)
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
