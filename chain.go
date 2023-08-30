package go_ucan_kl

import (
	"fmt"
	"github.com/ipfs/go-cid"
	. "go-ucan-kl/capability"
	"go-ucan-kl/store"
	"golang.org/x/exp/maps"

	"strings"
	"time"
)

type CapabilityInfo struct {
	originators map[string]bool
	notBefore   *int64
	expires     *int64
	capability  CapabilityView
}

type ProofChain struct {
	ucan          *Ucan
	proofs        []*ProofChain
	redelegations map[int]bool
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

func (pc *ProofChain) ReduceCapabilities(cs *CapabilitySemantics[Scope, Ability]) ([]*CapabilityInfo, error) {
	ancestralCapabilityInfos := make([]*CapabilityInfo, 0)
	for idx, prf := range pc.proofs {
		if _, exist := pc.redelegations[idx]; exist {
		} else {
			capInfos, err := prf.ReduceCapabilities(cs)
			if err != nil {
				return nil, err
			}
			ancestralCapabilityInfos = append(ancestralCapabilityInfos, capInfos...)
		}
	}

	redelegatedCapabilityInfos := make([]*CapabilityInfo, 0)
	for idx, _ := range pc.redelegations {
		capInfos, err := pc.proofs[idx].ReduceCapabilities(cs)
		if err != nil {
			return nil, err
		}
		for _, capInfo := range capInfos {
			capInfo.notBefore = pc.ucan.NotBefore()
			capInfo.expires = pc.ucan.Expires()
			redelegatedCapabilityInfos = append(redelegatedCapabilityInfos, capInfo)
		}
	}

	selfCapabilities := make([]*CapabilityView, 0)
	for _, cap := range pc.ucan.Capabilities().ToCapsArray() {
		capView, err := cs.ParseCapability(&cap)
		if err != nil {
			if strings.Contains(err.Error(), TypeParseError.Error()) {
				continue
			}
			return nil, err
		}
		selfCapabilities = append(selfCapabilities, capView)
	}

	selfCapabilityInfos := make([]*CapabilityInfo, 0)
	if len(pc.proofs) == 0 {
		for _, capView := range selfCapabilities {
			capInfo := &CapabilityInfo{
				originators: map[string]bool{pc.ucan.Issuer(): true},
				notBefore:   pc.ucan.NotBefore(),
				expires:     pc.ucan.Expires(),
				capability:  *capView,
			}
			selfCapabilityInfos = append(selfCapabilityInfos, capInfo)
		}
	} else {
		for _, capView := range selfCapabilities {
			originators := make(map[string]bool)
			for _, ancestralCapabilityInfo := range ancestralCapabilityInfos {
				if ancestralCapabilityInfo.capability.Enables(capView) {
					for ori, _ := range ancestralCapabilityInfo.originators {
						originators[ori] = true
					}
				} else {
					continue
				}
			}

			if len(originators) == 0 {
				originators[pc.ucan.Issuer()] = true
			}

			capInfo := &CapabilityInfo{
				originators: originators,
				notBefore:   pc.ucan.NotBefore(),
				expires:     pc.ucan.Expires(),
				capability:  *capView,
			}
			selfCapabilityInfos = append(selfCapabilityInfos, capInfo)
		}
	}

	selfCapabilityInfos = append(selfCapabilityInfos, redelegatedCapabilityInfos...)

	mergedCapabilityInfos := make([]*CapabilityInfo, 0)
Merge:
	if len(selfCapabilityInfos) > 0 {
		capInfo := selfCapabilityInfos[len(selfCapabilityInfos)-1]
		selfCapabilityInfos = selfCapabilityInfos[:len(selfCapabilityInfos)-1]
		for _, remainCapInfo := range selfCapabilityInfos {
			if remainCapInfo.capability.Enables(&capInfo.capability) {
				maps.Copy(remainCapInfo.originators, capInfo.originators)
				continue Merge
			}
		}
		mergedCapabilityInfos = append(mergedCapabilityInfos, capInfo)
	} else {
		return mergedCapabilityInfos, nil
	}

	return mergedCapabilityInfos, nil
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

	redelegations := make(map[int]bool, 0)
	caps := uc.Capabilities().ToCapsArray()
	for _, cap := range caps {
		if capView, err := ProofDelegationSemantics.ParseCapability(&cap); err != nil {
			// todo: better choose for ProofDelegation and error handling
			if strings.Contains(err.Error(), TypeParseError.Error()) {
				continue
			} else {
				return nil, err
			}
		} else {
			scope := capView.Resource.ResourceUri.Scope()
			proofSelection, ok := scope.(ProofSelection)
			if !ok {
				return nil, fmt.Errorf("%t is not ProofSelection type", scope)
			}
			chosenIdx := proofSelection.Index
			if chosenIdx == -1 {
				for i := 0; i < len(proofs); i++ {
					//redelegations = append(redelegations, i)
					redelegations[i] = true
				}
			} else if 0 < chosenIdx && chosenIdx < len(proofs) {
				//redelegations = append(redelegations, chosenIdx)
				redelegations[chosenIdx] = true
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

func ProofChainFromUcanCid(c cid.Cid, nowTime *time.Time, store store.UcanStore) (*ProofChain, error) {
	ucan, err := store.ReadUcan(c)
	if err != nil {
		return nil, err
	}
	return ProofChainFromUcan(ucan, nowTime, store)
}
