package flagbind

import "encoding/json"

type JSONRawMessage json.RawMessage

func (data *JSONRawMessage) Set(text string) error {
	return json.Unmarshal([]byte(text), (*json.RawMessage)(data))
}

func (data JSONRawMessage) String() string {
	return string(data)
}

func (data JSONRawMessage) Type() string { return "JSON" }
