package flagbuilder

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
