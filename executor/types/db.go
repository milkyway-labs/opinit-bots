package types

type WithdrawalData struct {
	Sequence       uint64 `json:"sequence"`
	From           string `json:"from"`
	To             string `json:"to"`
	Amount         uint64 `json:"amount"`
	BaseDenom      string `json:"base_denom"`
	WithdrawalHash []byte `json:"withdrawal_hash"`
}

type TreeExtraData struct {
	BlockNumber int64  `json:"block_number"`
	BlockHash   []byte `json:"block_hash"`
}

type Deposit struct {
	Sequence    uint64 `json:"sequence"`
	BlockHeight int64  `json:"block_height"`
	From        string `json:"from"`
	To          string `json:"to"`
	L1Denom     string `json:"l1_denom"`
	L2Denom     string `json:"l2_denom"`
	Amount      string `json:"amount"`
	Data        []byte `json:"data"`
}
