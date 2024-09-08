package ucan

import (
	"github.com/KenCloud-Tech/go-ucan-kc/test/fixtures"
	"github.com/stretchr/testify/assert"
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
