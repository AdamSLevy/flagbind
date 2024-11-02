package flagbind

import (
	"encoding"
	"encoding/json"
)

type JSONRawMessage json.RawMessage

func (data *JSONRawMessage) Set(text string) error {
	return json.Unmarshal([]byte(text), (*json.RawMessage)(data))
}

func (data JSONRawMessage) String() string {
	return string(data)
}

func (data JSONRawMessage) Type() string { return "JSON" }

type pflagMarshalerValue struct {
	marshaler textBidiMarshaler
	typeStr   string
}

type textBidiMarshaler interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

func (val pflagMarshalerValue) String() string {
	text, err := val.marshaler.MarshalText()
	if err != nil {
		return "<invalid>"
	}
	return string(text)
}

func (val pflagMarshalerValue) Type() string {
	return val.typeStr
}

func (val *pflagMarshalerValue) Set(v string) error {
	return val.marshaler.UnmarshalText([]byte(v))
}
