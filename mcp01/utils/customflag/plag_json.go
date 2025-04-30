package customflag

import (
	"encoding/json"
	"fmt"
)

type FlagJsonMap map[string]interface{}

func (s *FlagJsonMap) String() string {
	return fmt.Sprint(*s)
}

func (s *FlagJsonMap) Set(val string) error {
	res := make(FlagJsonMap)
	err := json.Unmarshal([]byte(val), &res)
	if err != nil {
		return fmt.Errorf("FlagJsonMap error -> %w", err)
	}
	*s = res
	return nil
}

func (s *FlagJsonMap) Type() string {
	return "FlagJsonMap"
}

type FlagJsonSliceMap []FlagJsonMap

func (s *FlagJsonSliceMap) String() string {
	return fmt.Sprint(*s)
}

func (s *FlagJsonSliceMap) Set(val string) error {
	res := make(FlagJsonSliceMap, 0)
	err := json.Unmarshal([]byte(val), &res)
	if err != nil {
		return fmt.Errorf("FlagJsonSliceMap error -> %w", err)
	}
	*s = res
	return nil
}

func (s *FlagJsonSliceMap) Type() string {
	return "FlagJsonMap"
}
