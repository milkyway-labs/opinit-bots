package host

import (
	comettypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// If the relay oracle is enabled and the extended commit info contains votes, create a new MsgUpdateOracle message.
// Else return nil.
func (h *Host) oracleTxHandler(blockHeight int64, extCommitBz comettypes.Tx) (sdk.Msg, error) {
	if !h.OracleEnabled() {
		return nil, nil
	}
	//if !h.monitor.IsOurTurn() {
	//	h.Logger().Info("it's not our turn to update oracle, skipping")
	//	return nil, nil
	//}
	return h.child.GetMsgUpdateOracle(
		blockHeight,
		extCommitBz,
	)
}
