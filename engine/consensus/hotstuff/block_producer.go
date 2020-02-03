package hotstuff

import (
	"github.com/dapperlabs/flow-go/engine/consensus/hotstuff/types"
	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/module"
)

// BlockProposalProducer is responsible for producing new block proposals
type BlockProposalProducer struct {
	signer    Signer
	viewState ViewState
	builder   module.Builder

	// chainID is used for specifying the chainID field for new blocks
	chainID string
}

func NewBlockProposalProducer(signer Signer, viewState ViewState, builder module.Builder, chainID string) (*BlockProposalProducer, error) {
	bp := &BlockProposalProducer{
		signer:    signer,
		viewState: viewState,
		builder:   builder,
		chainID:   chainID,
	}
	return bp, nil
}

// MakeBlockProposal will build a proposal for the given view with the given QC
func (bp *BlockProposalProducer) MakeBlockProposal(view uint64, qcblock *types.QCBlock) *types.BlockProposal {
	block := bp.makeBlockForView(view, qcblock)

	unsignedBlockProposal := bp.propose(block)

	signedBlockProposal := bp.signBlockProposal(unsignedBlockProposal)

	return signedBlockProposal
}

// makeBlockForView gets the payload hash from mempool and build a block on top of the given qc for the given view.
func (bp *BlockProposalProducer) makeBlockForView(view uint64, qcblock *types.QCBlock) *types.Block {
	// TODO block should use flow.Identifier, bubble up error
	parentID := flow.HashToID(qcblock.Block.BlockMRH())
	payloadHash, err := bp.builder.BuildOn(parentID)
	if err != nil {
		panic(err)
	}

	// new block's height = parent.height + 1
	height := qcblock.Block.Height() + 1

	block := types.NewBlock(view, qcblock.QC, payloadHash[:], height, bp.chainID)
	return block
}

// propose get the consensus payload from mempool and make unsigned block proposal with the given block
func (bp *BlockProposalProducer) propose(block *types.Block) *types.UnsignedBlockProposal {
	// TODO consensus payload should not be included in HotStuff, addressed in https://github.com/dapperlabs/flow-go/pull/2365
	consensusPayload := types.ConsensusPayload{}
	unsignedBlockProposal := types.NewUnsignedBlockProposal(block, &consensusPayload)
	return unsignedBlockProposal
}

// signBlockProposal takes a unsigned proposal, signes it and returns a signed block proposal
func (bp *BlockProposalProducer) signBlockProposal(proposal *types.UnsignedBlockProposal) *types.BlockProposal {
	sig := bp.signer.SignBlockProposal(proposal, bp.viewState.GetSelfIdxForView(proposal.View()))

	blockProposal := types.NewBlockProposal(proposal.Block, proposal.ConsensusPayload, sig)
	return blockProposal
}
