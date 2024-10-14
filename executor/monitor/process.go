package monitor

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/initia-labs/opinit-bots/types"
)

func (m *Monitor) runLoop(ctx context.Context) error {
	ticker := time.NewTicker(types.PollingInterval(ctx))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}

		err := m.handleEpoch(ctx)
		if err != nil {
			m.Logger().Error("failed to handle epoch", zap.String("error", err.Error()))
			continue
		}
	}
}

