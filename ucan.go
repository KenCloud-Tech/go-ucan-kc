package go_ucan_kl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-cid"
	mb "github.com/multiformats/go-multibase"
	"go-ucan-kl/capability"
	"go-ucan-kl/key"
	"strings"
	"time"
)

const (
	UCAN_VERSION = "0.10.0-dev"
)

var (
	// NotImplementedError will be deleted later
	NotImplementedError = fmt.Errorf("Not Implemented")
	EncodingError       = fmt.Errorf("invalid encoding")
	UcanForamtError     = fmt.Errorf("Invalid Ucan foramt")
	UcanExpiredError    = fmt.Errorf("Expired")
	UcanNotActiveError  = fmt.Errorf("Not active yet")
)

type UcanHeader struct {
	Algorithm string
	Type      string
}

func (uh *UcanHeader) Encode() (string, error) {
	// todo: dose direct json bytes equal with the DagJson bytes, not sure?
	jsonBytes, err := json.Marshal(uh)
	if err != nil {
		return "", err
	}
	//buffer := bytes.NewReader(jsonBytes)
	//mapBuilder := basicnode.Prototype.Map.NewBuilder()
	//err = dagjson.Decode(mapBuilder, buffer)
	//if err != nil {
	//	return "", err
	//}
	//
	//node := mapBuilder.Build()
	//resBuffer := new(bytes.Buffer)
	//err = dagjson.Encode(node, resBuffer)
	//if err != nil {
	//	return "", err
	//}
	//
	//return mb.Encode(mb.Base64url, resBuffer.Bytes())

	return mb.Encode(mb.Base64url, jsonBytes)
}

func DecodeUcanHeaderBytes(uhBytes []byte) (*UcanHeader, error) {
	uh := &UcanHeader{}
	err := json.Unmarshal(uhBytes, uh)
	return uh, err
}

type UcanPayload struct {
	Ucv  string
	Iss  string
	Aud  string
	Exp  *int64
	Nbf  *int64
	Nnc  string
	Caps capability.Capabilities
	Fct  map[string]interface{}
	Prf  []string
}

func (up *UcanPayload) Encode() (string, error) {
	jsonBytes, err := json.Marshal(up)
	if err != nil {
		return "", err
	}
	//buffer := bytes.NewReader(jsonBytes)
	//mapBuilder := basicnode.Prototype.Map.NewBuilder()
	//err = dagjson.Decode(mapBuilder, buffer)
	//if err != nil {
	//	return "", err
	//}
	//
	//node := mapBuilder.Build()
	//resBuffer := new(bytes.Buffer)
	//err = dagjson.Encode(node, resBuffer)
	//if err != nil {
	//	return "", err
	//}
	//
	//return mb.Encode(mb.Base64url, resBuffer.Bytes())

	return mb.Encode(mb.Base64url, jsonBytes)
}

func DecodeUcanPayloadBytes(upBytes []byte) (*UcanPayload, error) {
	up := &UcanPayload{}
	err := json.Unmarshal(upBytes, up)
	return up, err
}

type Ucan struct {
	Header     UcanHeader
	Payload    UcanPayload
	DataToSign []byte
	Signature  []byte
}

func NewUcan(header UcanHeader, payload UcanPayload, signedData []byte, signature []byte) (Ucan, error) {
	return Ucan{
		header,
		payload,
		signedData,
		signature,
	}, nil
}

func (uc *Ucan) Validate(checkTime *time.Time) error {
	if uc.isExpired(checkTime) {
		return UcanExpiredError
	}
	if uc.isTooEarly(checkTime) {
		return UcanNotActiveError
	}

	return uc.checkSignature()
}

func (uc *Ucan) checkSignature() error {
	keyMaterial, err := key.ParseDidStringAndGetVertifyKey(uc.Payload.Iss)
	if err != nil {
		return err
	}
	return keyMaterial.Verify(string(uc.DataToSign), string(uc.Signature))
}

func (uc *Ucan) Audience() string {
	return uc.Payload.Aud
}

func (uc *Ucan) Issuer() string {
	return uc.Payload.Iss
}

func (uc *Ucan) Proofs() []string {
	return uc.Payload.Prf
}

func (uc *Ucan) Capabilities() capability.Capabilities {
	return uc.Payload.Caps
}

