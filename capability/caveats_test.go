package capability

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmptyCaveats(t *testing.T) {
	_, err := BuildCaveat([]byte(""))
	assert.Contains(t, err.Error(), "caveat must be json object, but got")

	cvOne, err := EmailSemantics.Parse("mailto:bob@email.com", "email/send", []byte(""))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cvOne.Caveat, NullJson)

	cvTwo, err := EmailSemantics.Parse("mailto:bob@email.com", "email/send", nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cvTwo.Caveat, NullJson)

	cvThree, err := EmailSemantics.Parse("mailto:bob@email.com", "email/send", NullJson)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cvThree.Caveat, NullJson)
}
