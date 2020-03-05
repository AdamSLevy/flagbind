package flagbind

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type KebabTest struct {
	Camel string
	Kebab string
}

var kebabTests = []KebabTest{
	{"ID", "id"},
	{"NewID", "new-id"},
	{"FAAddress", "fa-address"},
	{"URL", "url"},
	{"ServerURL", "server-url"},
	{"APIServerURL", "api-server-url"},
	{"APIUrlID", "api-url-id"},
	{"AutoKebab", "auto-kebab"},
	{"StructABool", "struct-a-bool"},
}

func TestKebab(t *testing.T) {
	for _, test := range kebabTests {
		t.Run(test.Camel, func(t *testing.T) {
			assert.Equal(t, test.Kebab, CamelToKebabCase(test.Camel), test.Camel)
		})
	}
}
