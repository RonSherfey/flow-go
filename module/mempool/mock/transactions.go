// Code generated by mockery v1.0.0. DO NOT EDIT.

package mempool

import (
	flow "github.com/dapperlabs/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"
)

// Transactions is an autogenerated mock type for the Transactions type
type Transactions struct {
	mock.Mock
}

// Add provides a mock function with given fields: tx
func (_m *Transactions) Add(tx *flow.TransactionBody) error {
	ret := _m.Called(tx)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.TransactionBody) error); ok {
		r0 = rf(tx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// All provides a mock function with given fields:
func (_m *Transactions) All() []*flow.TransactionBody {
	ret := _m.Called()

	var r0 []*flow.TransactionBody
	if rf, ok := ret.Get(0).(func() []*flow.TransactionBody); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*flow.TransactionBody)
		}
	}

	return r0
}

// ByID provides a mock function with given fields: txID
func (_m *Transactions) ByID(txID flow.Identifier) (*flow.TransactionBody, error) {
	ret := _m.Called(txID)

	var r0 *flow.TransactionBody
	if rf, ok := ret.Get(0).(func(flow.Identifier) *flow.TransactionBody); ok {
		r0 = rf(txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.TransactionBody)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Identifier) error); ok {
		r1 = rf(txID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Has provides a mock function with given fields: txID
func (_m *Transactions) Has(txID flow.Identifier) bool {
	ret := _m.Called(txID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(flow.Identifier) bool); ok {
		r0 = rf(txID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Hash provides a mock function with given fields:
func (_m *Transactions) Hash() flow.Identifier {
	ret := _m.Called()

	var r0 flow.Identifier
	if rf, ok := ret.Get(0).(func() flow.Identifier); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(flow.Identifier)
		}
	}

	return r0
}

// Rem provides a mock function with given fields: txID
func (_m *Transactions) Rem(txID flow.Identifier) bool {
	ret := _m.Called(txID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(flow.Identifier) bool); ok {
		r0 = rf(txID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Size provides a mock function with given fields:
func (_m *Transactions) Size() uint {
	ret := _m.Called()

	var r0 uint
	if rf, ok := ret.Get(0).(func() uint); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint)
	}

	return r0
}
