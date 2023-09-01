package go_ucan_kl_test

import (
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
	. "go-ucan-kl"
	"go-ucan-kl/capability"
	"go-ucan-kl/test/fixtures"
	"testing"
	"time"
)

func TestBuildsSimpleRoundTrip(t *testing.T) {
	factOne := `{"test":true}`
	factTwo := `{"preimage":"abc", "hash":"sth"}`

	expectedFactsMap := map[string]interface{}{
		"abc/challenge": factOne,
		"def/challenge": factTwo,
	}

	capA, err := capability.EmailSemantics.Parse("mailto:alice@gmail.com", "email/send", nil)
	if err != nil {
		t.Fatal(err)
	}
	capB, err := capability.WNFSSemantics.Parse("wnfs://alice.fission.name/public", "wnfs/super_user", nil)
	if err != nil {
		t.Fatal(err)
	}

	expectedCaps, err := capability.BuildCapsFromArray([]capability.Capability{*capA.ToCapability(), *capB.ToCapability()})
	if err != nil {
		t.Fatal(err)
	}

	exp := time.Now().Add(time.Second * 30)
	nbf := time.Now().Add(-time.Second * 30)

	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithExpiration(exp.Unix()).
		WithNotBefore(nbf.Unix()).
		WithFact("abc/challenge", factOne).
		WithFact("def/challenge", factTwo).
		ClaimingCapability(capA.ToCapability()).
		ClaimingCapability(capB.ToCapability()).
		WithNonce().
		Build()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, ucan.Issuer(), fixtures.TestIdentities.AliceDidString)
	assert.Equal(t, ucan.Audience(), fixtures.TestIdentities.BobDidString)
	assert.Equal(t, *(ucan.Expires()), exp.Unix())
	assert.Equal(t, *(ucan.NotBefore()), nbf.Unix())
	assert.Equal(t, ucan.Facts(), expectedFactsMap)
	assert.Equal(t, ucan.Capabilities(), expectedCaps)
	assert.True(t, ucan.Nonce() != "")
}

func TestBuildsWithLiftetimeInSeconds(t *testing.T) {
	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(100).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, *ucan.Expires() > time.Now().Add(time.Second*90).Unix())
}

func TestPreventsDuplicateProofs(t *testing.T) {
	parentCap, err := capability.WNFSSemantics.Parse("wnfs://alice.fission.name/public", "wnfs/super_user", nil)
	if err != nil {
		t.Fatal(err)
	}

	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(30).
		ClaimingCapability(parentCap.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	attenuatedCapOne, err := capability.WNFSSemantics.Parse("wnfs://alice.fission.name/public/Apps", "wnfs/create", nil)
	if err != nil {
		t.Fatal(err)
	}
	attenuatedCapTwo, err := capability.WNFSSemantics.Parse("wnfs://alice.fission.name/public/Domains", "wnfs/create", nil)
	if err != nil {
		t.Fatal(err)
	}

	nextUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(30).
		WitnessedBy(ucan, nil).
		ClaimingCapability(attenuatedCapOne.ToCapability()).
		ClaimingCapability(attenuatedCapTwo.ToCapability()).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	ucanCid, _, err := ucan.ToCid(nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, ucanCid.String(), nextUcan.Proofs()[0])
}

func TestCustomPrefix(t *testing.T) {
	prefix := &cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.MURMUR3X64_64,
		MhLength: -1,
	}
	leafUcan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		WithLifetime(60).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	delegatedToken, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.BobKey).
		ForAudience(fixtures.TestIdentities.MalloryDidString).
		WithLifetime(50).
		WitnessedBy(leafUcan, prefix).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	store := NewMemoryStore()
	_, err = store.WriteUcan(leafUcan, prefix)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.WriteUcan(delegatedToken, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ProofChainFromUcan(delegatedToken, nil, store)
	if err != nil {
		t.Fatal(err)
	}

}

//func TestMalloryKey(t *testing.T) {
//	pri, _, err := crypto.GenerateRSAKeyPair(2048, rand.Reader)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	priBytes, err := crypto.MarshalPrivateKey(pri)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	resBytes := make([]byte, base64.StdEncoding.EncodedLen(len(priBytes)))
//	base64.StdEncoding.Encode(resBytes, priBytes)
//	t.Logf("%s", string(resBytes))
//
//	dp, err := didkey.NewDidKeyPairFromPrivateKeyString(string(resBytes))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Logf("%#v", dp)
//}
