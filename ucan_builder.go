package go_ucan_kl

import (
	"fmt"
	mb "github.com/multiformats/go-multibase"
	"go-ucan-kl/capability"
	didkey "go-ucan-kl/key"
	"math/rand"
	"time"
)

// todo: just for test
var randSource = rand.New(rand.NewSource(time.Now().Unix()))

type UcanBuilder struct {
	issuer   didkey.KeyMaterial
	audience string

	capabilities []capability.Capability
	lifetime     uint64
	expiration   int64
	notBefore    int64

	facts    map[string]interface{}
	proofs   []string
	addNonce bool
}

func Default() *UcanBuilder {
	return &UcanBuilder{
		capabilities: make([]capability.Capability, 0),
		facts:        make(map[string]interface{}),
		proofs:       make([]string, 0),
	}
}

func (ub *UcanBuilder) IssuedBy(issuer didkey.KeyMaterial) *UcanBuilder {
	ub.issuer = issuer
	return ub
}

func (ub *UcanBuilder) ForAudience(audience string) *UcanBuilder {
	ub.audience = audience
	return ub
}

func (ub *UcanBuilder) WithLifetime(seconds uint64) *UcanBuilder {
	ub.lifetime = seconds
	return ub
}

func (ub *UcanBuilder) WithExpiration(timestamp int64) *UcanBuilder {
	ub.expiration = timestamp
	return ub
}

func (ub *UcanBuilder) WithNotBefore(timestamp int64) *UcanBuilder {
	ub.notBefore = timestamp
	return ub
}

func (ub *UcanBuilder) WithFact(key string, fact interface{}) *UcanBuilder {
	ub.facts[key] = fact
	return ub
}

func (ub *UcanBuilder) WithNonce() *UcanBuilder {
	ub.addNonce = true
	return ub
}

func (ub *UcanBuilder) Expiration() *int64 {
	if ub.expiration == 0 {
		if ub.lifetime == 0 {
			panic("expiration and lifetime can not both be empty")
		}
		exp := int64(ub.lifetime) + time.Now().Unix()
		return &exp
	}
	return &ub.expiration
}

func (ub *UcanBuilder) NotBefore() *int64 {
	if ub.notBefore == 0 {
		return nil
	}
	return &ub.notBefore
}

func (ub *UcanBuilder) Build() (*Ucan, error) {
	if ub.issuer == nil {
		return nil, fmt.Errorf("nil issuer")
	}
	if ub.audience == "" {
		return nil, fmt.Errorf("nil audience")
	}

	var err error
	nnc := ""
	if ub.addNonce {
		randNonce := make([]byte, 32)
		randSource.Read(randNonce)
		nnc, err = mb.Encode(mb.Base64url, randNonce)
		if err != nil {
			return nil, fmt.Errorf("failed to generate nonce string, err: %v", err)
		}
	}

	issString, err := ub.issuer.DidString()
	if err != nil {
		return nil, err
	}

	caps, err := capability.BuildCapsFromArray(ub.capabilities)
	if err != nil {
		return nil, err
	}

	ucan := &Ucan{
		Header: UcanHeader{
			ub.issuer.GetJwtAlgorithmName(),
			"JWT",
		},
		Payload: UcanPayload{
			Ucv:  UCAN_VERSION,
			Iss:  issString,
			Aud:  ub.audience,
			Exp:  ub.Expiration(),
			Nbf:  ub.NotBefore(),
			Caps: caps,
			Fct:  ub.facts,
			Prf:  ub.proofs,
			Nnc:  nnc,
		},
		DataToSign: nil,
		Signature:  nil,
	}

	headerBase64, err := ucan.Header.Encode()
	if err != nil {
		return nil, err
	}

	payloadBase64, err := ucan.Payload.Encode()
	if err != nil {
		return nil, err
	}

	dataToSign := headerBase64 + "." + payloadBase64
	signature, err := ub.issuer.Sign(dataToSign)
	if err != nil {
		return nil, err
	}

	ucan.DataToSign = []byte(dataToSign)
	ucan.Signature = []byte(signature)

	return ucan, nil
}
