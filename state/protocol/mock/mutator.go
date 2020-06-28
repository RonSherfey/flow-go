// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import flow "github.com/dapperlabs/flow-go/model/flow"
import mock "github.com/stretchr/testify/mock"

// Mutator is an autogenerated mock type for the Mutator type
type Mutator struct {
	mock.Mock
}

// Bootstrap provides a mock function with given fields: root, result, seal
func (_m *Mutator) Bootstrap(root *flow.Block, result *flow.ExecutionResult, seal *flow.Seal) error {
	ret := _m.Called(root, result, seal)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.Block, *flow.ExecutionResult, *flow.Seal) error); ok {
		r0 = rf(root, result, seal)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Extend provides a mock function with given fields: block
func (_m *Mutator) Extend(block *flow.Block) error {
	ret := _m.Called(block)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.Block) error); ok {
		r0 = rf(block)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Finalize provides a mock function with given fields: blockID
func (_m *Mutator) Finalize(blockID flow.Identifier) error {
	ret := _m.Called(blockID)

	var r0 error
	if rf, ok := ret.Get(0).(func(flow.Identifier) error); ok {
		r0 = rf(blockID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
