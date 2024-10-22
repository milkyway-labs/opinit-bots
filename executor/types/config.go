package types

import (
	"errors"
	"time"

	btypes "github.com/initia-labs/opinit-bots/node/broadcaster/types"
	nodetypes "github.com/initia-labs/opinit-bots/node/types"
)

type NodeConfig struct {
	ChainID       string  `json:"chain_id"`
	Bech32Prefix  string  `json:"bech32_prefix"`
	RPCAddress    string  `json:"rpc_address"`
	GasPrice      string  `json:"gas_price"`
	GasAdjustment float64 `json:"gas_adjustment"`
	TxTimeout     int64   `json:"tx_timeout"` // seconds
}

func (nc NodeConfig) Validate() error {
	if nc.ChainID == "" {
		return errors.New("chain ID is required")
	}
	if nc.Bech32Prefix == "" {
		return errors.New("bech32 prefix is required")
	}
	if nc.RPCAddress == "" {
		return errors.New("RPC address is required")
	}
	return nil
}

type Config struct {
	// Version is the version used to build output root.
	Version uint8 `json:"version"`

	// ListenAddress is the address to listen for incoming requests.
	ListenAddress string `json:"listen_address"`

	// L1Node is the configuration for the l1 node.
	L1Node NodeConfig `json:"l1_node"`
	// L2Node is the configuration for the l2 node.
	L2Node NodeConfig `json:"l2_node"`
	// DANode is the configuration for the data availability node.
	DANode NodeConfig `json:"da_node"`
	// MilkyWayNode is the configuration for the MilkyWay node.
	MilkyWayNode NodeConfig `json:"milkyway_node"`

	// OperatorID is the MilkyWay restaking operator ID of this bot.
	OperatorID uint32 `json:"operator_id"`
	// OperatorAdmin is the key name in the keyring for operator admin on MilkyWay.
	OperatorAdmin string `json:"operator_admin"`

	// OutputSubmitter is the key name in the keyring for the output submitter,
	// which is used to relay the output transaction from l2 to l1.
	// The L2's output submitter must have granted the output submission permission
	// to this account.
	//
	// If you don't want to use the output submitter feature, you can leave it empty.
	OutputSubmitter string `json:"output_submitter"`

	// BatchSubmitter is the key name in the keyring for the batch submitter,
	// which is used to batch submission from l2 to a DA layer.
	//
	// If you don't want to use the batch submitter feature, you can leave it empty.
	BatchSubmitter string `json:"batch_submitter"`

	// BridgeExecutor is the key name in the keyring for the bridge executor,
	// which is used to relay initiate token bridge transaction from l1 to l2.
	// The L2's bridge executor must have granted the correct permissions
	// to this account.
	//
	// If you don't want to use the bridge executor feature, you can leave it empty.
	BridgeExecutor string `json:"bridge_executor"`

	// L2BridgeExecutor is the address of the L2 bridge executor,
	// which is used to relay initiate token bridge transaction from l1 to l2.
	L2BridgeExecutor string `json:"l2_bridge_executor"`

	// MaxChunks is the maximum number of chunks in a batch.
	MaxChunks int64 `json:"max_chunks"`
	// MaxChunkSize is the maximum size of a chunk in a batch.
	MaxChunkSize int64 `json:"max_chunk_size"`
	// MaxSubmissionTime is the maximum time to submit a batch.
	MaxSubmissionTime int64 `json:"max_submission_time"` // seconds

	// DisableAutoSetL1Height is the flag to disable the automatic setting of the l1 height.
	// If it is false, it will finds the optimal height and sets l1_start_height automatically
	// from l2 start height and l1_start_height is ignored.
	// It can be useful when you don't want to use TxSearch.
	DisableAutoSetL1Height bool `json:"disable_auto_set_l1_height"`
	// L1StartHeight is the height to start the l1 node.
	L1StartHeight int64 `json:"l1_start_height"`
	// L2StartHeight is the height to start the l2 node. If it is 0, it will start from the latest height.
	// If the latest height stored in the db is not 0, this config is ignored.
	// L2 starts from the last submitted output l2 block number + 1 before L2StartHeight.
	// L1 starts from the block number of the output tx + 1
	L2StartHeight int64 `json:"l2_start_height"`
	// BatchStartHeight is the height to start the batch. If it is 0, it will start from the latest height.
	// If the latest height stored in the db is not 0, this config is ignored.
	BatchStartHeight int64 `json:"batch_start_height"`
	// MilkyWayStartHeight is the height to start the MilkyWay node. If it is 0, it
	// will start from the latest height.
	MilkyWayStartHeight int64 `json:"milkyway_start_height"`

	// MonitorContract is the contract address to monitor.
	MonitorContract string `json:"monitor_contract"`
}

