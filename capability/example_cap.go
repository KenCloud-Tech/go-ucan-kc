package capability

import (
	"fmt"
	"net/url"
)

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

func (e EmailAddress) ParseScope(url url.URL) Scope {
	switch url.Scheme {
	case "mailto":
		return &EmailAddress{url.Path}
	default:
		panic(fmt.Sprintf("Could not interpret URI as an email address: %s", url.String()))
	}
}

func (e EmailAddress) ToString() string {
	return fmt.Sprintf("mailto:%s", e.str)
}

var _ Ability = &EmailAction{}

type EmailAction struct{}

func (e EmailAction) ParseAbility(str string) Ability {
	if str == "email/send" {
		return &EmailAction{}
	}
	panic(fmt.Sprintf("Unrecognized action: %s", str))
}

func (e EmailAction) ToString() string {
	return "email/send"
}