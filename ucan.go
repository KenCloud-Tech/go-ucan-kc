package go_ucan_kl

import (
	"fmt"
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
	return "", NotImplementedError
}

func DecodeUcanHeaderBytes(uhBytes []byte) (uh *UcanHeader, err error) {
	return nil, NotImplementedError
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
	return "", NotImplementedError
}

func DecodeUcanPayloadBytes(upBytes []byte) (*UcanPayload, error) {
	return nil, NotImplementedError
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
	keyMaterial, err := key.NewDidKeyPairFromString(uc.Payload.Iss)
	if err != nil {
		return err
	}
	return keyMaterial.Verify(string(uc.DataToSign), string(uc.Signature))
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
