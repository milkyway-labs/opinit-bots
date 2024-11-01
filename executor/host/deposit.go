package host

import (
	"context"
	"errors"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dbtypes "github.com/initia-labs/opinit-bots/db/types"
	executortypes "github.com/initia-labs/opinit-bots/executor/types"
	nodetypes "github.com/initia-labs/opinit-bots/node/types"
	hostprovider "github.com/initia-labs/opinit-bots/provider/host"
)

func (h *Host) initiateDepositHandler(_ context.Context, args nodetypes.EventHandlerArgs) error {
	bridgeId, l1Sequence, from, to, l1Denom, l2Denom, amount, data, err := hostprovider.ParseMsgInitiateDeposit(args.EventAttributes)
	if err != nil {
		return err
	}
	if bridgeId != h.BridgeId() {
		// pass other bridge deposit event
		return nil
	}
	if l1Sequence < h.initialL1Sequence {
		// pass old deposit event
		return nil
	}

	h.depositQueue = append(h.depositQueue, executortypes.Deposit{
		Sequence:    l1Sequence,
		BlockHeight: args.BlockHeight,
		From:        from,
		To:          to,
		L1Denom:     l1Denom,
		L2Denom:     l2Denom,
		Amount:      amount,
		Data:        data,
	})
	return nil
}

func (h *Host) handleInitiateDeposit(
	l1Sequence uint64,
	blockHeight int64,
	from string,
	to string,
	l1Denom string,
	l2Denom string,
	amount string,
	data []byte,
) (sdk.Msg, error) {
	coinAmount, ok := math.NewIntFromString(amount)
	if !ok {
		return nil, errors.New("invalid amount")
	}
	coin := sdk.NewCoin(l2Denom, coinAmount)

	return h.child.GetMsgFinalizeTokenDeposit(
		from,
		to,
		coin,
		l1Sequence,
		blockHeight,
		l1Denom,
		data,
	)
}

func (h *Host) pruneFinalizedDeposits(lastFinalizedSequence uint64) error {
	var prunedSequences []uint64
	err := h.DB().PrefixedIterate(executortypes.DepositKey, func(key, value []byte) (bool, error) {
		sequence := dbtypes.ToUint64Key(key[len(key)-8:])
		if sequence <= lastFinalizedSequence {
			err := h.DB().Delete(key)
			prunedSequences = append(prunedSequences, sequence)
			return false, err
		}
		return false, nil
	})
	return err
}