func (uc *Ucan) Expires() *int64 {
	return uc.Payload.Exp
}

func (uc *Ucan) Facts() map[string]interface{} {
	return uc.Payload.Fct
}

func (uc *Ucan) Nonce() string {
	return uc.Payload.Nnc
}

func (uc *Ucan) NotBefore() *int64 {
	return uc.Payload.Nbf
}

func (uc *Ucan) isExpired(checkTime *time.Time) bool {
	exp := uc.Payload.Exp
	if exp == nil {
		return true
	}

	var timeInt int64
	if checkTime == nil {
		timeInt = time.Now().Unix()
	} else {
		timeInt = checkTime.Unix()
	}

	return *exp < timeInt
}

func (uc *Ucan) isTooEarly(checkTime *time.Time) bool {
	nbf := uc.Payload.Nbf
	if nbf == nil {
		return false
	}

	var timeInt int64
	if checkTime == nil {
		timeInt = time.Now().Unix()
	} else {
		timeInt = checkTime.Unix()
	}

	return *nbf > timeInt
}

func (uc *Ucan) LifetimeBeginsBefore(other *Ucan) bool {
	if uc.Payload.Nbf == nil {
		return true
	} else if other.Payload.Nbf == nil {
		return false
	} else {
		return *uc.Payload.Nbf <= *other.Payload.Nbf
	}
}

func (uc *Ucan) LifetimeEndsAfter(other *Ucan) bool {
	if uc.Payload.Exp == nil {
		return true
	} else if other.Payload.Exp == nil {
		return false
	} else {
		return *uc.Payload.Exp >= *other.Payload.Exp
	}
}

func (uc *Ucan) LifetimeEncompasses(other *Ucan) bool {
	return uc.LifetimeBeginsBefore(other) && uc.LifetimeEndsAfter(other)
}

func (uc *Ucan) Encode() (string, error) {
	header, err := uc.Header.Encode()
	if err != nil {
		return "", err
	}
	payload, err := uc.Payload.Encode()
	if err != nil {
		return "", err
	}
	signature, err := mb.Encode(mb.Base64url, uc.Signature)
	if err != nil {
		return "", err
	}

	return header + "." + payload + "." + signature, nil
}

func (uc *Ucan) ToCid(prefix *cid.Prefix) (cid.Cid, string, error) {
	if prefix == nil {
		prefix = &DefaultPrefix
	} else {
		if prefix.Codec != cid.Raw {
			return cid.Undef, "", fmt.Errorf("Ucan cid codec must be RawCodec(0x55)")
		}
	}
	ucanStr, err := uc.Encode()
	if err != nil {
		return cid.Undef, "", err
	}
	c, err := prefix.Sum([]byte(ucanStr))
	return c, ucanStr, err
}

func (uc *Ucan) Equals(other *Ucan) bool {
	ucBytes, err := json.Marshal(uc)
	if err != nil {
		panic(err.Error())
	}
	otherBytes, err := json.Marshal(other)
	if err != nil {
		panic(err.Error())
	}
	return bytes.Equal(ucBytes, otherBytes)
}

func DecodeUcanString(ucStr string) (*Ucan, error) {
	parts := strings.Split(ucStr, ".")
	if len(parts) != 3 {
		return nil, UcanForamtError
	}
	dataToSign := strings.Join(parts[:2], ".")

	var err error
	var encoding mb.Encoding
	partsBytes := make([][]byte, 3)
	for i := range parts {
		encoding, partsBytes[i], err = mb.Decode(parts[i])
		if err != nil {
			return nil, fmt.Errorf("%d-%v", EncodingError, err)
		}
		if encoding != mb.Base64url {
			return nil, fmt.Errorf("%d-%d", UcanForamtError, EncodingError)
		}
	}

	header, err := DecodeUcanHeaderBytes(partsBytes[0])
	if err != nil {
		return nil, err
	}
	payload, err := DecodeUcanPayloadBytes(partsBytes[1])
	if err != nil {
		return nil, err
	}
	signature := partsBytes[2]

	return &Ucan{
		Header:     *header,
		Payload:    *payload,
		DataToSign: []byte(dataToSign),
		Signature:  signature,
	}, nil
}
