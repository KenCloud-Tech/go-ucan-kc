package util

import "encoding/json"

func IsJson(a interface{}) (bool, []byte) {
	var jsonBytes []byte
	switch a.(type) {
	case string:
		jsonBytes = []byte(a.(string))
	case []byte:
		jsonBytes = a.([]byte)
	default:
		return false, nil
	}

	return json.Valid(jsonBytes), jsonBytes
}
