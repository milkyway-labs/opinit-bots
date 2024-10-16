package child

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	"go.uber.org/zap"

	dbtypes "github.com/initia-labs/opinit-bots/db/types"
	executortypes "github.com/initia-labs/opinit-bots/executor/types"
	nodetypes "github.com/initia-labs/opinit-bots/node/types"
	childprovider "github.com/initia-labs/opinit-bots/provider/child"
	"github.com/initia-labs/opinit-bots/types"
)

func (ch *Child) initiateWithdrawalHandler(_ context.Context, args nodetypes.EventHandlerArgs) error {
	l2Sequence, amount, from, to, baseDenom, err := childprovider.ParseInitiateWithdrawal(args.EventAttributes)
	if err != nil {
		return err
	}
	return ch.handleInitiateWithdrawal(l2Sequence, from, to, baseDenom, amount)
}

func (ch *Child) handleInitiateWithdrawal(l2Sequence uint64, from string, to string, baseDenom string, amount uint64) error {
	withdrawalHash := ophosttypes.GenerateWithdrawalHash(ch.BridgeId(), l2Sequence, from, to, baseDenom, amount)
	data := executortypes.WithdrawalData{
		Sequence:       l2Sequence,
		From:           from,
		To:             to,
		Amount:         amount,
		BaseDenom:      baseDenom,
		WithdrawalHash: withdrawalHash[:],
	}

	// store to database
	err := ch.SetWithdrawal(l2Sequence, data)
	if err != nil {
		return err
	}

	// generate merkle tree
	err = ch.Merkle().InsertLeaf(withdrawalHash[:])
	if err != nil {
		return err
	}

	ch.Logger().Info("initiate token withdrawal",
		zap.Uint64("l2_sequence", l2Sequence),
		zap.String("from", from),
		zap.String("to", to),
		zap.Uint64("amount", amount),
		zap.String("base_denom", baseDenom),
		zap.String("withdrawal", base64.StdEncoding.EncodeToString(withdrawalHash[:])),
	)

	return nil
}

func (ch *Child) prepareTree(blockHeight int64) error {
	if ch.InitializeTree(blockHeight) {
		return nil
	}

	err := ch.Merkle().LoadWorkingTree(types.MustInt64ToUint64(blockHeight) - 1)
	if errors.Is(err, dbtypes.ErrNotFound) {
		// must not happen
		panic(fmt.Errorf("working tree not found at height: %d, current: %d", blockHeight-1, blockHeight))
	} else if err != nil {
		return err
	}

	return nil
}

func (ch *Child) prepareOutput(ctx context.Context, blockHeight int64) (retry bool, err error) {
	workingOutputIndex := ch.Merkle().GetWorkingTreeIndex()

	// initialize next output time
	if ch.nextOutputTime.IsZero() && workingOutputIndex > 1 {
		output, err := ch.host.QueryOutput(ctx, ch.BridgeId(), workingOutputIndex-1, 0)
		if err != nil {
			// TODO: maybe not return error here and roll back
			return false, fmt.Errorf("output does not exist at index: %d", workingOutputIndex-1)
		}
		ch.lastOutputTime = output.OutputProposal.L1BlockTime
		ch.nextOutputTime = output.OutputProposal.L1BlockTime.Add(ch.BridgeInfo().BridgeConfig.SubmissionInterval * 2 / 3)
	}

	output, err := ch.host.QueryOutput(ctx, ch.BridgeId(), ch.Merkle().GetWorkingTreeIndex(), 0)
	if err != nil {
		if strings.Contains(err.Error(), "collections: not found") {
			return false, nil
		}
		return false, err
	} else {
		// we are syncing
		finalizingBlockHeight := types.MustUint64ToInt64(output.OutputProposal.L2BlockNumber)
		if finalizingBlockHeight < blockHeight {
			// Another operator has already proposed an output.
			// We need to reconcile trees.
			finalizedBlockHeight := finalizingBlockHeight
			lastL2Sequence := ch.Merkle().GetStartLeafIndex() + ch.Merkle().GetWorkingTreeLeafCount() - 1
			err = ch.reconcileTrees(ctx, finalizedBlockHeight, lastL2Sequence)
			if err != nil {
				return false, fmt.Errorf("reconcile trees: %w", err)
			}
			return true, nil // Retry prepareOutput
		} else {
			ch.finalizingBlockHeight = finalizingBlockHeight
			ch.Logger().Info("we're syncing", zap.Int64("finalizing_block_height", ch.finalizingBlockHeight))
		}
	}
	return false, nil
}

func (ch *Child) handleTree(blockHeight int64, latestHeight int64, blockId []byte, blockHeader cmtproto.Header) (kvs []types.RawKV, storageRoot []byte, err error) {
	// panic if we are syncing and passed the finalizing block height
	// this must not happened
	if ch.finalizingBlockHeight != 0 && ch.finalizingBlockHeight < blockHeight {
		panic(fmt.Errorf("INVARIANT failed; handleTree expect to finalize tree at block `%d` but we got block `%d`", blockHeight-1, blockHeight))
	}

	// finalize working tree if we are fully synced or block time is over next output time(only if it's our turn)
	if ch.finalizingBlockHeight == blockHeight ||
		(ch.monitor.IsOurTurn() &&
			ch.finalizingBlockHeight == 0 &&
			blockHeight == latestHeight &&
			blockHeader.Time.After(ch.nextOutputTime)) {
		kvs, storageRoot, err = ch.finalizeTree(blockHeight, blockId)

		// skip output submission when it is already submitted
		if ch.finalizingBlockHeight == blockHeight {
			storageRoot = nil
		}

		ch.finalizingBlockHeight = 0
		ch.lastOutputTime = blockHeader.Time
		ch.nextOutputTime = blockHeader.Time.Add(ch.BridgeInfo().BridgeConfig.SubmissionInterval * 2 / 3)
	}

	version := types.MustInt64ToUint64(blockHeight)
	err = ch.Merkle().SaveWorkingTree(version)
	if err != nil {
		return nil, nil, err
	}

	return kvs, storageRoot, nil
}

