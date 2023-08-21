package go_ucan_kl

import (
	"go-ucan-kl/test/fixtures"
	"testing"
)

func TestRoundTrips(t *testing.T) {
	ucan, err := Default().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).WithLifetime(30).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	err = ucan.Validate(nil)
	if err != nil {
		t.Fatal(err)
	}

	ucanStr, err := ucan.Encode()
	if err != nil {
		t.Fatal(err)
	}

	reUcan, err := DecodeUcanString(ucanStr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(reUcan)
}
