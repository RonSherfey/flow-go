package handler

import (
	"fmt"
	"sync"

	"github.com/onflow/cadence/runtime"

	"github.com/onflow/flow-go/fvm/errors"
	"github.com/onflow/flow-go/fvm/programs"
	"github.com/onflow/flow-go/fvm/state"
	"github.com/onflow/flow-go/model/flow"
)

// ContractHandler handles all interaction
// with smart contracts such as get/set/update
// it also captures all changes as deltas and
// only commit them when called so smart contract
// updates can be delayed until end of the tx execution
type ContractHandler struct {
	accounts                    *state.Accounts
	draftUpdates                map[programs.ContractUpdateKey]programs.ContractUpdate
	restrictedDeploymentEnabled bool
	authorizedAccounts          []runtime.Address
	// handler doesn't have to be thread safe and right now
	// is only used in a single thread but a mutex has been added
	// here to prevent accidental multi-thread use in the future
	lock sync.Mutex
}

func NewContractHandler(accounts *state.Accounts, restrictedDeploymentEnabled bool, authorizedAccounts []runtime.Address) *ContractHandler {
	return &ContractHandler{
		accounts:                    accounts,
		draftUpdates:                make(map[programs.ContractUpdateKey]programs.ContractUpdate),
		restrictedDeploymentEnabled: restrictedDeploymentEnabled,
		authorizedAccounts:          authorizedAccounts,
	}
}

func (h *ContractHandler) GetContractNames(address runtime.Address) (names []string, err error) {
	names, err = h.accounts.GetContractNames(flow.Address(address))
	return
}

func (h *ContractHandler) GetContract(address runtime.Address, name string) (code []byte, err error) {
	code, err = h.accounts.GetContract(name, flow.Address(address))
	return
}

func (h *ContractHandler) SetContract(address runtime.Address, name string, code []byte, signingAccounts []runtime.Address) (err error) {
	// check if authorized
	if !h.isAuthorized(signingAccounts) {
		err = errors.NewOperationAuthorizationErrorf("SetContract", "setting contracts requires authorization from specific accounts")
		return fmt.Errorf("setting contract failed: %w", err)
	}
	add := flow.Address(address)
	h.lock.Lock()
	defer h.lock.Unlock()
	uk := programs.ContractUpdateKey{Address: add, Name: name}
	u := programs.ContractUpdate{ContractUpdateKey: uk, Code: code}
	h.draftUpdates[uk] = u

	return nil
}

func (h *ContractHandler) RemoveContract(address runtime.Address, name string, signingAccounts []runtime.Address) (err error) {
	// check if authorized
	if !h.isAuthorized(signingAccounts) {
		err = errors.NewOperationAuthorizationErrorf("RemoveContract", "removing contracts requires authorization from specific accounts")
		return fmt.Errorf("removing contract failed: %w", err)
	}

	add := flow.Address(address)
	// removes are stored in the draft updates with code value of nil
	h.lock.Lock()
	defer h.lock.Unlock()
	uk := programs.ContractUpdateKey{Address: add, Name: name}
	u := programs.ContractUpdate{ContractUpdateKey: uk}
	h.draftUpdates[uk] = u

	return nil
}

func (h *ContractHandler) Commit() ([]programs.ContractUpdateKey, error) {

	h.lock.Lock()
	defer h.lock.Unlock()

	updatedKeys := h.UpdateKeys()
	var err error
	for _, v := range h.draftUpdates {
		if len(v.Code) > 0 {
			err = h.accounts.SetContract(v.Name, v.Address, v.Code)
			if err != nil {
				return nil, err
			}
		} else {
			err = h.accounts.DeleteContract(v.Name, v.Address)
			if err != nil {
				return nil, err
			}
		}
	}

	// reset draft
	h.draftUpdates = make(map[programs.ContractUpdateKey]programs.ContractUpdate)
	return updatedKeys, nil
}

func (h *ContractHandler) Rollback() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.draftUpdates = make(map[programs.ContractUpdateKey]programs.ContractUpdate)
	return nil
}

func (h *ContractHandler) HasUpdates() bool {
	return len(h.draftUpdates) > 0
}

func (h *ContractHandler) UpdateKeys() []programs.ContractUpdateKey {
	if len(h.draftUpdates) == 0 {
		return nil
	}
	keys := make([]programs.ContractUpdateKey, 0, len(h.draftUpdates))
	for k := range h.draftUpdates {
		keys = append(keys, k)
	}
	return keys
}

func (h *ContractHandler) isAuthorized(signingAccounts []runtime.Address) bool {
	if h.restrictedDeploymentEnabled {
		for _, authorized := range h.authorizedAccounts {
			for _, signer := range signingAccounts {
				if signer == authorized {
					// a single authorized singer is enough
					return true
				}
			}
		}
		return false
	}
	return true
}
