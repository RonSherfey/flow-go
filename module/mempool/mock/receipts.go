// Code generated by mockery v1.0.0. DO NOT EDIT.

package mempool

import (
	flow "github.com/dapperlabs/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"
)

// Receipts is an autogenerated mock type for the Receipts type
type Receipts struct {
	mock.Mock
}

// Add provides a mock function with given fields: receipt
func (_m *Receipts) Add(receipt *flow.ExecutionReceipt) error {
	ret := _m.Called(receipt)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.ExecutionReceipt) error); ok {
		r0 = rf(receipt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// All provides a mock function with given fields:
func (_m *Receipts) All() []*flow.ExecutionReceipt {
	ret := _m.Called()

	var r0 []*flow.ExecutionReceipt
	if rf, ok := ret.Get(0).(func() []*flow.ExecutionReceipt); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*flow.ExecutionReceipt)
		}
	}

	return r0
}

// ByBlockID provides a mock function with given fields: blockID
func (_m *Receipts) ByBlockID(blockID flow.Identifier) []*flow.ExecutionReceipt {
	ret := _m.Called(blockID)

	var r0 []*flow.ExecutionReceipt
	if rf, ok := ret.Get(0).(func(flow.Identifier) []*flow.ExecutionReceipt); ok {
		r0 = rf(blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*flow.ExecutionReceipt)
		}
	}

	return r0
}

// DropForBlock provides a mock function with given fields: blockID
func (_m *Receipts) DropForBlock(blockID flow.Identifier) {
	_m.Called(blockID)
}

// Has provides a mock function with given fields: receiptID
func (_m *Receipts) Has(receiptID flow.Identifier) bool {
	ret := _m.Called(receiptID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(flow.Identifier) bool); ok {
		r0 = rf(receiptID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Rem provides a mock function with given fields: receiptID
func (_m *Receipts) Rem(receiptID flow.Identifier) bool {
	ret := _m.Called(receiptID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(flow.Identifier) bool); ok {
		r0 = rf(receiptID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Size provides a mock function with given fields:
func (_m *Receipts) Size() uint {
	ret := _m.Called()

	var r0 uint
	if rf, ok := ret.Get(0).(func() uint); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint)
	}

	return r0
}
