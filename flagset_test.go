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

package flagbinder

import (
	"bytes"
	"flag"

	"github.com/spf13/pflag"
)

// flagSetUsage is an interface adapter that exposes the Usage func() as a
// method.
type flagSetUsage struct {
	*flag.FlagSet
}

func (flg flagSetUsage) Usage() string {
	usageOutput := bytes.NewBuffer(nil)
	flg.SetOutput(usageOutput)
	flg.FlagSet.Usage()
	return usageOutput.String()
}

// pflagSetUsage is an interface adapter that exposes the Usage func() as a
// method.
type pflagSetUsage struct {
	*pflag.FlagSet
}

func (flg pflagSetUsage) Usage() string {
	// It appears to be a bug but pflag.NewFlagSet doesn't populate the
	// Usage func, but uses internal trickery to avoid panic during Parse.
	// So we fall back to calling this.
	return flg.FlagSet.FlagUsages()
}
