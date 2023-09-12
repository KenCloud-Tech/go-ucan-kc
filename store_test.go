package go_ucan_kl

import (
	"github.com/stretchr/testify/assert"
	"go-ucan-kl/test/fixtures"
	"testing"
)

func TestStore(t *testing.T) {
	store := NewMemoryStore()
	ucan, err := DefaultBuilder().
		IssuedBy(fixtures.TestIdentities.AliceKey).
		ForAudience(fixtures.TestIdentities.BobDidString).
		Build()
	if err != nil {
		t.Fatal(err.Error())
	}

	c, err := store.WriteUcan(ucan, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	reUcan, err := store.ReadUcan(c)
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, ucan, reUcan)
}
