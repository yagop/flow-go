package utils

import (
	bambooProto "github.com/dapperlabs/bamboo-node/grpc/shared"
	"github.com/dapperlabs/bamboo-node/internal/types"
	"github.com/dapperlabs/bamboo-node/pkg/crypto"
)

func (m *bambooProto.Register) FromMessage() *types.Register {
	return &types.Register{
		ID:    m.GetId(),
		Value: m.GetValue(),
	}
}

func (t *types.Register) ToMessage() *bambooProto.Register {
	return &bambooProto.Register{
		Id:    t.ID,
		Value: t.Value,
	}
}

func (m *bambooProto.IntermediateRegisters) FromMessage() *types.IntermediateRegisters {
	registers := make([]types.Register, 0)
	for _, r := range m.GetRegisters() {
		registers = append(registers, *r.FromMessage())
	}

	return &types.IntermediateRegisters{
		TransactionHash: crypto.BytesToHash(m.GetTransactionHash()),
		Registers:       registers,
		ComputeUsed:     m.GetComputeUsed(),
	}
}

func (t *types.IntermediateRegisters) ToMessage() *bambooProto.IntermediateRegisters {
	registers := make([]*bambooProto.Register, 0)
	for _, r := range t.Registers {
		registers = append(registers, r.ToMessage())
	}

	return &bambooProto.IntermediateRegisters{
		TransactionHash: t.TransactionHash.Bytes(),
		Registers:       registers,
		ComputeUsed:     t.ComputeUsed,
	}
}

func (m *bambooProto.TransactionRegister) FromMessage() *types.TransactionRegister {
	return &types.TransactionRegister{}
}

func (t *types.TransactionRegister) ToMessage() *bambooProto.TransactionRegister {
	return &bambooProto.TransactionRegister{}
}

func (m *bambooProto.Collection) FromMessage() *types.Collection {
	return &types.Collection{}
}

func (t *types.Collection) ToMessage() *bambooProto.Collection {
	return &bambooProto.Collection{}
}

func (m *bambooProto.SignedCollectionHash) FromMessage() *types.SignedCollectionHash {
	return &types.SignedCollectionHash{}
}

func (t *types.SignedCollectionHash) ToMessage() *bambooProto.SignedCollectionHash {
	return &bambooProto.SignedCollectionHash{}
}

func (m *bambooProto.Block) FromMessage() *types.Block {
	return &types.Block{}
}

func (t *types.Block) ToMessage() *bambooProto.Block {
	return &bambooProto.Block{}
}

func (m *bambooProto.BlockSeal) FromMessage() *types.BlockSeal {
	return &types.BlockSeal{}
}

func (t *types.BlockSeal) ToMessage() *bambooProto.BlockSeal {
	return &bambooProto.BlockSeal{}
}

func (m *bambooProto.Transaction) FromMessage() *types.Transaction {
	return &types.Transaction{}
}

func (t *types.Transaction) ToMessage() *bambooProto.Transaction {
	return &bambooProto.Transaction{}
}

func (m *bambooProto.SignedTransaction) FromMessage() *types.SignedTransaction {
	return &types.SignedTransaction{}
}

func (t *types.SignedTransaction) ToMessage() *bambooProto.SignedTransaction {
	return &bambooProto.SignedTransaction{}
}

func (m *bambooProto.ExecutionReceipt) FromMessage() *types.ExecutionReceipt {
	return &types.ExecutionReceipt{}
}

func (t *types.ExecutionReceipt) ToMessage() *bambooProto.ExecutionReceipt {
	return &bambooProto.ExecutionReceipt{}
}

func (m *bambooProto.InvalidExecutionReceiptChallenge) FromMessage() *types.InvalidExecutionReceiptChallenge {
	partTransactions := make([]types.IntermediateRegisters, 0)
	for _, r := range m.GetPartTransactions() {
		partTransactions = append(partTransactions, *r.FromMessage())
	}

	return &types.InvalidExecutionReceiptChallenge{
		ExecutionReceiptHash:      crypto.BytesToHash(m.GetExecutionReceiptHash()),
		ExecutionReceiptSignature: crypto.BytesToSig(m.GetExecutionReceiptSignature()),
		PartIndex:                 m.GetPartIndex(),
		PartTransactions:          partTransactions,
		Signature:                 crypto.BytesToSig(m.GetSignature()),
	}
}

func (t *types.InvalidExecutionReceiptChallenge) ToMessage() *bambooProto.InvalidExecutionReceiptChallenge {
	partTransactions := make([]*bambooProto.IntermediateRegisters, 0)
	for _, r := range t.PartTransactions {
		partTransactions = append(partTransactions, r.ToMessage())
	}

	return &bambooProto.InvalidExecutionReceiptChallenge{
		ExecutionReceiptHash:      t.ExecutionReceiptHash.Bytes(),
		ExecutionReceiptSignature: t.ExecutionReceiptSignature.Bytes(),
		PartIndex:                 t.PartIndex,
		PartTransactions:          partTransactions,
		Signature:                 t.Signature.Bytes(),
	}
}

func (m *bambooProto.ResultApproval) FromMessage() *types.ResultApproval {
	return &types.ResultApproval{
		BlockHeight:             m.GetBlockHeight(),
		ExecutionReceiptHash:    crypto.BytesToHash(m.GetExecutionReceiptHash()),
		ResultApprovalSignature: crypto.BytesToSig(m.GetResultApprovalSignature()),
		Proof:                   m.GetProof(),
		Signature:               crypto.BytesToSig(m.GetSignature()),
	}
}

func (t *types.ResultApproval) ToMessage() *bambooProto.ResultApproval {
	return &bambooProto.ResultApproval{
		BlockHeight:             t.BlockHeight,
		ExecutionReceiptHash:    t.ExecutionReceiptHash.Bytes(),
		ResultApprovalSignature: t.ResultApprovalSignature.Bytes(),
		Proof:                   t.Proof,
		Signature:               t.Signature.Bytes(),
	}
}
