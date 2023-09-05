package capability

import (
	"fmt"
	"go-ucan-kl/util"
)

type Capability struct {
	Resource string
	Ability  string
	Caveat   interface{}
}

func NewCapability(resource string, ability string, caveat interface{}) *Capability {
	return &Capability{
		Resource: resource,
		Ability:  ability,
		Caveat:   caveat,
	}
}

// Abilities
type Abilities map[string][]interface{}

// Capabilities is a set of all Capabilities
//
// A capability is the association of an "Ability" to a "Resource": Resource x Ability x caveats.
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

func (caps Capabilities) ToCapsArray() []Capability {
	capArray := make([]Capability, 0)
	for res, abi := range caps {
		for abiName, cavs := range abi {
			if len(cavs) == 0 {
				continue
			}
			for _, cav := range cavs {
				cap := Capability{
					Resource: res,
					Ability:  abiName,
					Caveat:   cav,
				}
				capArray = append(capArray, cap)
			}
		}
	}
	return capArray
}

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

		// todo not sure, check whether Caveat is a json object
		//if !json.Valid(caveat.([]byte)) {
		//	return nil, fmt.Errorf("Caveat must be an json object, %v", caveat)
		//}
		if ok, _ := util.IsJson(caveat); !ok {
			return nil, fmt.Errorf("caveat must be an json object, but got: %v", caveat)
		}

		if _, ok := resource[ability]; ok {
			resource[ability] = append(resource[ability], caveat)
		} else {
			resource[ability] = []interface{}{caveat}
		}
	}

	return caps, nil
}
