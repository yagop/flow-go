// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocknetwork

import (
	cid "github.com/ipfs/go-cid"
	mock "github.com/stretchr/testify/mock"

	network "github.com/onflow/flow-go/network"
)

// BlockExchangeFetcher is an autogenerated mock type for the BlockExchangeFetcher type
type BlockExchangeFetcher struct {
	mock.Mock
}

// GetBlocks provides a mock function with given fields: cids
func (_m *BlockExchangeFetcher) GetBlocks(cids ...cid.Cid) network.BlocksPromise {
	_va := make([]interface{}, len(cids))
	for _i := range cids {
		_va[_i] = cids[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 network.BlocksPromise
	if rf, ok := ret.Get(0).(func(...cid.Cid) network.BlocksPromise); ok {
		r0 = rf(cids...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(network.BlocksPromise)
		}
	}

	return r0
}