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
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorDefaultValueUnwrap(t *testing.T) {
	err := ErrorDefaultValue{"", "", strconv.ErrSyntax}
	assert.True(t, errors.Is(err, strconv.ErrSyntax))
}
func TestErrorNestedStructUnwrap(t *testing.T) {
	err := newErrorNestedStruct("C", strconv.ErrSyntax)
	assert := assert.New(t)
	assert.EqualError(err, "C: invalid syntax")

	err = newErrorNestedStruct("B", err)
	assert.EqualError(err, "B.C: invalid syntax")

	err = newErrorNestedStruct("A", err)
	assert.EqualError(err, "A.B.C: invalid syntax")
	assert.True(errors.Is(err, strconv.ErrSyntax))
}
