package host

import (
	"context"
	"fmt"

	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	"go.uber.org/zap"

	"github.com/initia-labs/opinit-bots/executor/types"
	nodetypes "github.com/initia-labs/opinit-bots/node/types"
	hostprovider "github.com/initia-labs/opinit-bots/provider/host"
)

func (h *Host) recordBatchHandler(_ context.Context, args nodetypes.EventHandlerArgs) error {
	submitter, err := hostprovider.ParseMsgRecordBatch(args.EventAttributes)
	if err != nil {
		return err
	}

	if submitter != h.BridgeInfo().BridgeConfig.BatchInfo.Submitter {
		return nil
	}
	h.Logger().Info("record batch",
		zap.String("submitter", submitter),
	)
	return nil
}

func (h *Host) recordBatchMsgHandler(_ context.Context, msg *ophosttypes.MsgRecordBatch) error {
	bridgeInfo := h.BridgeInfo()
	if msg.BridgeId != bridgeInfo.BridgeId || msg.Submitter != bridgeInfo.BridgeConfig.BatchInfo.Submitter {
		return nil
	}

	// We're only interested in BatchDataTypeHeader.
	if msg.BatchBytes[0] != byte(types.BatchDataTypeHeader) {
		return nil
	}
	header, err := types.UnmarshalBatchDataHeader(msg.BatchBytes)
	if err != nil {
		return fmt.Errorf("unmarshal batch data header: %w", err)
	}

	h.batch.UpdateLastSubmittedBatchEndBlockNumber(int64(header.End))
	h.Logger().Info("updated last submitted batch end block number",
		zap.Uint64("height", header.End))
	return nil
}

func (h *Host) updateBatchInfoHandler(_ context.Context, args nodetypes.EventHandlerArgs) error {
	bridgeId, submitter, chain, outputIndex, l2BlockNumber, err := hostprovider.ParseMsgUpdateBatchInfo(args.EventAttributes)
	if err != nil {
		return err
	}
	if bridgeId != h.BridgeId() {
		// pass other bridge deposit event
		return nil
	}

	h.Logger().Info("update batch info",
		zap.String("chain", chain),
		zap.String("submitter", submitter),
		zap.Uint64("output_index", outputIndex),
		zap.Int64("l2_block_number", l2BlockNumber),
	)

	h.batch.UpdateBatchInfo(chain, submitter, outputIndex, l2BlockNumber)
	return nil
}
