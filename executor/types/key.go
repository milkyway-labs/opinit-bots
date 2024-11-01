package types

import (
	dbtypes "github.com/initia-labs/opinit-bots/db/types"
)

var (
	WithdrawalKey = []byte("withdrawal")
	DepositKey    = []byte("deposit")
)

func PrefixedWithdrawalKey(sequence uint64) []byte {
	return append(append(WithdrawalKey, dbtypes.Splitter), dbtypes.FromUint64Key(sequence)...)
}

func PrefixedDepositKey(sequence uint64) []byte {
	return append(append(DepositKey, dbtypes.Splitter), dbtypes.FromUint64Key(sequence)...)
}
