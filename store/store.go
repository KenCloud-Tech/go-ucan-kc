package store

import (
	"fmt"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	. "go-ucan-kl"
)

var DefaultPrefix = cid.Prefix{
	Version:  1,
	Codec:    cid.Raw,
	MhType:   mh.BLAKE3,
	MhLength: -1, // default length
}

type UcanStore interface {
	ReadUcan(c cid.Cid) (*Ucan, error)
	WriteUcan(uc *Ucan) (cid.Cid, error)
	ReadUcanStr(c cid.Cid) (string, error)
	WriteUcanStr(str string) (cid.Cid, error)
}

var _ UcanStore = &MemoryStore{}

type MemoryStore struct {
	store map[cid.Cid]string
}

func (m MemoryStore) ReadUcan(c cid.Cid) (*Ucan, error) {
	if str, ok := m.store[c]; !ok {
		return nil, fmt.Errorf("ucan for cid:%s not exist", c.String())
	} else {
		return DecodeUcanString(str)
	}
}

func (m MemoryStore) WriteUcan(uc *Ucan) (cid.Cid, error) {
	c, str, err := uc.ToCid(nil)
	if err != nil {
		return cid.Undef, err
	}
	m.store[c] = str
	return c, err
}

func (m MemoryStore) ReadUcanStr(c cid.Cid) (string, error) {
	if str, ok := m.store[c]; !ok {
		return "", fmt.Errorf("ucan for cid:%s not exist", c.String())
	} else {
		return str, nil
	}
}

func (m MemoryStore) WriteUcanStr(str string) (cid.Cid, error) {
	_, err := DecodeUcanString(str)
	if err != nil {
		return cid.Undef, err
	}
	c, err := DefaultPrefix.Sum([]byte(str))
	if err != nil {
		return cid.Undef, err
	}
	m.store[c] = str
	return c, nil
}
