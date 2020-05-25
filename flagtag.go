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
	"strings"
)

type flagTag struct {
	// `flag:"<name>,<short name>"`
	// Number int `flag:"num,n"`
	Name            string
	ShortName       string
	HasExplicitName bool
	IsIgnored       bool

	// `flag:";<default value>"`
	// Number int `flag:";5"`
	DefValue string

	// `flag:";;<usage>"`
	// Number int `flag:";;Number of times to do"`
	// _ struct{} `use:"something"`
	Usage string

	// Options
	// `flag:";;;<options>"`

	// Number int `flag:";;;hide-default,hidden"`
	HideDefault bool // `flag:";;;hide-default"`
	Hidden      bool // `flag:";;;hidden"`

	// Nested struct
	Flatten bool // `flag:";;;flatten"`
}

// newFlagTag parses all possible tag settings.
func newFlagTag(tag string) (fTag flagTag) {
	if tag == "" {
		return
	}
	args := strings.Split(tag, ";")
	fTag.IsIgnored = args[0] == "-"
	if fTag.IsIgnored {
		return
	}

	fTag.parseNames(args[0])
	if len(args) == 1 {
		return
	}

	fTag.DefValue = args[1]
	if len(args) == 2 {
		return
	}

	fTag.Usage = strings.TrimSpace(args[2])
	if len(args) == 3 {
		return
	}

	fTag.parseOptions(args[3])
	return
}

// parseNames parses and sorts the long and short flag names.
func (fTag *flagTag) parseNames(name string) {

	names := strings.Split(name, ",")

	fTag.Name = strings.TrimLeft(names[0], "-")
	if len(names) > 1 {
		fTag.ShortName = strings.TrimLeft(names[1], "-")
	}

	// Ensure Name is longer than ShortName.
	if len(fTag.Name) < len(fTag.ShortName) {
		fTag.Name, fTag.ShortName = fTag.ShortName, fTag.Name
	}

	// If short name is too long, censor it.
	if len(fTag.ShortName) > 1 {
		fTag.ShortName = ""
	}

	// If Name qualifies as short, override ShortName.
	if len(fTag.Name) == 1 {
		fTag.ShortName = fTag.Name
	}

	fTag.HasExplicitName = fTag.Name != ""
}

// parseOptions parses the hidden, hide-default, and flatten options.
func (fTag *flagTag) parseOptions(opts string) {
	opts = strings.ToLower(opts)
	fTag.Hidden = strings.Contains(opts, "hidden")
	fTag.HideDefault = strings.Contains(opts, "hide-default")
	fTag.Flatten = strings.Contains(opts, "flatten")
}
