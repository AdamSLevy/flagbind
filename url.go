package flagbind

import "net/url"

type URL url.URL

func (u *URL) Set(text string) error {
	_u, err := url.Parse(text)
	if err != nil {
		return err
	}
	*u = (URL)(*_u)
	return nil
}

func (u URL) String() string {
	return (*url.URL)(&u).String()
}

func (u URL) Type() string { return "URL" }
