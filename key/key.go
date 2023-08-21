package key

import (
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
	//"github.com/libp2p/go-libp2p-core/crypto"
	mb "github.com/multiformats/go-multibase"
	varint "github.com/multiformats/go-varint"
)

const (
	// KeyPrefix indicates a decentralized identifier that uses the key method
	KeyPrefix = "did:key"
	// MulticodecKindRSAPubKey rsa-x509-pub https://github.com/multiformats/multicodec/pull/226
	MulticodecKindRSAPubKey = 0x1205
	// MulticodecKindEd25519PubKey ed25519-pub
	MulticodecKindEd25519PubKey = 0xed
)

var (
	InvalidSigningMethod = fmt.Errorf("invalid signing method")
)

var (
	DidKeyPairOne *DidKeyPair
	DidOneStr     string
	DidKeyPairTwo *DidKeyPair
	DidTwoStr     string
)

type KeyMaterial interface {
	GetJwtAlgorithmName() string
	DidString() (string, error)
	Verify(payload string, signature string) error // Returns nil if signature is valid
	Sign(payload string) (string, error)           // Returns encoded signature or error
}

var _ KeyMaterial = &DidKeyPair{}

// todo: sperate into different key type, Verify and Sign may be different
type DidKeyPair struct {
	pubKey    crypto.PubKey
	signKey   interface{}
	verifyKey interface{}
	jwt.SigningMethod
	DIDString string
}

func NewDidKeyPairFromPrivateKeyString(str string) (*DidKeyPair, error) {
	dkp := &DidKeyPair{}
	privKeyBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	privKey, err := crypto.UnmarshalPrivateKey(privKeyBytes)
	if err != nil {
		return nil, err
	}
	rawPrivBytes, err := privKey.Raw()
	if err != nil {
		return nil, fmt.Errorf("getting private key bytes: %w", err)
	}
	dkp.pubKey = privKey.GetPublic()

	keyType := privKey.Type()
	methodStr := ""
	switch keyType {
	case crypto.RSA:
		methodStr = "RS256"
		dkp.signKey, err = x509.ParsePKCS1PrivateKey(rawPrivBytes)
		if err != nil {
			return nil, err
		}
	case crypto.Ed25519:
		methodStr = "EdDSA"
	}

	dkp.SigningMethod = jwt.GetSigningMethod(methodStr)
	// todo string

	return dkp, nil
}

func ParseDidStringAndGetVertifyKey(did string) (*DidKeyPair, error) {
	key := &DidKeyPair{}
	if !strings.HasPrefix(did, KeyPrefix) {
		return nil, fmt.Errorf("decentralized identifier is not a 'key' type")
	}

	str := strings.TrimPrefix(did, KeyPrefix+":")

	enc, data, err := mb.Decode(str)
	if err != nil {
		return nil, fmt.Errorf("decoding multibase: %w", err)
	}

	if enc != mb.Base58BTC {
		return nil, fmt.Errorf("unexpected multibase encoding: %s", mb.EncodingToStr[enc])
	}

	keyType, n, err := varint.FromUvarint(data)
	if err != nil {
		return nil, err
	}

	methodStr := ""
	switch keyType {
	case MulticodecKindRSAPubKey:
		methodStr = "RS256"
		pub, err := crypto.UnmarshalRsaPublicKey(data[n:])
		if err != nil {
			return nil, err
		}
		key.pubKey = pub
	case MulticodecKindEd25519PubKey:
		methodStr = "EdDSA"
		pub, err := crypto.UnmarshalEd25519PublicKey(data[n:])
		if err != nil {
			return nil, err
		}
		key.pubKey = pub
	}

	verifyKey, err := key.getVerifyKey()
	if err != nil {
		return nil, err
	}
	key.verifyKey = verifyKey
	key.SigningMethod = jwt.GetSigningMethod(methodStr)
	return key, nil
}

func (dkp *DidKeyPair) DidString() (string, error) {
	if dkp.DIDString != "" {
		return dkp.DIDString, nil
	}

	var multiCodec uint64
	switch dkp.pubKey.Type() {
	case crypto.RSA:
		multiCodec = MulticodecKindRSAPubKey
	case crypto.Ed25519:
		multiCodec = MulticodecKindEd25519PubKey
	default:
		panic("unexpected crypto type")
	}

	raw, err := dkp.pubKey.Raw()
	if err != nil {
		return "", err
	}

	size := varint.UvarintSize(multiCodec)
	data := make([]byte, size+len(raw))
	n := varint.PutUvarint(data, multiCodec)
	copy(data[n:], raw)

	b58BKeyStr, err := mb.Encode(mb.Base58BTC, data)
	if err != nil {
		return "", err
	}

	str := fmt.Sprintf("%s:%s", KeyPrefix, b58BKeyStr)
	dkp.DIDString = str

	return dkp.DIDString, nil
}

func (dkp *DidKeyPair) GetJwtAlgorithmName() string {
	return dkp.Alg()
}

