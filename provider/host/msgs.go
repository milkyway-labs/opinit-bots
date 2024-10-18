package host

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"

	"github.com/initia-labs/opinit-bots/types"
)

func (b BaseHost) GetMsgProposeOutput(
	bridgeId uint64,
	outputIndex uint64,
	l2BlockNumber int64,
	outputRoot []byte,
) (sdk.Msg, error) {
	sender, err := b.GetAddressStr()
	if err != nil {
		if errors.Is(err, types.ErrKeyNotSet) {
			return nil, nil
		}
		return nil, err
	}

	msg := ophosttypes.NewMsgProposeOutput(
		// The L2 output submitter grants permission to the operator's output submitter
		b.bridgeInfo.BridgeConfig.Proposer,
		bridgeId,
		outputIndex,
		types.MustInt64ToUint64(l2BlockNumber),
		outputRoot,
	)
	err = msg.Validate(b.node.AccountCodec())
	if err != nil {
		return nil, err
	}
	execMsg, err := types.NewMsgExec(sender, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	return execMsg, nil
}

func (b BaseHost) CreateBatchMsg(batchBytes []byte) (sdk.Msg, error) {
	submitter, err := b.GetAddressStr()
	if err != nil {
		if errors.Is(err, types.ErrKeyNotSet) {
			return nil, nil
		}
		return nil, err
	}

	msg := ophosttypes.NewMsgRecordBatch(
		// The L2 batch submitter grants permission to the operator's batch submitter
		b.bridgeInfo.BridgeConfig.BatchInfo.Submitter,
		b.BridgeId(),
		batchBytes,
	)
	err = msg.Validate(b.node.AccountCodec())
	if err != nil {
		return nil, err
	}
	execMsg, err := types.NewMsgExec(submitter, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	return execMsg, nil
}
