package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// NewMsgExec creates a new authz MsgExec instance.
// The implementation is copied from the authz module, but NewMsgExec accepts
// string address instead of sdk.AccAddress.
func NewMsgExec(grantee string, msgs []sdk.Msg) (*authz.MsgExec, error) {
	msgsAny := make([]*cdctypes.Any, len(msgs))
	for i, msg := range msgs {
		any, err := cdctypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
		msgsAny[i] = any
	}
	return &authz.MsgExec{
		Grantee: grantee,
		Msgs:    msgsAny,
	}, nil
}
