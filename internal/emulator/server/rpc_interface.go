package server

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dapperlabs/bamboo-node/grpc/services/observe"
	"github.com/dapperlabs/bamboo-node/pkg/types"

	crypto "github.com/dapperlabs/bamboo-node/pkg/crypto/oldcrypto"
)

// Ping pings the Observation API server for a response.
func (s *EmulatorServer) Ping(ctx context.Context, req *observe.PingRequest) (*observe.PingResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// SendTransaction submits a transaction to the network.
func (s *EmulatorServer) SendTransaction(ctx context.Context, req *observe.SendTransactionRequest) (*observe.SendTransactionResponse, error) {
	txMsg := req.GetTransaction()
	payerSig := txMsg.GetPayerSignature()

	tx := &types.SignedTransaction{
		Script:       txMsg.GetScript(),
		Nonce:        txMsg.GetNonce(),
		ComputeLimit: txMsg.GetComputeLimit(),
		Timestamp:    time.Now(),
		PayerSignature: crypto.Signature{
			Account: crypto.BytesToAddress(payerSig.GetAccountAddress()),
			// TODO: update this (default signature for now)
			Sig: crypto.Sig{},
		},
		Status: types.TransactionPending,
	}

	s.transactionsIn <- tx
	response := &observe.SendTransactionResponse{
		Hash: tx.Hash().Bytes(),
	}

	return response, nil
}

// GetBlockByHash gets a block by hash.
func (s *EmulatorServer) GetBlockByHash(ctx context.Context, req *observe.GetBlockByHashRequest) (*observe.GetBlockByHashResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetBlockByNumber gets a block by number.
func (s *EmulatorServer) GetBlockByNumber(ctx context.Context, req *observe.GetBlockByNumberRequest) (*observe.GetBlockByNumberResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetLatestBlock gets the latest sealed block.
func (s *EmulatorServer) GetLatestBlock(ctx context.Context, req *observe.GetLatestBlockRequest) (*observe.GetLatestBlockResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetTransaction gets a transaction by hash.
func (s *EmulatorServer) GetTransaction(ctx context.Context, req *observe.GetTransactionRequest) (*observe.GetTransactionResponse, error) {
	hash := crypto.BytesToHash(req.GetHash())
	tx := s.blockchain.GetTransaction(hash)

	s.logger.WithFields(log.Fields{
		"txHash": hash,
	}).Debugf("💵  GetTransaction called: %s", hash)

	txMsg := &observe.GetTransactionResponse_Transaction{
		Script:       tx.Script,
		Nonce:        tx.Nonce,
		ComputeLimit: tx.ComputeLimit,
		ComputeUsed:  tx.ComputeUsed,
		PayerSignature: &observe.Signature{
			AccountAddress: tx.PayerSignature.Account.Bytes(),
			// TODO: update this (default signature bytes for now)
			Signature: tx.PayerSignature.Sig[:],
		},
		Status: observe.GetTransactionResponse_Transaction_Status(tx.Status),
	}

	response := &observe.GetTransactionResponse{
		Transaction: txMsg,
	}

	return response, nil
}

// GetAccount returns the info associated with an address.
func (s *EmulatorServer) GetAccount(ctx context.Context, req *observe.GetAccountRequest) (*observe.GetAccountResponse, error) {
	address := crypto.BytesToAddress(req.GetAddress())
	account := s.blockchain.GetAccount(address)

	s.logger.WithFields(log.Fields{
		"address": address,
	}).Debugf("👤  GetAccount called: %s", address)

	accMsg := &observe.GetAccountResponse_Account{
		Address:    account.Address.Bytes(),
		Balance:    account.Balance,
		Code:       account.Code,
		PublicKeys: account.PublicKeys,
	}

	response := &observe.GetAccountResponse{
		Account: accMsg,
	}

	return response, nil
}

// CallContract performs a contract call.
func (s *EmulatorServer) CallContract(ctx context.Context, req *observe.CallContractRequest) (*observe.CallContractResponse, error) {
	s.logger.Debug("📞  Contract script called")

	script := req.GetScript()
	value, _ := s.blockchain.CallScript(script)
	// TODO: add error handling besides just this
	if value == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid script")
	}

	// TODO: change this to whatever interface -> byte encoding decided on
	valueMsg := []byte(fmt.Sprintf("%v", value.(interface{})))

	response := &observe.CallContractResponse{
		Script: valueMsg,
	}

	return response, nil
}
