package customflag

import (
	"errors"
	"fmt"
	"os"
)

type FlagPath string

// implement flag Value
func (e *FlagPath) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *FlagPath) Set(val string) error {
	if _, err := os.Stat(val); os.IsNotExist(err) {
		fmt.Printf("err: %s\n", err.Error())
		return errors.New(fmt.Sprintf("path[%s] do not exist", val))
	} else {
		*e = FlagPath(val)
		return nil
	}
}

// Type is only used in help text
func (e *FlagPath) Type() string {
	return "FlagPath"
}
