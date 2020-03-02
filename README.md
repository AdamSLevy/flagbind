# `package flagbinder`

[![GoDoc](https://godoc.org/github.com/AdamSLevy/flagbinder?status.svg)](https://godoc.org/github.com/AdamSLevy/flagbinder)

Package flagbinder parses the exported fields of a struct and binds them to
flags in a `flag.FlagSet` or `pflag.FlagSet`.

`Bind` allows for creating flags declaratively right alongside the definition
of their containing struct. For example, the following stuct could be passed
to Bind to populate a `flag.FlagSet` or `pflag.FlagSet`.

```go
     flags := type struct {
             StringFlag string `flag:"flag-name;default value;Usage for string-flag"`
             Int        int `flag:"integer;5;Usage for string-flag"`

             // Flag names default to `auto-kebab-case`
             AutoKebabCase int

             // If pflag is used, -s will be used as the shorthand flag
             // name, otherwise it is ignored for use with the standard flag
             // package.
             ShortName bool `flag:"short,s"`

             // Ignored by Bind
             ExplicitlyIgnored bool `flag:"-"`
             unexported        bool
     }{
             // Default values may also be set directly if not already
             // specified.
             ShortName: true,
     }

     fs := pflag.NewFlagSet("", pflag.ContinueOnError)
     flagbinder.Bind(fs, &flags)
     fs.Parse([]string{"--auto-kebab-case"})
```

Bind works seemlessly with both the standard library flag package and the
popular pflag package.

For types that only implement `flag.Value`, Bind wraps them in an adapter so
that they can be used as a `pflag.Value`. The return value of the added
function `Type() string` is the type name of the struct field.

Additional options may be set for each flag. See Bind for the full
documentation details.
