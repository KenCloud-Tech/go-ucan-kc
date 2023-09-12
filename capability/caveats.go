package capability

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-ucan-kl/util"
	//go_ucan_kl "go-ucan-kl"
)

type Caveat struct {
	caveat map[string]interface{}
}

func (c *Caveat) enables(other *Caveat) bool {
	if c.isEmpty() {
		return true
	}
	if other.isEmpty() {
		return false
	}

	return c.equalOrContain(other)
}

func (c *Caveat) isEmpty() bool {
	return c.caveat == nil || len(c.caveat) == 0
}

func (c *Caveat) equalOrContain(other *Caveat) bool {
	if c == other {
		return true
	}

	for str, cav := range c.caveat {
		if oCav, ok := other.caveat[str]; !ok {
			return false
		} else {
			// todo:  judgement of equality, not implement now
			if cav != oCav {
				return false
			}
		}
	}
	return true
}

func BuildCaveat(val []byte) (Caveat, error) {
	if ok, jsonBytes := util.IsJson(val); ok {
		if bytes.Equal(jsonBytes, NullJson) {
			return Caveat{}, nil
		}
		mp := make(map[string]interface{})
		err := json.Unmarshal(jsonBytes, &mp)
		if err != nil {
			return Caveat{}, err
		}
		return Caveat{mp}, nil
	}

	return Caveat{}, fmt.Errorf("caveat must be json object, but got %v", val)
}
