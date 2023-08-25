package go_ucan_kl

import (
	"fmt"
	"github.com/ipfs/go-cid"
	"go-ucan-kl/capability"
	"go-ucan-kl/store"
	"time"
)

type ProofChain struct {
	ucan          *Ucan
	proofs        []*ProofChain
	redelegations []int
}

func (pc *ProofChain) validateLinkTo(uc *Ucan) error {
	audience := pc.ucan.Audience()
	issuer := uc.Issuer()

	if audience == issuer {
		if pc.ucan.LifetimeEncompasses(uc) {
			return nil
		}
		return fmt.Errorf("Invalid UCAN link: lifetime exceeds attenuation")
	}
	return fmt.Errorf("Invalid UCAN link: audience %s does not match issuer %s", audience, issuer)
}

func ProofChainFromUcan(uc *Ucan, nowTime *time.Time, store store.UcanStore) (*ProofChain, error) {
	err := uc.Validate(nowTime)
	if err != nil {
		return nil, err
	}
	proofs := make([]*ProofChain, 0)
	for _, cidStr := range uc.Proofs() {
		c, err := cid.Decode(cidStr)
		if err != nil {
			return nil, err
		}
		ucanStr, err := store.ReadUcanStr(c)
		if err != nil {
			return nil, err
		}
		proofChain, err := ProofChainFromUcanStr(ucanStr, nowTime, store)
		if err != nil {
			return nil, err
		}
		err = proofChain.validateLinkTo(uc)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, proofChain)
	}

	redelegations := make([]int, 0)
	caps := uc.Capabilities().ToCapsArray()
	for _, cap := range caps {
		if capView, err := capability.ProofDelegationSemantics.ParseCapability(&cap); err != nil {
			// todo: better choose for ProofDelegation and error handling
			continue
		} else {
			scope := capView.Resource.ResourceUri.Scope()
			proofSelection, ok := scope.(capability.ProofSelection)
			if !ok {
				return nil, fmt.Errorf("%t is not ProofSelection type", scope)
			}
			chosenIdx := proofSelection.Index
			if chosenIdx == -1 {
				for i := 0; i < len(proofs); i++ {
					redelegations = append(redelegations, i)
				}
			} else if 0 < chosenIdx && chosenIdx < len(proofs) {
				redelegations = append(redelegations, chosenIdx)
			} else {
				return nil, fmt.Errorf("invalid chosen redelegate proof index:%d", chosenIdx)
			}
		}
	}

	return &ProofChain{
		ucan:          uc,
		proofs:        proofs,
		redelegations: redelegations,
	}, nil
}

func ProofChainFromUcanStr(ucanStr string, nowTime *time.Time, store store.UcanStore) (*ProofChain, error) {
	ucan, err := DecodeUcanString(ucanStr)
	if err != nil {
		return nil, err
	}
	return ProofChainFromUcan(ucan, nowTime, store)
}
