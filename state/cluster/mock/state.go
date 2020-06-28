// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import cluster "github.com/dapperlabs/flow-go/state/cluster"
import flow "github.com/dapperlabs/flow-go/model/flow"
import mock "github.com/stretchr/testify/mock"

// State is an autogenerated mock type for the State type
type State struct {
	mock.Mock
}

// AtBlockID provides a mock function with given fields: blockID
func (_m *State) AtBlockID(blockID flow.Identifier) cluster.Snapshot {
	ret := _m.Called(blockID)

	var r0 cluster.Snapshot
	if rf, ok := ret.Get(0).(func(flow.Identifier) cluster.Snapshot); ok {
		r0 = rf(blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cluster.Snapshot)
		}
	}

	return r0
}

// Final provides a mock function with given fields:
func (_m *State) Final() cluster.Snapshot {
	ret := _m.Called()

	var r0 cluster.Snapshot
	if rf, ok := ret.Get(0).(func() cluster.Snapshot); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cluster.Snapshot)
		}
	}

	return r0
}

// Mutate provides a mock function with given fields:
func (_m *State) Mutate() cluster.Mutator {
	ret := _m.Called()

	var r0 cluster.Mutator
	if rf, ok := ret.Get(0).(func() cluster.Mutator); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cluster.Mutator)
		}
	}

	return r0
}
