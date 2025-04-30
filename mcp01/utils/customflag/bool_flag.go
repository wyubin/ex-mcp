package customflag

import (
	"database/sql"
	"fmt"
)

type FlagBool sql.NullBool

// implement flag Value
func (e FlagBool) String() string {
	if !e.Valid {
		return ""
	}
	return fmt.Sprint(e.Bool)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e FlagBool) Set(val string) error {
	choice := map[string]struct{}{
		"true":  {},
		"false": {},
	}
	if _, found := choice[val]; !found {
		return fmt.Errorf("%s must be true or false", e.Type())
	}
	if val == "true" {
		e.Bool = true
	} else {
		e.Bool = false
	}
	e.Valid = true
	return nil
}

// Type is only used in help text
func (e FlagBool) Type() string {
	return "FlagBool"
}
