package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	btypes "github.com/initia-labs/opinit-bots/node/broadcaster/types"
)

func (m *Monitor) handleEpoch(ctx context.Context) error {
	// Query the current block time.
	rpcClient := m.Node().GetRPCClient()
	headerRes, err := rpcClient.Header(ctx, nil)
	if err != nil {
		return fmt.Errorf("query abci info: %w", err)
	}
	blockTime := headerRes.Header.Time

	currentEpoch, err := m.queryCurrentEpoch(ctx)
	if err != nil {
		return fmt.Errorf("query current epoch: %w", err)
	}

	// If EpochEndsAt is not set or the current time is after it, advance the epoch.
	// Note advancing epoch is permissionless.
	if currentEpoch.EpochEndsAt == nil || !blockTime.Before(time.Unix(0, int64(*currentEpoch.EpochEndsAt))) {
		m.Logger().Info("advancing epoch")
		err = m.advanceEpoch(ctx)
		if err != nil {
			return fmt.Errorf("advance epoch: %w", err)
		}
		// Exit and check if the current operator is me in the next block.
		return nil
	}

	isOurTurnBefore := m.isOurTurn.Load()
	isOurTurn := currentEpoch.OperatorID != nil && *currentEpoch.OperatorID == m.operatorID
	m.isOurTurn.Store(isOurTurn)

	if isOurTurnBefore != isOurTurn {
		m.Logger().Info("turn changed", zap.Bool("is_our_turn", isOurTurn))
	}
	return nil
}

// currentEpochResponse is used to unmarshal the response from the query.
type currentEpochResponse struct {
	//EpochID     uint64  `json:"epoch_id"`
	OperatorID *uint32 `json:"operator_id"`
	//Operator    *string `json:"operator"`
	EpochEndsAt *uint64 `json:"epoch_ends_at,string"`
}

func (m Monitor) queryCurrentEpoch(ctx context.Context) (currentEpochResponse, error) {
	// We don't need to specify query height since we want the latest info.
	resp, err := m.wasmQueryClient.SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   m.contractAddr,
		QueryData: []byte(`{"current_epoch":{}}`),
	})
	if err != nil {
		return currentEpochResponse{}, fmt.Errorf("query current_epoch: %w", err)
	}

	var res currentEpochResponse
	err = json.Unmarshal(resp.Data, &res)
	if err != nil {
		return currentEpochResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}
	return res, nil
}

func (m *Monitor) advanceEpoch(ctx context.Context) error {
	b := m.Node().MustGetBroadcaster()
	sender, err := b.GetAddressString()
	if err != nil {
		return fmt.Errorf("get sender: %w", err)
	}
	m.Node().MustGetBroadcaster().BroadcastMsgs(btypes.ProcessedMsgs{
		Msgs: []sdk.Msg{
			&wasmtypes.MsgExecuteContract{
				Sender:   sender,
				Contract: m.contractAddr,
				Msg:      []byte(`{"advance_epoch":{}}`),
			},
		},
		Timestamp: time.Now().UnixNano(),
		Save:      false, // We don't need to retry
	})
	return nil
}
