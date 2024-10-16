package monitor

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	"go.uber.org/zap"

	"github.com/initia-labs/opinit-bots/keys"
	"github.com/initia-labs/opinit-bots/node"
	nodetypes "github.com/initia-labs/opinit-bots/node/types"
	"github.com/initia-labs/opinit-bots/types"
)

type Monitor struct {
	version uint8

	node *node.Node

	bridgeInfo opchildtypes.BridgeInfo

	cfg    nodetypes.NodeConfig
	logger *zap.Logger

	wasmQueryClient wasmtypes.QueryClient

	contractAddr string
	operatorID   uint32

	isOurTurn atomic.Bool
}

func NewMonitorV1(
	cfg nodetypes.NodeConfig,
	db types.DB,
	logger *zap.Logger,
	bech32Prefix string,
	contractAddr string,
	operatorID uint32,
) *Monitor {
	appCodec, txConfig, err := GetCodec(bech32Prefix)
	if err != nil {
		panic(err)
	}

	node, err := node.NewNode(cfg, db, logger, appCodec, txConfig)
	if err != nil {
		panic(err)
	}

	m := &Monitor{
		version: 1,

		node: node,

		cfg:    cfg,
		logger: logger,

		wasmQueryClient: wasmtypes.NewQueryClient(node.GetRPCClient()),
		contractAddr:    contractAddr,
		operatorID:      operatorID,
	}

	return m
}

func GetCodec(bech32Prefix string) (codec.Codec, client.TxConfig, error) {
	unlock := keys.SetSDKConfigContext(bech32Prefix)
	defer unlock()

	return keys.CreateCodec([]keys.RegisterInterfaces{
		auth.AppModuleBasic{}.RegisterInterfaces,
		authz.AppModuleBasic{}.RegisterInterfaces,
		wasm.AppModuleBasic{}.RegisterInterfaces,
	})
}

func (m *Monitor) Initialize(ctx context.Context, processedHeight int64, bridgeInfo opchildtypes.BridgeInfo) error {
	err := m.node.Initialize(ctx, processedHeight)
	if err != nil {
		return fmt.Errorf("initialize node: %w", err)
	}
	m.SetBridgeInfo(bridgeInfo)
	// TODO: check if the operator's admin is the `OperatorAdmin` in the config
	return nil
}

func (m *Monitor) Start(ctx context.Context) {
	m.logger.Info("monitor start", zap.Int64("height", m.node.GetHeight()))
	m.node.Start(ctx)

	eg := types.ErrGrp(ctx)
	eg.Go(func() (err error) {
		defer func() {
			m.Logger().Info("monitor looper stopped")
			if r := recover(); r != nil {
				m.Logger().Error("monitor looper panic", zap.Any("recover", r))
				err = fmt.Errorf("monitor looper panic: %v", r)
			}
		}()
		return m.runLoop(ctx)
	})
}

func (m *Monitor) SetBridgeInfo(bridgeInfo opchildtypes.BridgeInfo) {
	m.bridgeInfo = bridgeInfo
}

func (m Monitor) Node() *node.Node {
	return m.node
}

func (m Monitor) Logger() *zap.Logger {
	return m.logger
}

func (m Monitor) GetAddressStr() (string, error) {
	broadcaster, err := m.node.GetBroadcaster()
	if err != nil {
		return "", err
	}
	return broadcaster.GetAddressString()
}

func (m Monitor) IsOurTurn() bool {
	return m.isOurTurn.Load()
}
