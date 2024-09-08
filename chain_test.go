package ucan

import (
	"github.com/KenCloud-Tech/go-ucan-kc/test/fixtures"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestDecodesDeepUcanChains(t *testing.T) {
	leafUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	delegatedUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(50).
		WitnessedBy(leafUcan, nil).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	_, err = store.WriteUcan(leafUcan, nil)
	if err != nil {
		t.Fatal(err)
	}

	delegatedUcanStr, err := delegatedUcan.Encode()
	if err != nil {
		t.Fatal(err)
	}

	chain, err := ProofChainFromUcanStr(delegatedUcanStr, nil, store)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, chain.ucan.Audience(), fixtures.TestIdentities.MalloryDidString)
	assert.Equal(t, chain.proofs[0].ucan.Issuer(), fixtures.TestIdentities.AliceDidString)
}

func TestFailedWithIncorrectChaining(t *testing.T) {
	leafUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	delegatedToken, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(50).
		WitnessedBy(leafUcan, nil).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	delegatedTokenStr, err := delegatedToken.Encode()
	if err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	_, err = store.WriteUcan(leafUcan, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ProofChainFromUcanStr(delegatedTokenStr, nil, store)
	if err == nil {
		t.FailNow()
	}

	assert.True(t, strings.Contains(err.Error(), "Invalid UCAN link: audience"))
}

func TestChainFromUcanCid(t *testing.T) {
	leafUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	delegatedUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(50).
		WitnessedBy(leafUcan, nil).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	_, err = store.WriteUcan(leafUcan, nil)
	if err != nil {
		t.Fatal(err)
	}

	c, err := store.WriteUcan(delegatedUcan, nil)
	if err != nil {
		t.Fatal(err)
	}

	chain, err := ProofChainFromUcanCid(c, nil, store)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, chain.ucan.Audience(), fixtures.TestIdentities.MalloryDidString)
	assert.Equal(t, chain.proofs[0].ucan.Issuer(), fixtures.TestIdentities.AliceDidString)
}

func TestHandleMultiLeaves(t *testing.T) {
	leafUcanOne, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	leafUcanTwo, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.MalloryKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	delegatedUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.AliceDidString).
		WithLifetime(50).
		WitnessedBy(leafUcanOne, nil).
		WitnessedBy(leafUcanTwo, nil).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	delegatedUcanStr, err := delegatedUcan.Encode()
	if err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	_, err = store.WriteUcan(leafUcanOne, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.WriteUcan(leafUcanTwo, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ProofChainFromUcanStr(delegatedUcanStr, nil, store)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUseCustomTimestampToValidateUcan(t *testing.T) {
	leafUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	leafUcanStr, err := leafUcan.Encode()
	if err != nil {
		t.Fatal(err)
	}

	delegatedUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(50).
		WitnessedBy(leafUcan, nil).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	_, err = store.WriteUcanStr(leafUcanStr, nil)
	if err != nil {
		t.Fatal(err)
	}

	c, err := store.WriteUcan(delegatedUcan, nil)
	if err != nil {
		t.Fatal(err)
	}

	nowTime := time.Now()
	_, err = ProofChainFromUcanCid(c, &nowTime, store)
	if err != nil {
		t.Fatal(err)
	}

	invalidTime := nowTime.Add(time.Second * 51)
	_, err = ProofChainFromUcanCid(c, &invalidTime, store)
	if err == nil {
		t.FailNow()
	}
	assert.Equal(t, err, UcanExpiredError)
}
