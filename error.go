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

import "fmt"

// ErrorInvalidType is returned from Bind if v is not a pointer to a struct."
var ErrorInvalidType = fmt.Errorf("v must be a pointer to a struct")

// ErrorInvalidFlagSet is returned from Bind if flg doesn't implement
// STDFlagSet or PFlagSet.
var ErrorInvalidFlagSet = fmt.Errorf("flg must implement STDFlagSet or PFlagSet")

// ErrorDefaultValue is returned from Bind if the <default> value given in the
// tag cannot be parsed and assigned to the field.
type ErrorDefaultValue struct {
	FieldName string
	Value     string
	Err       error
}

// Error implements error.
func (err ErrorDefaultValue) Error() string {
	return fmt.Sprintf("%v: cannot assign default value %q: %v",
		err.FieldName, err.Value, err.Err)
}

// Unwrap implements Unwrap.
func (err ErrorDefaultValue) Unwrap() error {
	return err.Err
}
