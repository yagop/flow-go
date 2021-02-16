// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	flow "github.com/onflow/flow-go/model/flow"
	mock "github.com/stretchr/testify/mock"
)

// ReceiptValidator is an autogenerated mock type for the ReceiptValidator type
type ReceiptValidator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: receipts
func (_m *ReceiptValidator) Validate(receipts []*flow.ExecutionReceipt) error {
	ret := _m.Called(receipts)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*flow.ExecutionReceipt) error); ok {
		r0 = rf(receipts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidatePayload provides a mock function with given fields: candidate
func (_m *ReceiptValidator) ValidatePayload(candidate *flow.Block) error {
	ret := _m.Called(candidate)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.Block) error); ok {
		r0 = rf(candidate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
