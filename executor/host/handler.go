package host

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"

	btypes "github.com/initia-labs/opinit-bots/node/broadcaster/types"
	nodetypes "github.com/initia-labs/opinit-bots/node/types"
	"github.com/initia-labs/opinit-bots/txutils"
	"github.com/initia-labs/opinit-bots/types"
)

func (h *Host) beginBlockHandler(_ context.Context, args nodetypes.BeginBlockArgs) error {
	h.EmptyMsgQueue()
	h.EmptyProcessedMsgs()
	return nil
}

func (h *Host) endBlockHandler(_ context.Context, args nodetypes.EndBlockArgs) error {
	// collect more msgs if block height is not latest
	blockHeight := args.Block.Header.Height
	msgQueue := h.GetMsgQueue()

	batchKVs := []types.RawKV{
		h.Node().SyncInfoToRawKV(blockHeight),
	}
	if h.Node().HasBroadcaster() {
		if len(msgQueue) != 0 {
			h.AppendProcessedMsgs(btypes.ProcessedMsgs{
				Msgs:      msgQueue,
				Timestamp: time.Now().UnixNano(),
				Save:      true,
			})
		}

		msgkvs, err := h.child.ProcessedMsgsToRawKV(h.GetProcessedMsgs(), false)
		if err != nil {
			return err
		}
		batchKVs = append(batchKVs, msgkvs...)
	}

	err := h.DB().RawBatchSet(batchKVs...)
	if err != nil {
		return err
	}

	for _, processedMsg := range h.GetProcessedMsgs() {
		h.child.BroadcastMsgs(processedMsg)
	}
	return nil
}

func (h *Host) txHandler(ctx context.Context, args nodetypes.TxHandlerArgs) error {
	if args.BlockHeight == args.LatestHeight && args.TxIndex == 0 {
		if msg, err := h.oracleTxHandler(args.BlockHeight, args.Tx); err != nil {
			return err
		} else if msg != nil {
			h.AppendProcessedMsgs(btypes.ProcessedMsgs{
				Msgs:      []sdk.Msg{msg},
				Timestamp: time.Now().UnixNano(),
				Save:      false,
			})
		}
	}

	txConfig := h.Node().GetTxConfig()
	tx, err := txutils.DecodeTx(txConfig, args.Tx)
	if err != nil {
		// if tx has other messages than MsgRecordBatch, tx parse error is expected
		// ignore decoding error
		return nil
	}
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		// we only expect one message for MsgRecordBatch tx
		return nil
	}
	msg, ok := msgs[0].(*ophosttypes.MsgRecordBatch)
	if !ok {
		// We handle MsgExec too.
		execMsg, ok := msgs[0].(*authz.MsgExec)
		if !ok {
			return nil
		}
		if len(execMsg.Msgs) != 1 {
			return nil
		}
		var tmpMsg sdk.Msg
		err = h.Node().GetCodec().UnpackAny(execMsg.Msgs[0], &tmpMsg)
		if err != nil {
			return fmt.Errorf("unpack exec msg: %w", err)
		}
		msg, ok = tmpMsg.(*ophosttypes.MsgRecordBatch)
		if !ok {
			return nil
		}
	}
	err = h.recordBatchMsgHandler(ctx, msg)
	if err != nil {
		return fmt.Errorf("handle record batch msg: %w", err)
	}
	return nil
}