func DefaultConfig() *Config {
	return &Config{
		Version:       1,
		ListenAddress: "localhost:3000",

		L1Node: NodeConfig{
			ChainID:       "testnet-l1-1",
			Bech32Prefix:  "init",
			RPCAddress:    "tcp://localhost:26657",
			GasPrice:      "0.15uinit",
			GasAdjustment: 1.5,
			TxTimeout:     60,
		},

		L2Node: NodeConfig{
			ChainID:       "testnet-l2-1",
			Bech32Prefix:  "init",
			RPCAddress:    "tcp://localhost:27657",
			GasPrice:      "",
			GasAdjustment: 1.5,
			TxTimeout:     60,
		},

		DANode: NodeConfig{
			ChainID:       "testnet-l1-1",
			Bech32Prefix:  "init",
			RPCAddress:    "tcp://localhost:26657",
			GasPrice:      "0.15uinit",
			GasAdjustment: 1.5,
			TxTimeout:     60,
		},

		MilkyWayNode: NodeConfig{
			ChainID:       "testnet-milkyway-1",
			Bech32Prefix:  "init",
			RPCAddress:    "tcp://localhost:28657",
			GasPrice:      "",
			GasAdjustment: 1.5,
			TxTimeout:     60,
		},

		MaxChunks:         5000,
		MaxChunkSize:      300000,  // 300KB
		MaxSubmissionTime: 60 * 60, // 1 hour

		DisableAutoSetL1Height: false,
		L1StartHeight:          0,
		L2StartHeight:          0,
		BatchStartHeight:       0,
	}
}

func (cfg Config) Validate() error {
	if cfg.Version == 0 {
		return errors.New("version is required")
	}

	if cfg.Version != 1 {
		return errors.New("only version 1 is supported")
	}

	if cfg.ListenAddress == "" {
		return errors.New("listen address is required")
	}

	if err := cfg.L1Node.Validate(); err != nil {
		return err
	}

	if err := cfg.L2Node.Validate(); err != nil {
		return err
	}

	if err := cfg.DANode.Validate(); err != nil {
		return err
	}

	if err := cfg.MilkyWayNode.Validate(); err != nil {
		return err
	}

	if cfg.MaxChunks <= 0 {
		return errors.New("max chunks must be greater than 0")
	}

	if cfg.MaxChunkSize <= 0 {
		return errors.New("max chunk size must be greater than 0")
	}

	if cfg.MaxSubmissionTime <= 0 {
		return errors.New("max submission time must be greater than 0")
	}

	if cfg.L1StartHeight < 0 {
		return errors.New("l1 start height must be greater than or equal to 0")
	}

	if cfg.L2StartHeight < 0 {
		return errors.New("l2 start height must be greater than or equal to 0")
	}

	if cfg.BatchStartHeight < 0 {
		return errors.New("batch start height must be greater than or equal to 0")
	}

	if cfg.MilkyWayStartHeight < 0 {
		return errors.New("milkyway start height must be greater than or equal to 0")
	}

	if cfg.BridgeExecutor != "" && cfg.L2BridgeExecutor == "" {
		return errors.New("l2 bridge executor must be set when bridge executor is set")
	}

	if cfg.MonitorContract == "" {
		return errors.New("monitor contract must be set")
	}
	return nil
}

