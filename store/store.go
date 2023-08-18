package store

import (
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

var DefaultPrefix = cid.Prefix{
	Version:  1,
	Codec:    cid.Raw,
	MhType:   mh.SHA2_256,
	MhLength: -1, // default length
}
