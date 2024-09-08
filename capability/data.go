package capability

import (
	"encoding/json"
	"fmt"
	"github.com/KenCloud-Tech/go-ucan-kc/util"
)

type Capability struct {
	Resource string
	Ability  string
	Caveat   interface{}
}

func NewCapability(resource string, ability string, caveat []byte) *Capability {
	isJson, jsonBytes := util.IsJson(caveat)
	if !isJson {
		panic(fmt.Sprintf("caveat must be json object, but got: %v", caveat))
	}
	return &Capability{
		Resource: resource,
		Ability:  ability,
		// todo: if cav is not string(json object), get error while ucan unmarshal
		Caveat: string(jsonBytes),
	}
}

// Abilities
type Abilities map[string][]interface{}

//type Caveat map[string]interface{}

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

func BuildCapsFromJsonBytes(val []byte) (Capabilities, error) {
	caps := make(Capabilities)
	err := json.Unmarshal(val, &caps)
	if err != nil {
		return nil, err
	}

	// validate and set
	for res, abilities := range caps {
		if abilities == nil || len(abilities) == 0 {
			return nil, fmt.Errorf("resource must have at least one ability")
		}
		for abiName, cavs := range abilities {
			if cavs == nil || len(cavs) == 0 {
				if len(abilities) == 1 {
					// invalid res
					delete(caps, res)
				} else {
					// invalid ability
					delete(abilities, abiName)
				}
				continue
			}
			for idx, cav := range cavs {
				// todo: if cav is not string(json object), get error while ucan unmarshal
				switch cav.(type) {
				case string:
				case []byte:
					cavs[idx] = string(cav.([]byte))
				default:
					cavBytes, err := json.Marshal(cav)
					if err != nil {
						return nil, fmt.Errorf("invalid caveat: %v", cav)
					}
					if json.Valid(cavBytes) {
						cavs[idx] = string(cavBytes)
					} else {
						return nil, fmt.Errorf("caveat must be json object")
					}
				}
				if !util.IsJsonObject([]byte(cavs[idx].(string))) {
					return nil, fmt.Errorf("caveat should be json object, but got: %v", cavs[idx])
				}
			}
		}
	}
	return caps, nil
}
