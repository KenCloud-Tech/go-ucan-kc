package util

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
)

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

func IsJsonObject(val []byte) bool {
	js, err := simplejson.NewJson(val)
	if err != nil {
		return false
	}
	if _, err = js.Map(); err != nil {
		return false
	}
	return true
}

func CaveatBytes(caveat interface{}) []byte {
	switch caveat.(type) {
	case string:
		return []byte(caveat.(string))
	case []byte:
		return caveat.([]byte)
	default:
		panic(fmt.Sprintf("can not convert caveat: %v to bytes", caveat))
	}
}
