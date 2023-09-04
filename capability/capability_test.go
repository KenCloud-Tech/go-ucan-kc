package capability_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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

	capabilityInfos, err := ReduceCapabilities[capability.EmailAddress, capability.EmailAction](chain)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(capabilityInfos), 1)
	assert.Equal(t, capabilityInfos[0].Capability.Resource.ToString(), "mailto:alice@email.com")
	assert.Equal(t, capabilityInfos[0].Capability.Ability.ToString(), "email/send")
}

func TestReportTheFirstIssuerInTheChainAsOriginator(t *testing.T) {
	sendEmailAsBob, err := capability.EmailSemantics.Parse("mailto:bob@email.com", "email/send", nil)
	if err != nil {
		t.Fatal(err)
	}

	leftUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(50).
		WitnessedBy(leftUcan, nil).
		ClaimingCapability(sendEmailAsBob.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	ucanStr, err := ucan.Encode()
	if err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	_, err = store.WriteUcan(leftUcan, nil)
	if err != nil {
		t.Fatal(err)
	}

	pc, err := ProofChainFromUcanStr(ucanStr, nil, store)
	if err != nil {
		t.Fatal(err)
	}

	capInfos, err := ReduceCapabilities[capability.EmailAddress, capability.EmailAction](pc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(capInfos), 1)

	originators := make([]string, 0)
	for ori, ok := range capInfos[0].Originators {
		if !ok {
			panic(fmt.Sprintf("invalid originator %s", ori))
		}
		originators = append(originators, ori)
	}

	assert.Equal(t, originators, []string{fixtures.TestIdentities.BobDidString})
	assert.Equal(t, capInfos[0].Capability.ToCapability(), sendEmailAsBob.ToCapability())
}

func TestFindsTheRightProofChainForTheOriginator(t *testing.T) {
	store := NewMemoryStore()
	sendEmailAsBob, err := capability.EmailSemantics.Parse("mailto:bob@email.com", "email/send", nil)
	if err != nil {
		t.Fatal(err)
	}
	sendEmailAsAlice, err := capability.EmailSemantics.Parse("mailto:alice@email.com", "email/send", nil)
	if err != nil {
		t.Fatal(err)
	}

	leafUcanAlice, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(60).
		ClaimingCapability(sendEmailAsAlice.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	leafUcanBob, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(60).
		ClaimingCapability(sendEmailAsBob.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.MalloryKey).
		ForAudience(fixtures.TestIdentities.AliceDidString).
		WithLifetime(50).
		WitnessedBy(leafUcanBob, nil).
		WitnessedBy(leafUcanAlice, nil).
		ClaimingCapability(sendEmailAsAlice.ToCapability()).
		ClaimingCapability(sendEmailAsBob.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.WriteUcan(leafUcanAlice, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.WriteUcan(leafUcanBob, nil)
	if err != nil {
		t.Fatal(err)
	}

	pc, err := ProofChainFromUcan(ucan, nil, store)
	if err != nil {
		t.Fatal(err)
	}

	capInfos, err := ReduceCapabilities[capability.EmailAddress, capability.EmailAction](pc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(capInfos), 2)
	assert.Equal(t, capInfos[0], &CapabilityInfo{
		Originators: map[string]bool{fixtures.TestIdentities.BobDidString: true},
		NotBefore:   ucan.NotBefore(),
		Expires:     ucan.Expires(),
		Capability:  *sendEmailAsBob,
	})
	assert.Equal(t, capInfos[1], &CapabilityInfo{
		Originators: map[string]bool{fixtures.TestIdentities.AliceDidString: true},
		Capability:  *sendEmailAsAlice,
		NotBefore:   ucan.NotBefore(),
		Expires:     ucan.Expires(),
	})

}
