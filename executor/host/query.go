package host

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	query "github.com/cosmos/cosmos-sdk/types/query"

	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"

	"github.com/initia-labs/opinit-bots-go/node/rpcclient"
)

func (h Host) GetAddress() sdk.AccAddress {
	return h.node.MustGetBroadcaster().GetAddress()
}

func (h Host) GetAddressStr() (string, error) {
	return h.node.MustGetBroadcaster().GetAddressString()
}

func (h Host) QueryLastOutput() (*ophosttypes.QueryOutputProposalResponse, error) {
	req := &ophosttypes.QueryOutputProposalsRequest{
		BridgeId: uint64(h.bridgeId),
		Pagination: &query.PageRequest{
			Limit:   1,
			Reverse: true,
		},
	}
	ctx, cancel := rpcclient.GetQueryContext(0)
	defer cancel()

	res, err := h.ophostQueryClient.OutputProposals(ctx, req)
	if err != nil {
		return nil, err
	}
	if res.OutputProposals == nil || len(res.OutputProposals) == 0 {
		return nil, nil
	}
	return &res.OutputProposals[0], nil
}

func (h Host) QueryOutput(outputIndex uint64) (*ophosttypes.QueryOutputProposalResponse, error) {
	req := &ophosttypes.QueryOutputProposalRequest{
		BridgeId:    uint64(h.bridgeId),
		OutputIndex: outputIndex,
	}
	ctx, cancel := rpcclient.GetQueryContext(0)
	defer cancel()

	return h.ophostQueryClient.OutputProposal(ctx, req)
}

func (h Host) QueryBatchInfos() (*ophosttypes.QueryBatchInfosResponse, error) {
	req := &ophosttypes.QueryBatchInfosRequest{
		BridgeId: uint64(h.bridgeId),
	}
	ctx, cancel := rpcclient.GetQueryContext(0)
	defer cancel()
	return h.ophostQueryClient.BatchInfos(ctx, req)
}