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

func TestReportsAllChainOptions(t *testing.T) {
	store := NewMemoryStore()
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
		ClaimingCapability(sendEmailAsAlice.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.MalloryKey).
		ForAudience(fixtures.TestIdentities.AliceDidString).
		WithLifetime(40).
		ClaimingCapability(sendEmailAsAlice.ToCapability()).
		WitnessedBy(leafUcanBob, nil).
		WitnessedBy(leafUcanAlice, nil).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	ucanStr, err := ucan.Encode()
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

	pc, err := ProofChainFromUcanStr(ucanStr, nil, store)
	if err != nil {
		t.Fatal(err)
	}

	capInfos, err := ReduceCapabilities[capability.EmailAddress, capability.EmailAction](pc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(capInfos), 1)

	capInfo := capInfos[0]
	assert.Equal(t, &CapabilityInfo{
		Originators: map[string]bool{fixtures.TestIdentities.AliceDidString: true, fixtures.TestIdentities.BobDidString: true},
		NotBefore:   ucan.NotBefore(),
		Expires:     ucan.Expires(),
		Capability:  *sendEmailAsAlice,
	}, capInfo)

}

func TestValidatesCaveats(t *testing.T) {
	resource := "mailto:alice@email.com"
	ability := "email/send"
	noCaveat := capability.NewCapability(resource, ability, "{}")
	xCaveat := capability.NewCapability(resource, ability, `{"x":true}`)
	yCaveat := capability.NewCapability(resource, ability, `{"y":true}`)
	zCaveat := capability.NewCapability(resource, ability, `{"z":true}`)
	yzCaveat := capability.NewCapability(resource, ability, `{"y":true, "z": true}`)

	//valid := make([][2][]capability.Capability, 0)
	valid := [][][]*capability.Capability{
		{{noCaveat}, {noCaveat}},
		{{xCaveat}, {xCaveat}},
		{{noCaveat}, {xCaveat}},
		{{xCaveat, yCaveat}, {xCaveat}},
		{{xCaveat, yCaveat}, {xCaveat, yzCaveat}},
	}

	invalid := [][][]*capability.Capability{
		{{xCaveat}, {noCaveat}},
		{{xCaveat}, {yCaveat}},
		{{xCaveat, yCaveat}, {xCaveat, yCaveat, zCaveat}},
	}

	for _, caps := range valid {
		successful := testCapabilitiesDelegation(t, caps[0], caps[1])
		assert.True(t, successful)
	}

	for _, caps := range invalid {
		successful := testCapabilitiesDelegation(t, caps[0], caps[1])
		assert.True(t, !successful)
	}
}

func testCapabilitiesDelegation(t *testing.T, proofCapabilities []*capability.Capability, delegatedCapabilities []*capability.Capability) bool {
	store := NewMemoryStore()

	proofUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(60).
		ClaimingCapabilities(proofCapabilities).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.MalloryKey).
		ForAudience(fixtures.TestIdentities.AliceDidString).
		WithLifetime(50).
		WitnessedBy(proofUcan, nil).
		ClaimingCapabilities(delegatedCapabilities).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.WriteUcan(proofUcan, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.WriteUcan(ucan, nil)
	if err != nil {
		t.Fatal(err)
	}

	pc, err := ProofChainFromUcan(ucan, nil, store)
	if err != nil {
		t.Fatal(err)
	}

	return enablesCapabilities(t, pc, fixtures.TestIdentities.AliceDidString, delegatedCapabilities)
}

func enablesCapabilities(t *testing.T, pc *ProofChain, ori string, desiredCaps []*capability.Capability) bool {
	capInfos, err := ReduceCapabilities[capability.EmailAddress, capability.EmailAction](pc)
	if err != nil {
		t.Fatal(err)
	}

	for _, desiredCap := range desiredCaps {
		hasCap := false
		capView, err := capability.EmailSemantics.ParseCapability(desiredCap)
		if err != nil {
			t.Fatal(err)
		}
		for _, info := range capInfos {
			if info.Originators[ori] && info.Capability.Enables(capView) {
				hasCap = true
				break
			}
		}
		if !hasCap {
			return false
		}
	}
	return true
}