func (ch *Child) finalizeTree(blockHeight int64, blockId []byte) (kvs []types.RawKV, storageRoot []byte, err error) {
	data, err := json.Marshal(executortypes.TreeExtraData{
		BlockNumber: blockHeight,
		BlockHash:   blockId,
	})
	if err != nil {
		return nil, nil, err
	}

	kvs, storageRoot, err = ch.Merkle().FinalizeWorkingTree(data)
	if err != nil {
		return nil, nil, err
	}

	ch.Logger().Info("finalize working tree",
		zap.Uint64("tree_index", ch.Merkle().GetWorkingTreeIndex()),
		zap.Int64("height", blockHeight),
		zap.Uint64("num_leaves", ch.Merkle().GetWorkingTreeLeafCount()),
		zap.String("storage_root", base64.StdEncoding.EncodeToString(storageRoot)),
	)

	return kvs, storageRoot, err
}

func (ch *Child) handleOutput(blockHeight int64, version uint8, blockId []byte, outputIndex uint64, storageRoot []byte) error {
	if !ch.monitor.IsOurTurn() {
		ch.Logger().Info("it's not our turn to propose output, skipping")
		return nil
	}

	outputRoot := ophosttypes.GenerateOutputRoot(version, storageRoot, blockId)
	msg, err := ch.host.GetMsgProposeOutput(
		ch.BridgeId(),
		outputIndex,
		blockHeight,
		outputRoot[:],
	)
	if err != nil {
		return err
	} else if msg != nil {
		ch.AppendMsgQueue(msg)
	}
	return nil
}

func (ch *Child) reconcileTrees(ctx context.Context, finalizedBlockHeight int64, lastL2Sequence uint64) error {
	// Load the old working tree at the finalized output's block height.
	version := types.MustInt64ToUint64(finalizedBlockHeight)
	err := ch.Merkle().LoadWorkingTree(version)
	if err != nil {
		return fmt.Errorf("load working tree at height %d: %w", finalizedBlockHeight, err)
	}

	// Finalize the old tree. We ignore the returned storage root since we
	// are not going to propose it(it's already proposed).
	block, err := ch.Node().GetRPCClient().Block(ctx, &finalizedBlockHeight)
	if err != nil {
		return fmt.Errorf("get block at height %d: %w", finalizedBlockHeight, err)
	}
	kvs, _, err := ch.finalizeTree(finalizedBlockHeight, block.BlockID.Hash)
	if err != nil {
		return fmt.Errorf("finalize tree at height %d: %w", finalizedBlockHeight, err)
	}
	err = ch.DB().RawBatchSet(kvs...)
	if err != nil {
		return err
	}

	// Save the old tree and re-load the tree.
	// This will effectively increment the working tree index and start leaf index.
	err = ch.Merkle().SaveWorkingTree(version)
	if err != nil {
		return fmt.Errorf("save working tree at height %d: %w", version, err)
	}
	err = ch.Merkle().LoadWorkingTree(version)
	if err != nil {
		return fmt.Errorf("reload working tree at height %d: %w", version, err)
	}

	nextSequence := ch.Merkle().GetStartLeafIndex()

	err = ch.DB().PrefixedIterate(executortypes.WithdrawalKey, func(key, value []byte) (bool, error) {
		sequence := dbtypes.ToUint64Key(key[len(key)-8:])
		if sequence >= nextSequence && sequence <= lastL2Sequence {
			var data executortypes.WithdrawalData
			err = json.Unmarshal(value, &data)
			if err != nil {
				return true, err
			}
			err = ch.Merkle().InsertLeaf(data.WithdrawalHash)
			if err != nil {
				return true, err
			}
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("re-insert leaves: %w", err)
	}

	ch.Logger().Info("reconciled trees",
		zap.Int64("finalized_block_height", finalizedBlockHeight),
		zap.Uint64("from_sequence", nextSequence),
		zap.Uint64("to_sequence", lastL2Sequence))

	return nil
}

// GetWithdrawal returns the withdrawal data for the given sequence from the database
func (ch *Child) GetWithdrawal(sequence uint64) (executortypes.WithdrawalData, error) {
	dataBytes, err := ch.DB().Get(executortypes.PrefixedWithdrawalKey(sequence))
	if err != nil {
		return executortypes.WithdrawalData{}, err
	}
	var data executortypes.WithdrawalData
	err = json.Unmarshal(dataBytes, &data)
	return data, err
}

// SetWithdrawal store the withdrawal data for the given sequence to the database
func (ch *Child) SetWithdrawal(sequence uint64, data executortypes.WithdrawalData) error {
	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	return ch.DB().Set(executortypes.PrefixedWithdrawalKey(sequence), dataBytes)
}

func (ch *Child) DeleteFutureWithdrawals(fromSequence uint64) error {
	return ch.DB().PrefixedIterate(executortypes.WithdrawalKey, func(key, _ []byte) (bool, error) {
		sequence := dbtypes.ToUint64Key(key[len(key)-8:])
		if sequence >= fromSequence {
			err := ch.DB().Delete(key)
			if err != nil {
				return true, err
			}
		}
		return false, nil
	})
}