func (dkp *DidKeyPair) Verify(payload string, signature string) error {
	if dkp.SigningMethod == nil {
		return InvalidSigningMethod
	}

	return dkp.SigningMethod.Verify(payload, signature, dkp.verifyKey)
}

func (dkp *DidKeyPair) Sign(payload string) (string, error) {
	if dkp.SigningMethod == nil {
		return "", InvalidSigningMethod
	}

	return dkp.SigningMethod.Sign(payload, dkp.signKey)
}

// VerifyKey returns the backing implementation for a public key, one of:
// *rsa.PublicKey, ed25519.PublicKey
func (dkp *DidKeyPair) getVerifyKey() (interface{}, error) {
	if dkp.verifyKey != nil {
		return dkp.verifyKey, nil
	}
	var verifyKey interface{}

	rawPubBytes, err := dkp.pubKey.Raw()
	if err != nil {
		return nil, err
	}
	switch dkp.pubKey.Type() {
	case crypto.RSA:
		verifyKeyiface, err := x509.ParsePKIXPublicKey(rawPubBytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		verifyKey, ok = verifyKeyiface.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key is not an RSA key. got type: %T", verifyKeyiface)
		}
	case crypto.Ed25519:
		verifyKey = ed25519.PublicKey(rawPubBytes)
	default:
		return nil, fmt.Errorf("unrecognized Public Key type: %s", dkp.pubKey.Type())
	}

	dkp.verifyKey = verifyKey
	return verifyKey, nil
}

// ID is a DID:key identifier
type ID struct {
	crypto.PubKey
}

// NewID constructs an Identifier from a public key
func NewID(pub crypto.PubKey) (ID, error) {
	switch pub.Type() {
	case crypto.Ed25519, crypto.RSA:
		return ID{PubKey: pub}, nil
	default:
		return ID{}, fmt.Errorf("unsupported key type: %s", pub.Type())
	}
}

// MulticodecType indicates the type for this multicodec
func (id ID) MulticodecType() uint64 {
	switch id.Type() {
	case crypto.RSA:
		return MulticodecKindRSAPubKey
	case crypto.Ed25519:
		return MulticodecKindEd25519PubKey
	default:
		panic("unexpected crypto type")
	}
}

// String returns this did:key formatted as a string
func (id ID) String() string {
	raw, err := id.Raw()
	if err != nil {
		return ""
	}

	t := id.MulticodecType()
	size := varint.UvarintSize(t)
	data := make([]byte, size+len(raw))
	n := varint.PutUvarint(data, t)
	copy(data[n:], raw)

	b58BKeyStr, err := mb.Encode(mb.Base58BTC, data)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s:%s", KeyPrefix, b58BKeyStr)
}

// VerifyKey returns the backing implementation for a public key, one of:
// *rsa.PublicKey, ed25519.PublicKey
func (id ID) VerifyKey() (interface{}, error) {
	rawPubBytes, err := id.PubKey.Raw()
	if err != nil {
		return nil, err
	}
	switch id.PubKey.Type() {
	case crypto.RSA:
		verifyKeyiface, err := x509.ParsePKIXPublicKey(rawPubBytes)
		if err != nil {
			return nil, err
		}
		verifyKey, ok := verifyKeyiface.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key is not an RSA key. got type: %T", verifyKeyiface)
		}
		return verifyKey, nil
	case crypto.Ed25519:
		return ed25519.PublicKey(rawPubBytes), nil
	default:
		return nil, fmt.Errorf("unrecognized Public Key type: %s", id.PubKey.Type())
	}
}

// Parse turns a string into a key method ID
//func Parse(keystr string) (ID, error) {
//	var id ID
//	if !strings.HasPrefix(keystr, KeyPrefix) {
//		return id, fmt.Errorf("decentralized identifier is not a 'key' type")
//	}
//
//	keystr = strings.TrimPrefix(keystr, KeyPrefix+":")
//
//	enc, data, err := mb.Decode(keystr)
//	if err != nil {
//		return id, fmt.Errorf("decoding multibase: %w", err)
//	}
//
//	if enc != mb.Base58BTC {
//		return id, fmt.Errorf("unexpected multibase encoding: %s", mb.EncodingToStr[enc])
//	}
//
//	keyType, n, err := varint.FromUvarint(data)
//	if err != nil {
//		return id, err
//	}
//
//	switch keyType {
//	case MulticodecKindRSAPubKey:
//		pub, err := crypto.UnmarshalRsaPublicKey(data[n:])
//		if err != nil {
//			return id, err
//		}
//		return ID{pub}, nil
//	case MulticodecKindEd25519PubKey:
//		pub, err := crypto.UnmarshalEd25519PublicKey(data[n:])
//		if err != nil {
//			return id, err
//		}
//		return ID{pub}, nil
//	}
//
//	return id, fmt.Errorf("unrecognized key type multicodec prefix: %x", data[0])
//}
