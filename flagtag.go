package flagbuilder

import "strings"

type flagTag struct {
	Name  string
	Value string
	Usage string

	Ignored bool
}

func newFlagTag(tag string) (fTag flagTag) {
	if len(tag) == 0 {
		return
	}
	args := strings.Split(tag, ";")
	fTag.Ignored = args[0] == "-" // Ignore this field
	if fTag.Ignored {
		return
	}
	fTag.Name = strings.TrimLeft(args[0], "-")
	if len(args) == 1 {
		return
	}
	fTag.Value = args[1]
	if len(args) == 2 {
		return
	}
	fTag.Usage = args[2]
	return
}
