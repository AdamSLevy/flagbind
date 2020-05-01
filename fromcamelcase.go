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
	"unicode"
)

// FromCamelCase converts CamelCase to kebab-case, or snake_case, or lowercase,
// depending on `sep`.
//
// It makes a best effort at respecting capitalized acronyms. For example:
//
//      camel -> camel
//      CamelCamel -> camel-camel
//      CamelID -> camel-id
//      IDCamel -> id-camel
//      APICamel -> api-camel
//      APIURL -> apiurl
//      ApiUrl -> api-url
//      APIUrlID -> api-url-id
func FromCamelCase(name, sep string) string {

	var kebab string
	var acronym string
	for _, r := range name {

		if unicode.IsUpper(r) {
			acronym += string(unicode.ToLower(r))
			continue
		}

		if len(acronym) > 1 {

			if kebab != "" {
				kebab += sep
			}

			// omit last char
			kebab += acronym[:len(acronym)-1]

			// add last char after separator
			kebab += sep + acronym[len(acronym)-1:]

			acronym = ""

		} else if acronym != "" {

			if kebab != "" {
				kebab += sep
			}

			kebab += acronym
			acronym = ""

		}

		kebab += string(r)
	}

	if kebab != "" && acronym != "" {
		kebab += sep
	}

	return kebab + acronym
}
