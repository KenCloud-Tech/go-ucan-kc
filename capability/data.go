package capability

import (
	"encoding/json"
	"fmt"
)

type Capability struct {
	Resource string
	Ability  string
	Caveat   interface{}
}

// Abilities
type Abilities map[string][]interface{}

// Capabilities is a set of all Capabilities
//
// A capability is the association of an "ability" to a "resource": resource x ability x caveats.
// { $RESOURCE: { $ABILITY: [ $CAVEATS ] } }
//
//{
//  "example://example.com/public/photos/": {
//    "crud/read": [{}],
//    "crud/delete": [
//      {
//        "matching": "/(?i)(\\W|^)(baloney|darn|drat|fooey|gosh\\sdarnit|heck)(\\W|$)/"
//      }
//    ]
//  },
//  "example://example.com/private/84MZ7aqwKn7sNiMGsSbaxsEa6EPnQLoKYbXByxNBrCEr": {
//    "wnfs/append": [{}]
//  },
//  "mailto:username@example.com": {
//    "msg/send": [{}],
//    "msg/receive": [
//      {
//        "max_count": 5,
//        "templates": [
//          "newsletter",
//          "marketing"
//        ]
//      }
//    ]
//  }
//}
//

type Capabilities map[string]Abilities

func BuildCapsFromArray(capArray []Capability) (Capabilities, error) {
	caps := make(Capabilities)
	for _, capability := range capArray {
		resourceName := capability.Resource
		ability := capability.Ability
		caveat := capability.Caveat

		var resource Abilities
		if res, ok := caps[resourceName]; !ok {
			resource = make(Abilities)
			caps[resourceName] = resource
		} else {
			resource = res
		}

		// todo not sure, check whether caveat is a json object
		if !json.Valid(caveat.([]byte)) {
			return nil, fmt.Errorf("caveat must be an json object, %v", caveat)
		}

		if _, ok := resource[ability]; ok {
			resource[ability] = append(resource[ability], caveat)
		} else {
			resource[ability] = []interface{}{caveat}
		}
	}

	return caps, nil
}
