package flagbind

func newBind(opts ...Option) bind {
	var b bind
	for _, opt := range opts {
		opt(&b)
	}
	return b
}

type bind struct {
	Prefix        string
	NoAutoFlatten bool
	CobraFilter   string
}

func (b bind) Option() Option {
	return func(bb *bind) {
		*bb = b
	}
}

func (b bind) IsIgnored(tag flagTag) bool {
	if tag.IsIgnored {
		return true
	}

	switch b.CobraFilter {
	case "persistent":
		return !tag.Persistent
	case "local":
		return !tag.Local
	case "flags":
		fallthrough
	default:
		return !tag.Flags
	}
}

// Option is an option that may be passed to Bind.
type Option func(*bind)

func CobraFilter(name string) Option {
	return func(b *bind) {
		b.CobraFilter = name
	}
}

// Prefix all flag names with prefix, which should include any final separator
// (e.g. 'http-' or 'http.')
func Prefix(prefix string) Option {
	return func(b *bind) {
		b.Prefix = prefix
	}
}

// By default the flags in embedded struct fields are not given a prefix unless
// they explicitly have a `flag:"name"` in their tag.
//
// This overrides this behavior so the flags in embedded struct fields are
// prefixed with their type name unless explicitly flattened with the tag
// `flag:";;;flatten"`.
func NoAutoFlatten() Option {
	return func(b *bind) {
		b.NoAutoFlatten = true
	}
}
