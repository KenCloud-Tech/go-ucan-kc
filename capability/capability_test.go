package capability_test

import (
	. "go-ucan-kl"
	"go-ucan-kl/capability"
	"go-ucan-kl/test/fixtures"
	"testing"
)

func TestSimpleExample(t *testing.T) {
	sendEmailAsAlice, err := capability.EmailSemantics.Parse("mailto:alice@email.com", "email/send", nil)
	if err != nil {
		t.Fatal(err)
	}

	leafUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		ClaimingCapability(sendEmailAsAlice.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	attenuatedUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(50).
		WitnessedBy(leafUcan, nil).
		ClaimingCapability(sendEmailAsAlice.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	_, err = store.WriteUcan(leafUcan, nil)
	if err != nil {
		t.Fatal(err)
	}

	chain, err := ProofChainFromUcan(attenuatedUcan, nil, store)
	if err != nil {
		t.Fatal(err)
	}

	_, err = chain.ReduceCapabilities((*capability.CapabilitySemantics[capability.Scope, capability.Ability])(&capability.EmailSemantics))
	if err != nil {
		t.Fatal(err)
	}
}
