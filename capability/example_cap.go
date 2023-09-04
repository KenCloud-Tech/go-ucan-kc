package capability

import (
	"fmt"
	"net/url"
	"strings"
)

var EmailSemantics = CapabilitySemantics[EmailAddress, EmailAction]{}

var _ Scope = &EmailAddress{}

type EmailAddress struct {
	str string
}

func (e EmailAddress) Contains(other Scope) bool {
	if ea, ok := other.(*EmailAddress); ok {
		return ea.str == e.str
	} else {
		panic(fmt.Sprintf("invalid comparing between EmailAddress and %T", other))
	}
}

func (e EmailAddress) ParseScope(url url.URL) (Scope, error) {
	url.EscapedPath()
	switch url.Scheme {

	case "mailto":
		return &EmailAddress{url.Opaque}, nil
	default:
		return nil, fmt.Errorf("Could not interpret URI as an email address: %s", url.String())
	}
}

func (e EmailAddress) ToString() string {
	return fmt.Sprintf("mailto:%s", e.str)
}

var _ Ability = &EmailAction{}

type EmailAction struct{}

func (e EmailAction) Compare(abi Ability) int {
	if _, ok := abi.(*EmailAction); ok {
		return 0
	}
	panic(fmt.Sprintf("comparing between different ability: %t and %t", e, abi))
}

func (e EmailAction) ParseAbility(str string) (Ability, error) {
	if str == "email/send" {
		return &EmailAction{}, nil
	}
	return nil, fmt.Errorf("Unrecognized action: %s", str)
}

func (e EmailAction) ToString() string {
	return "email/send"
}

var WNFSSemantics = CapabilitySemantics[WNFSScope, WNFSCapLevel]{}

var _ Scope = &WNFSScope{}

type WNFSScope struct {
	origin string
	path   string
}

func (w WNFSScope) Contains(other Scope) bool {
	if otherWNFS, ok := other.(*WNFSScope); !ok {
		return false
	} else {
		if otherWNFS.origin != w.origin {
			return false
		}
		pathParts := strings.Split(w.path, "/")
		otherPathParts := strings.Split(otherWNFS.path, "/")

		for _, part := range pathParts {
			if len(otherPathParts) == 0 {
				return false
			}
			otherPart := otherPathParts[0]
			otherPathParts = otherPathParts[1:]

			if part != otherPart {
				return false
			}
		}
	}

	return true
}

func (w WNFSScope) ParseScope(url url.URL) (Scope, error) {
	sch, host, path := url.Scheme, url.Host, url.Path
	if sch != "wnfs" {
		return nil, fmt.Errorf("cannot interpret URI as WNFS scope: %s", url.String())
	}
	return &WNFSScope{
		origin: host,
		path:   path,
	}, nil
}

func (w WNFSScope) ToString() string {
	return fmt.Sprintf("wnfs://%s%s", w.origin, w.path)
}

var _ Ability = &WNFSCapLevel{}

type Level string

const (
	Create     Level = "wnfs/create"
	Revise     Level = "wnfs/revise"
	SoftDelete Level = "wnfs/soft_delete"
	OverWrite  Level = "wnfs/overwrite"
	SuperUser  Level = "wnfs/super_user"
)

//var levelSet = []Level{Create, Revise, SoftDelete, OverWrite, SuperUser}

var levelMap = map[Level]int{
	Create:     0,
	Revise:     1,
	SoftDelete: 2,
	OverWrite:  3,
	SuperUser:  4,
}

type WNFSCapLevel struct {
	level Level
}

func (w WNFSCapLevel) ParseAbility(str string) (Ability, error) {
	if _, exist := levelMap[Level(str)]; exist {
		return &WNFSCapLevel{Level(str)}, nil
	}
	return nil, fmt.Errorf("no such WNFS capability level: %s", str)
}

func (w WNFSCapLevel) ToString() string {
	if _, exist := levelMap[w.level]; exist {
		return string(w.level)
	}
	panic(fmt.Sprintf("invalid WNFSCapLevel: %s", w.level))
}

func (w WNFSCapLevel) Compare(abi Ability) int {
	var otherWeight = -1
	//var ok bool
	if otherWNFS, ok := abi.(*WNFSCapLevel); !ok {
		panic(fmt.Sprintf("comparing between different ability: %T and %T", w, abi))
	} else {
		if otherWeight, ok = levelMap[otherWNFS.level]; !ok {
			panic(fmt.Sprintf("invalid WNFSCapLevel: %s", otherWNFS.level))
		}
	}

	if weight, exist := levelMap[w.level]; !exist {
		panic(fmt.Sprintf("invalid WNFSCapLevel: %s", w.level))
	} else {
		if weight == otherWeight {
			return 0
		} else if weight > otherWeight {
			return 1
		} else {
			return -1
		}
	}
}
