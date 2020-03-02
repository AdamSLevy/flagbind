package flagbuilder

import "unicode"

func kebabCase(name string) string {
	var kebab string
	for _, r := range name {
		if unicode.IsUpper(r) {
			if len(kebab) > 0 {
				kebab += "-"
			}
			r = unicode.ToLower(r)
		}
		kebab += string(r)
	}
	return kebab
}
