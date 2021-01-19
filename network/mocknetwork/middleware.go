// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocknetwork

import (
	flow "github.com/onflow/flow-go/model/flow"
	message "github.com/onflow/flow-go/network/message"

	mock "github.com/stretchr/testify/mock"

	network "github.com/onflow/flow-go/network"

	time "time"
)

// Middleware is an autogenerated mock type for the Middleware type
type Middleware struct {
	mock.Mock
}

// Ping provides a mock function with given fields: targetID
func (_m *Middleware) Ping(targetID flow.Identifier) (time.Duration, error) {
	ret := _m.Called(targetID)

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func(flow.Identifier) time.Duration); ok {
		r0 = rf(targetID)
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Identifier) error); ok {
		r1 = rf(targetID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Publish provides a mock function with given fields: msg, channel
func (_m *Middleware) Publish(msg *message.Message, channel network.Channel) error {
	ret := _m.Called(msg, channel)

	var r0 error
	if rf, ok := ret.Get(0).(func(*message.Message, network.Channel) error); ok {
		r0 = rf(msg, channel)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Send provides a mock function with given fields: channel, msg, targetIDs
func (_m *Middleware) Send(channel network.Channel, msg *message.Message, targetIDs ...flow.Identifier) error {
	_va := make([]interface{}, len(targetIDs))
	for _i := range targetIDs {
		_va[_i] = targetIDs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, channel, msg)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(network.Channel, *message.Message, ...flow.Identifier) error); ok {
		r0 = rf(channel, msg, targetIDs...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendDirect provides a mock function with given fields: msg, targetID
func (_m *Middleware) SendDirect(msg *message.Message, targetID flow.Identifier) error {
	ret := _m.Called(msg, targetID)

	var r0 error
	if rf, ok := ret.Get(0).(func(*message.Message, flow.Identifier) error); ok {
		r0 = rf(msg, targetID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields: overlay
func (_m *Middleware) Start(overlay network.Overlay) error {
	ret := _m.Called(overlay)

	var r0 error
	if rf, ok := ret.Get(0).(func(network.Overlay) error); ok {
		r0 = rf(overlay)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stop provides a mock function with given fields:
func (_m *Middleware) Stop() {
	_m.Called()
}

// Subscribe provides a mock function with given fields: channel
func (_m *Middleware) Subscribe(channel network.Channel) error {
	ret := _m.Called(channel)

	var r0 error
	if rf, ok := ret.Get(0).(func(network.Channel) error); ok {
		r0 = rf(channel)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unsubscribe provides a mock function with given fields: channel
func (_m *Middleware) Unsubscribe(channel network.Channel) error {
	ret := _m.Called(channel)

	var r0 error
	if rf, ok := ret.Get(0).(func(network.Channel) error); ok {
		r0 = rf(channel)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAllowList provides a mock function with given fields:
func (_m *Middleware) UpdateAllowList() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
