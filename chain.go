package ucan

import (
	"fmt"
	. "github.com/KenCloud-Tech/go-ucan-kc/capability"
	"github.com/ipfs/go-cid"
	"golang.org/x/exp/maps"

	"strings"
	"time"
)

type CapabilityInfo struct {
	Originators map[string]bool
	NotBefore   *int64
	Expires     *int64
	Capability  CapabilityView
}

type ProofChain struct {
	ucan          *Ucan
	proofs        []*ProofChain
	redelegations map[int]bool
}

func (pc *ProofChain) ValidateLinkTo(uc *Ucan) error {
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

func ReduceCapabilities[S Scope, A Ability](pc *ProofChain) ([]*CapabilityInfo, error) {
	cs := &CapabilitySemantics[S, A]{}

	// get all ancestral CapabilityInfos(exclude delegated)
	ancestralCapabilityInfos := make([]*CapabilityInfo, 0)
	for idx, prf := range pc.proofs {
		if _, exist := pc.redelegations[idx]; exist {
		} else {
			capInfos, err := ReduceCapabilities[S, A](prf)
			if err != nil {
				return nil, err
			}
			ancestralCapabilityInfos = append(ancestralCapabilityInfos, capInfos...)
		}
	}

	// get all delegated CapabilityInfos from ancestral
	redelegatedCapabilityInfos := make([]*CapabilityInfo, 0)
	for idx, _ := range pc.redelegations {
		//capInfos, err := pc.proofs[idx].ReduceCapabilities(cs)
		capInfos, err := ReduceCapabilities[S, A](pc.proofs[idx])
		if err != nil {
			return nil, err
		}
		for _, capInfo := range capInfos {
			capInfo.NotBefore = pc.ucan.NotBefore()
			capInfo.Expires = pc.ucan.Expires()
			redelegatedCapabilityInfos = append(redelegatedCapabilityInfos, capInfo)
		}
	}

	// all self CapabilityView
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

	// get all CapabilityInfos in self caps and set the originators(may inherit from ancestral issuer if not the ori sets as self)
	selfCapabilityInfos := make([]*CapabilityInfo, 0)
	if len(pc.proofs) == 0 {
		for _, capView := range selfCapabilities {
			capInfo := &CapabilityInfo{
				Originators: map[string]bool{pc.ucan.Issuer(): true},
				NotBefore:   pc.ucan.NotBefore(),
				Expires:     pc.ucan.Expires(),
				Capability:  *capView,
			}
			selfCapabilityInfos = append(selfCapabilityInfos, capInfo)
		}
	} else {
		for _, capView := range selfCapabilities {
			originators := make(map[string]bool)
			for _, ancestralCapabilityInfo := range ancestralCapabilityInfos {
				if ancestralCapabilityInfo.Capability.Enables(capView) {
					for ori, _ := range ancestralCapabilityInfo.Originators {
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
				Originators: originators,
				NotBefore:   pc.ucan.NotBefore(),
				Expires:     pc.ucan.Expires(),
				Capability:  *capView,
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
			if remainCapInfo.Capability.Enables(&capInfo.Capability) {
				maps.Copy(remainCapInfo.Originators, capInfo.Originators)
				goto Merge
			}
		}
		mergedCapabilityInfos = append(mergedCapabilityInfos, capInfo)
		goto Merge
	} else {
		return mergedCapabilityInfos, nil
	}
}

func ProofChainFromUcan(uc *Ucan, nowTime *time.Time, store UcanStore) (*ProofChain, error) {
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
		err = proofChain.ValidateLinkTo(uc)
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

func ProofChainFromUcanStr(ucanStr string, nowTime *time.Time, store UcanStore) (*ProofChain, error) {
	ucan, err := DecodeUcanString(ucanStr)
	if err != nil {
		return nil, err
	}
	return ProofChainFromUcan(ucan, nowTime, store)
}

func ProofChainFromUcanCid(c cid.Cid, nowTime *time.Time, store UcanStore) (*ProofChain, error) {
	ucan, err := store.ReadUcan(c)
	if err != nil {
		return nil, err
	}
	return ProofChainFromUcan(ucan, nowTime, store)
}
