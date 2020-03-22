package flagbind

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FromCamelCaseTest struct {
	Camel string
	Kebab string
}

var fromCamelCaseTests = []FromCamelCaseTest{
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

func TestFromCamelCase(t *testing.T) {
	for _, test := range fromCamelCaseTests {
		for _, sep := range []string{"-", ".", "_", ""} {
			t.Run("sep/"+sep+"/"+test.Camel, func(t *testing.T) {
				assert.Equal(t,
					strings.Replace(test.Kebab, "-", sep, -1),
					FromCamelCase(test.Camel, sep),
					test.Camel)
			})
		}
	}
}
