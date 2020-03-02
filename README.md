# `package flagbind`

[![GoDoc](https://godoc.org/github.com/AdamSLevy/flagbind?status.svg)](https://godoc.org/github.com/AdamSLevy/flagbind)
[![Build Status](https://travis-ci.org/AdamSLevy/flagbind.svg?branch=master)](https://travis-ci.org/AdamSLevy/flagbind)
[![Coverage Status](https://coveralls.io/repos/github/AdamSLevy/flagbind/badge.svg?branch=master)](https://coveralls.io/github/AdamSLevy/flagbind?branch=master)

Package flagbind parses the exported fields of a struct and binds them to
flags in a `flag.FlagSet` or `pflag.FlagSet`.

`Bind` allows for creating flags declaratively right alongside the definition
of their containing struct. For example, the following stuct could be passed
to Bind to populate a `flag.FlagSet` or `pflag.FlagSet`.

```go
flags := struct {
        StringFlag string `flag:"flag-name;default value;Usage for string-flag"`
        Int        int    `flag:"integer;5"`

        // Flag names default to `auto-kebab-case`
        AutoKebabCase int

        // If pflag is used, -s is be used as the shorthand flag name,
        // otherwise it is ignored for use with the standard flag package.
        ShortName bool `flag:"short,s"`

        // Nested and Embedded structs can add a flag name prefix, or not.
        Nested     StructA
        NestedFlat StructB           `flag:";;;flatten"`
        StructA                      // Flat by default
        StructB    `flag:"embedded"` // Add prefix to nested field flag names.

        // Ignored
        ExplicitlyIgnored bool `flag:"-"`
        unexported        bool
}{
        // Default values may also be set directly to override the tag.
        StringFlag: "override default",
}

fs := pflag.NewFlagSet("", pflag.ContinueOnError)
flagbind.Bind(fs, &flags)
fs.Parse([]string{"--auto-kebab-case"})
```

Bind works seemlessly with both the standard library flag package and the
popular pflag package.

For types that only implement `flag.Value`, Bind wraps them in an adapter so
that they can be used as a `pflag.Value`. The return value of the added
function `Type() string` is the type name of the struct field.

Additional options may be set for each flag. See Bind for the full
documentation details.
