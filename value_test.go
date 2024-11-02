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
	"fmt"
	"strings"
)

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

type TestTextMarshaler struct {
	v   string
	err error
}

func (v *TestTextMarshaler) MarshalText() (text []byte, err error) {
	return []byte(v.v), v.err
}

func (v *TestTextMarshaler) UnmarshalText(text []byte) error {
	v.v = string(text)
	return v.err
}
