package host

import (
	"encoding/json"

	executortypes "github.com/initia-labs/opinit-bots/executor/types"
	"github.com/initia-labs/opinit-bots/types"
)

func (h *Host) depositsToRawKV(deposits []executortypes.Deposit) ([]types.RawKV, error) {
	kvs := make([]types.RawKV, 0, len(deposits))
	for _, deposit := range deposits {
		data, err := json.Marshal(deposit)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, types.RawKV{
			Key:   h.DB().PrefixedKey(executortypes.PrefixedDepositKey(deposit.Sequence)),
			Value: data,
		})
	}
	return kvs, nil
}
