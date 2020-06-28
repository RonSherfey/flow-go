// Code generated by mockery v1.0.0. DO NOT EDIT.

package mempool

import flow "github.com/dapperlabs/flow-go/model/flow"

import mock "github.com/stretchr/testify/mock"
import verification "github.com/dapperlabs/flow-go/model/verification"

// PendingReceipts is an autogenerated mock type for the PendingReceipts type
type PendingReceipts struct {
	mock.Mock
}

// Add provides a mock function with given fields: preceipt
func (_m *PendingReceipts) Add(preceipt *verification.PendingReceipt) bool {
	ret := _m.Called(preceipt)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*verification.PendingReceipt) bool); ok {
		r0 = rf(preceipt)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// All provides a mock function with given fields:
func (_m *PendingReceipts) All() []*verification.PendingReceipt {
	ret := _m.Called()

	var r0 []*verification.PendingReceipt
	if rf, ok := ret.Get(0).(func() []*verification.PendingReceipt); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*verification.PendingReceipt)
		}
	}

	return r0
}

// Get provides a mock function with given fields: preceiptID
func (_m *PendingReceipts) Get(preceiptID flow.Identifier) (*verification.PendingReceipt, bool) {
	ret := _m.Called(preceiptID)

	var r0 *verification.PendingReceipt
	if rf, ok := ret.Get(0).(func(flow.Identifier) *verification.PendingReceipt); ok {
		r0 = rf(preceiptID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*verification.PendingReceipt)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(flow.Identifier) bool); ok {
		r1 = rf(preceiptID)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// Has provides a mock function with given fields: preceiptID
func (_m *PendingReceipts) Has(preceiptID flow.Identifier) bool {
	ret := _m.Called(preceiptID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(flow.Identifier) bool); ok {
		r0 = rf(preceiptID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Rem provides a mock function with given fields: preceiptID
func (_m *PendingReceipts) Rem(preceiptID flow.Identifier) bool {
	ret := _m.Called(preceiptID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(flow.Identifier) bool); ok {
		r0 = rf(preceiptID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
