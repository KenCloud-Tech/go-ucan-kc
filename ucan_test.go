package go_ucan_kl

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	. "go-ucan-kl/capability"
	"go-ucan-kl/test/fixtures"
	"testing"
	"time"
)

func TestRoundTrips(t *testing.T) {
	ucan, err := DefaultBuilder().
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

	err = reUcan.Validate(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUcanTooEarly(t *testing.T) {
	ucan, err := DefaultBuilder().IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithNotBefore(time.Now().Unix() + 30).
		WithLifetime(30).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	if !ucan.isTooEarly(nil) {
		t.FailNow()
	}
}

func TestUcanSerializedJsonAndDeserialized(t *testing.T) {
	sendEmailAsAlice, err := CapabilitySemantics[EmailAddress, EmailAction]{}.Parse("mailto:alice@email.com", "email/send", nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithNotBefore(time.Now().Unix()/1000).
		WithLifetime(30).
		WithFact("abc/challenge", `{"foo":"bar"}`).
		ClaimingCapability(sendEmailAsAlice.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}
	ucanJsonBytes, err := json.Marshal(ucan)
	if err != nil {
		t.Fatal(err.Error())
	}
	reUcan := &Ucan{}
	err = json.Unmarshal(ucanJsonBytes, reUcan)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !ucan.Equals(reUcan) {
		t.FailNow()
	}
}

func TestBaseUcanSerializedAndDeserialized(t *testing.T) {
	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}
	ucanBytes, err := json.Marshal(ucan)
	if err != nil {
		t.Fatal(err.Error())
	}
	reUcan := &Ucan{}
	err = json.Unmarshal(ucanBytes, reUcan)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !ucan.Equals(reUcan) {
		t.FailNow()
	}
}

func TestUcanEqual(t *testing.T) {
	ucanA, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithExpiration(10000000).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	ucanB, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithExpiration(10000000).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	ucanC, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithExpiration(10000001).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	if !ucanA.Equals(ucanB) {
		t.FailNow()
	}
	if ucanA.Equals(ucanC) {
		t.FailNow()
	}
}

func TestUcanLifeTime(t *testing.T) {
	foreverUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	earlyUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(2000).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	laterUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(4000).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, foreverUcan.Expires(), (*int64)(nil))
	assert.True(t, foreverUcan.LifetimeEndsAfter(earlyUcan))
	assert.True(t, !earlyUcan.LifetimeEndsAfter(foreverUcan))
	assert.True(t, laterUcan.LifetimeEndsAfter(earlyUcan))

}
