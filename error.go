package flagbuilder

import "fmt"

var ErrorInvalidType = fmt.Errorf("v must be a pointer to a struct")

type ErrorDefaultValue struct {
	FieldName string
	Value     string
	Err       error
}

func (err ErrorDefaultValue) Error() string {
	return fmt.Sprintf("%v: cannot assign default value %q: %v",
		err.FieldName, err.Value, err.Err)
}
func (err ErrorDefaultValue) Unwrap() error {
	return err.Err
}
