package capability

import (
	"bytes"
	"fmt"
	"go-ucan-kl/util"
	"net/url"
	"strings"
)

var (
	TypeParseError = fmt.Errorf("type parse error")

	NullJson = []byte("{}")
)

type Scope interface {
	Contains(other Scope) bool
	ParseScope(url url.URL) (Scope, error)
	ToString() string
}

type Ability interface {
	ParseAbility(str string) (Ability, error)
	ToString() string
	Compare(abi Ability) int
}

type ResourceUri struct {
	isScope bool
	scope   Scope
}

func (ru *ResourceUri) ToString() string {
	if ru.isScope {
		return ru.scope.ToString()
	} else {
		return "*"
	}
}

func (ru *ResourceUri) Scope() Scope {
	return ru.scope
}

type CapabilityView struct {
	Resource Resource
	Ability  Ability
	// Caveat must be json bytes
	Caveat []byte
}

func (cv *CapabilityView) Enables(other *CapabilityView) bool {
	caveat, err := BuildCaveat(cv.Caveat)
	if err != nil {
		return false
	}
	otherCaveat, err := BuildCaveat(other.Caveat)
	if err != nil {
		return false
	}

	return cv.Resource.Contains(&other.Resource) &&
		cv.Ability.Compare(other.Ability) >= 0 &&
		caveat.enables(&otherCaveat)
}

func (cv *CapabilityView) ToCapability() *Capability {
	return &Capability{
		Resource: cv.Resource.ToString(),
		Ability:  cv.Ability.ToString(),
		// todo: if bytes json deserialize will get error
		Caveat: string(cv.Caveat),
	}
}

type ResourceType int

const (
	DefaultResource ResourceType = 0
	My                           = 1
	AS                           = 2
)

func (rt ResourceType) Name() string {
	switch rt {
	case DefaultResource:
		return "Resource"
	case My:
		return "My"
	case AS:
		return "As"
	default:
		return fmt.Sprintf("unknown type: %d", rt)
	}
}

type Resource struct {
	ResourceUri
	Type ResourceType
	Did  string
}

func (r *Resource) ToString() string {
	switch r.Type {
	case DefaultResource:
		return r.ResourceUri.ToString()
	case My:
		return fmt.Sprintf("my:%s", r.ResourceUri.ToString())
	case AS:
		return fmt.Sprintf("as:%s:%s", r.Did, r.ResourceUri.ToString())
	default:
		panic(fmt.Sprintf("invalid resource type: %d", r.Type))
	}
}

func (r *Resource) Contains(other *Resource) bool {
	if r.Type != other.Type {
		return false
	}
	switch r.Type {
	case DefaultResource, My:
		return r.ResourceUri.contains(&other.ResourceUri)
	case AS:
		if r.Did != "" && r.Did == other.Did {
			return r.ResourceUri.contains(&other.ResourceUri)
		} else {
			return false
		}
	default:
		panic(fmt.Sprintf("Unsupported Resource type: %s", r.Type.Name()))
	}
}

func (ru *ResourceUri) contains(other *ResourceUri) bool {
	if ru.isScope == false {
		return true
	}
	if other.isScope == false {
		return false
	}

	return ru.scope.Contains(other.scope)
}

var ProofDelegationSemantics = CapabilitySemantics[ProofSelection, ProofAction]{}

type CapabilitySemantics[S Scope, A Ability] struct {
}

func (cs CapabilitySemantics[S, A]) parseResource(resource *url.URL) (ResourceUri, error) {
	switch resource.Path {
	case "*":
		return ResourceUri{
			isScope: false,
		}, nil
	default:
		var scope S
		sc, err := scope.ParseScope(*resource)
		if err != nil {
			return ResourceUri{}, fmt.Errorf("%s : failed to parse resource:%s as %T, err: %v", TypeParseError, resource.String(), scope, err)
		}
		return ResourceUri{
			isScope: true,
			scope:   sc,
		}, nil
	}
}

func (cs CapabilitySemantics[S, A]) extractDid(path string) (string, string, error) {
	pathParts := strings.Split(path, ":")
	if len(pathParts) < 4 {
		return "", "", fmt.Errorf("invalid parts length")
	}

	if pathParts[0] != "did" || pathParts[1] != "key" {
		return "", "", fmt.Errorf("invalid did foramt: %s", path)
	}

	return strings.Join(pathParts[:3], ":"), strings.Join(pathParts[3:], ""), nil
}

func (cs CapabilitySemantics[S, A]) parseCaveat(caveat []byte) []byte {
	//var jsonBytes []byte
	//switch caveat.(type) {
	//case string:
	//	jsonBytes = []byte(caveat.(string))
	//case []byte:
	//	jsonBytes = caveat.([]byte)
	//}
	//if json.Valid(jsonBytes) {
	//	return jsonBytes
	//} else {
	//	return NullJson
	//}

	if caveat == nil || bytes.Equal(caveat, []byte("")) || bytes.Equal(caveat, NullJson) {
		return NullJson
	}
	if util.IsJsonObject(caveat) {
		return caveat
	}
	panic(fmt.Sprintf("%s is not json object", caveat))
}

func (cs CapabilitySemantics[S, A]) Parse(resource string, ability string, caveat []byte) (*CapabilityView, error) {
	uri, err := url.Parse(resource)
	if err != nil {
		return nil, err
	}

	res := &Resource{}
	switch uri.Scheme {
	case "my":
		res.Type = My
		res.ResourceUri, err = cs.parseResource(uri)
		if err != nil {
			return nil, err
		}
	case "as":
		did, resource, err := cs.extractDid(uri.Path)
		if err != nil {
			return nil, err
		}
		res.Type = AS
		res.Did = did
		didUri, err := url.Parse(resource)
		if err != nil {
			return nil, err
		}
		res.ResourceUri, err = cs.parseResource(didUri)
		if err != nil {
			return nil, err
		}
	default:
		res.Type = DefaultResource
		res.ResourceUri, err = cs.parseResource(uri)
		if err != nil {
			return nil, err
		}
	}

	var abi A
	capAbi, err := abi.ParseAbility(ability)
	if err != nil {
		return nil, fmt.Errorf("%s : failed to parse ability:%s as %T, err: %v", TypeParseError, ability, abi, err)
	}

	//capCav, err := cs.parseCaveat(caveat)
	//if err != nil {
	//	return nil, err
	//}
	cv := &CapabilityView{
		Resource: *res,
		Ability:  capAbi,
		Caveat:   cs.parseCaveat(caveat),
	}
	return cv, nil
}

func (cs CapabilitySemantics[S, A]) ParseCapability(cap *Capability) (*CapabilityView, error) {
	return cs.Parse(cap.Resource, cap.Ability, util.CaveatBytes(cap.Caveat))
}
