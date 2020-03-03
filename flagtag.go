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

// flagTag represents all possible tag settings.
type flagTag struct {
	ExplicitName bool
	Name         string
	ShortName    string
	Value        string
	Usage        string
	UsageRef     bool

	Ignored bool

	HideDefault bool
	Hidden      bool

	Flatten bool
}

// newFlagTag parses all possible tag settings.
func newFlagTag(tag string) (fTag flagTag) {
	if len(tag) == 0 {
		return
	}
	args := strings.Split(tag, ";")
	fTag.Ignored = args[0] == "-"
	if fTag.Ignored {
		return
	}

	fTag.parseNames(args[0])
	if len(args) == 1 {
		return
	}

	fTag.Value = args[1]
	if len(args) == 2 {
		return
	}

	fTag.Usage = args[2]
	fTag.UsageRef = fTag.Usage == "_"
	if len(args) == 3 {
		return
	}

	fTag.parseOptions(args[3])
	return
}

// parseNames parses and sorts long and short flag names
func (fTag *flagTag) parseNames(name string) {
	defer func() {
		if len(fTag.Name) < len(fTag.ShortName) { // ensure Name is longer
			fTag.Name, fTag.ShortName = fTag.ShortName, fTag.Name
		}
		if len(fTag.Name) == 1 {
			// If Name qualifies as short, override ShortName.
			fTag.ShortName = fTag.Name
		} else if len(fTag.ShortName) > 1 {
			// Short name is too long, so censor it.
			fTag.ShortName = ""
		}
		fTag.ExplicitName = fTag.Name != ""
	}()
	names := strings.Split(name, ",")
	fTag.Name = strings.TrimLeft(names[0], "-")
	if len(names) == 1 {
		return
	}
	fTag.ShortName = strings.TrimLeft(names[1], "-")
	return
}

func (fTag *flagTag) parseOptions(opts string) {
	opts = strings.ToLower(opts)
	fTag.Hidden = strings.Contains(opts, "hidden")
	fTag.HideDefault = strings.Contains(opts, "hide-default")
	fTag.Flatten = strings.Contains(opts, "flatten")
}