func (cfg Config) L1NodeConfig(homePath string) nodetypes.NodeConfig {
	nc := nodetypes.NodeConfig{
		RPC:         cfg.L1Node.RPCAddress,
		ProcessType: nodetypes.PROCESS_TYPE_DEFAULT,
	}

	if cfg.OutputSubmitter != "" {
		nc.BroadcasterConfig = &btypes.BroadcasterConfig{
			ChainID:       cfg.L1Node.ChainID,
			GasPrice:      cfg.L1Node.GasPrice,
			GasAdjustment: cfg.L1Node.GasAdjustment,
			Bech32Prefix:  cfg.L1Node.Bech32Prefix,
			TxTimeout:     time.Duration(cfg.L1Node.TxTimeout) * time.Second,
			KeyringConfig: btypes.KeyringConfig{
				Name:     cfg.OutputSubmitter,
				HomePath: homePath,
			},
		}
	}

	return nc
}

func (cfg Config) L2NodeConfig(homePath string) nodetypes.NodeConfig {
	nc := nodetypes.NodeConfig{
		RPC:         cfg.L2Node.RPCAddress,
		ProcessType: nodetypes.PROCESS_TYPE_DEFAULT,
	}

	if cfg.BridgeExecutor != "" {
		nc.BroadcasterConfig = &btypes.BroadcasterConfig{
			ChainID:       cfg.L2Node.ChainID,
			GasPrice:      cfg.L2Node.GasPrice,
			GasAdjustment: cfg.L2Node.GasAdjustment,
			Bech32Prefix:  cfg.L2Node.Bech32Prefix,
			TxTimeout:     time.Duration(cfg.L2Node.TxTimeout) * time.Second,
			KeyringConfig: btypes.KeyringConfig{
				Name:     cfg.BridgeExecutor,
				HomePath: homePath,
			},
		}
	}

	return nc
}

func (cfg Config) DANodeConfig(homePath string) nodetypes.NodeConfig {
	nc := nodetypes.NodeConfig{
		RPC:         cfg.DANode.RPCAddress,
		ProcessType: nodetypes.PROCESS_TYPE_ONLY_BROADCAST,
	}

	if cfg.BatchSubmitter != "" {
		nc.BroadcasterConfig = &btypes.BroadcasterConfig{
			ChainID:       cfg.DANode.ChainID,
			GasPrice:      cfg.DANode.GasPrice,
			GasAdjustment: cfg.DANode.GasAdjustment,
			Bech32Prefix:  cfg.DANode.Bech32Prefix,
			TxTimeout:     time.Duration(cfg.DANode.TxTimeout) * time.Second,
			KeyringConfig: btypes.KeyringConfig{
				Name:     cfg.BatchSubmitter,
				HomePath: homePath,
			},
		}
	}
	return nc
}

func (cfg Config) BatchConfig() BatchConfig {
	return BatchConfig{
		MaxChunks:         cfg.MaxChunks,
		MaxChunkSize:      cfg.MaxChunkSize,
		MaxSubmissionTime: cfg.MaxSubmissionTime,
	}
}

func (cfg Config) MilkyWayNodeConfig(homePath string) nodetypes.NodeConfig {
	nc := nodetypes.NodeConfig{
		RPC:         cfg.MilkyWayNode.RPCAddress,
		ProcessType: nodetypes.PROCESS_TYPE_ONLY_BROADCAST,
	}

	if cfg.OperatorAdmin != "" {
		nc.BroadcasterConfig = &btypes.BroadcasterConfig{
			ChainID:       cfg.MilkyWayNode.ChainID,
			GasPrice:      cfg.MilkyWayNode.GasPrice,
			GasAdjustment: cfg.MilkyWayNode.GasAdjustment,
			TxTimeout:     time.Duration(cfg.MilkyWayNode.TxTimeout) * time.Second,
			Bech32Prefix:  cfg.MilkyWayNode.Bech32Prefix,
			KeyringConfig: btypes.KeyringConfig{
				Name:     cfg.OperatorAdmin,
				HomePath: homePath,
			},
		}
	}
	return nc
}

type BatchConfig struct {
	MaxChunks         int64 `json:"max_chunks"`
	MaxChunkSize      int64 `json:"max_chunk_size"`
	MaxSubmissionTime int64 `json:"max_submission_time"` // seconds
}
