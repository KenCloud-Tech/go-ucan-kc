package capability

import (
	"fmt"
	"net/url"
	"strconv"
)

var _ Scope = &ProofSelection{}

type ProofSelection struct {
	Index int
}

func (p ProofSelection) ToString() string {
	if p.Index == -1 {
		return "prf:*"
	} else {
		if p.Index < -1 {
			panic(fmt.Sprintf("invalid Index: %d", p.Index))
		}
		return fmt.Sprintf("prf:%d", p.Index)
	}
}

func (p ProofSelection) Contains(other Scope) bool {
	if ps, ok := other.(*ProofSelection); ok {
		return p.Index == ps.Index || p.Index == -1
	} else {
		panic(fmt.Sprintf("invalid comparing between ProofSelection and %T", other))
	}
}

func (p ProofSelection) ParseScope(url url.URL) Scope {
	switch url.Scheme {
	case "prf":
		if url.Path == "*" {
			return &ProofSelection{-1}
		}
		idx, err := strconv.Atoi(url.Path)
		if err != nil {
			panic(err.Error())
		}
		return &ProofSelection{idx}
	default:
		panic(fmt.Sprintf("unsupported schema %s", url.Scheme))
	}
}

var _ Ability = &ProofAction{}

type ProofAction struct {
	str string
}

func (p ProofAction) ParseAbility(str string) Ability {
	if str == "ucan/DELEGATE" {
		return &ProofAction{str: "Delegate"}
	}
	panic(fmt.Sprintf("Unsupported action for proof Resource %s", str))
}

func (p ProofAction) ToString() string {
	return "ucan/DELEGATE"
}
