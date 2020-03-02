package flagbuilder

import (
	"flag"

	"github.com/spf13/pflag"
)

// flagSetUsage is an interface adapter that exposes the Usage func() as a
// method.
type flagSetUsage struct {
	*flag.FlagSet
}

func (flg flagSetUsage) Usage() {
	flg.FlagSet.Usage()
}

// pflagSetUsage is an interface adapter that exposes the Usage func() as a
// method.
type pflagSetUsage struct {
	*pflag.FlagSet
}

func (flg pflagSetUsage) Usage() {
	// It appears to be a bug but pflag.NewFlagSet doesn't populate the
	// Usage func, but uses internal trickery to avoid panic during Parse.
	// So we fall back to calling this.
	flg.FlagSet.PrintDefaults()
}
