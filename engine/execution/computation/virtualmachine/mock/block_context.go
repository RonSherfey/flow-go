// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	flow "github.com/dapperlabs/flow-go/model/flow"
	mock "github.com/stretchr/testify/mock"

	virtualmachine "github.com/dapperlabs/flow-go/engine/execution/computation/virtualmachine"
)

// BlockContext is an autogenerated mock type for the BlockContext type
type BlockContext struct {
	mock.Mock
}

// ExecuteScript provides a mock function with given fields: ledger, script
func (_m *BlockContext) ExecuteScript(ledger virtualmachine.Ledger, script []byte) (*virtualmachine.ScriptResult, error) {
	ret := _m.Called(ledger, script)

	var r0 *virtualmachine.ScriptResult
	if rf, ok := ret.Get(0).(func(virtualmachine.Ledger, []byte) *virtualmachine.ScriptResult); ok {
		r0 = rf(ledger, script)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*virtualmachine.ScriptResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(virtualmachine.Ledger, []byte) error); ok {
		r1 = rf(ledger, script)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExecuteTransaction provides a mock function with given fields: ledger, tx, options
func (_m *BlockContext) ExecuteTransaction(ledger virtualmachine.Ledger, tx *flow.TransactionBody, options ...virtualmachine.TransactionContextOption) (*virtualmachine.TransactionResult, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ledger, tx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *virtualmachine.TransactionResult
	if rf, ok := ret.Get(0).(func(virtualmachine.Ledger, *flow.TransactionBody, ...virtualmachine.TransactionContextOption) *virtualmachine.TransactionResult); ok {
		r0 = rf(ledger, tx, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*virtualmachine.TransactionResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(virtualmachine.Ledger, *flow.TransactionBody, ...virtualmachine.TransactionContextOption) error); ok {
		r1 = rf(ledger, tx, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
