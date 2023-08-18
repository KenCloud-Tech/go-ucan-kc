package fixtures

import didkey "go-ucan-kl/key"

type Identities struct {
	AliceKey   didkey.ID
	BobKey     didkey.ID
	MalloryKey didkey.ID

	AliceDidString   string
	BobDidString     string
	MalloryDidString string
}

var TestIdentities Identities

func init() {
	var err error
	TestIdentities.AliceKey, err = didkey.Parse("U+bzp2GaFQHso587iSFWPSeCzbSfn/CbNHEz7ilKRZ1UQMmMS7qq4UhTzKn3X9Nj/4xgrwa+UqhMOeo4Ki8JUw==")
	if err != nil {
		panic(err.Error())
	}
	TestIdentities.AliceDidString = TestIdentities.AliceKey.String()

	TestIdentities.BobKey, err = didkey.Parse("G4+QCX1b3a45IzQsQd4gFMMe0UB1UOx9bCsh8uOiKLER69eAvVXvc8P2yc4Iig42Bv7JD2zJxhyFALyTKBHipg==")
	if err != nil {
		panic(err.Error())
	}
	TestIdentities.BobDidString = TestIdentities.BobKey.String()

	TestIdentities.MalloryKey, err = didkey.Parse("LR9AL2MYkMARuvmV3MJV8sKvbSOdBtpggFCW8K62oZDR6UViSXdSV/dDcD8S9xVjS61vh62JITx7qmLgfQUSZQ==")
	if err != nil {
		panic(err.Error())
	}
	TestIdentities.MalloryDidString = TestIdentities.MalloryKey.String()
}
