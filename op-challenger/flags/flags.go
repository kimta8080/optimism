package flags

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-challenger/config"
	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	openum "github.com/ethereum-optimism/optimism/op-service/enum"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	oppprof "github.com/ethereum-optimism/optimism/op-service/pprof"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

const (
	envVarPrefix = "OP_CHALLENGER"
)

func prefixEnvVars(name string) []string {
	return opservice.PrefixEnvVar(envVarPrefix, name)
}

var (
	// Required Flags
	L1EthRpcFlag = &cli.StringFlag{
		Name:    "l1-eth-rpc",
		Usage:   "HTTP provider URL for L1.",
		EnvVars: prefixEnvVars("L1_ETH_RPC"),
	}
	FactoryAddressFlag = &cli.StringFlag{
		Name:    "game-factory-address",
		Usage:   "Address of the fault game factory contract.",
		EnvVars: prefixEnvVars("GAME_FACTORY_ADDRESS"),
	}
	GameAllowlistFlag = &cli.StringSliceFlag{
		Name: "game-allowlist",
		Usage: "List of Fault Game contract addresses the challenger is allowed to play. " +
			"If empty, the challenger will play all games.",
		EnvVars: prefixEnvVars("GAME_ALLOWLIST"),
	}
	TraceTypeFlag = &cli.GenericFlag{
		Name:    "trace-type",
		Usage:   "The trace type. Valid options: " + openum.EnumString(config.TraceTypes),
		EnvVars: prefixEnvVars("TRACE_TYPE"),
		Value: func() *config.TraceType {
			out := config.TraceType("") // No default value
			return &out
		}(),
	}
	AgreeWithProposedOutputFlag = &cli.BoolFlag{
		Name:    "agree-with-proposed-output",
		Usage:   "Temporary hardcoded flag if we agree or disagree with the proposed output.",
		EnvVars: prefixEnvVars("AGREE_WITH_PROPOSED_OUTPUT"),
	}
	DatadirFlag = &cli.StringFlag{
		Name:    "datadir",
		Usage:   "Directory to store data generated as part of responding to games",
		EnvVars: prefixEnvVars("DATADIR"),
	}
	// Optional Flags
	AlphabetFlag = &cli.StringFlag{
		Name:    "alphabet",
		Usage:   "Correct Alphabet Trace (alphabet trace type only)",
		EnvVars: prefixEnvVars("ALPHABET"),
	}
	CannonNetworkFlag = &cli.StringFlag{
		Name:    "cannon-network",
		Usage:   fmt.Sprintf("Predefined network selection. Available networks: %s (cannon trace type only)", strings.Join(chaincfg.AvailableNetworks(), ", ")),
		EnvVars: prefixEnvVars("CANNON_NETWORK"),
	}
	CannonRollupConfigFlag = &cli.StringFlag{
		Name:    "cannon-rollup-config",
		Usage:   "Rollup chain parameters (cannon trace type only)",
		EnvVars: prefixEnvVars("CANNON_ROLLUP_CONFIG"),
	}
	CannonL2GenesisFlag = &cli.StringFlag{
		Name:    "cannon-l2-genesis",
		Usage:   "Path to the op-geth genesis file (cannon trace type only)",
		EnvVars: prefixEnvVars("CANNON_L2_GENESIS"),
	}
	CannonBinFlag = &cli.StringFlag{
		Name:    "cannon-bin",
		Usage:   "Path to cannon executable to use when generating trace data (cannon trace type only)",
		EnvVars: prefixEnvVars("CANNON_BIN"),
	}
	CannonServerFlag = &cli.StringFlag{
		Name:    "cannon-server",
		Usage:   "Path to executable to use as pre-image oracle server when generating trace data (cannon trace type only)",
		EnvVars: prefixEnvVars("CANNON_SERVER"),
	}
	CannonPreStateFlag = &cli.StringFlag{
		Name:    "cannon-prestate",
		Usage:   "Path to absolute prestate to use when generating trace data (cannon trace type only)",
		EnvVars: prefixEnvVars("CANNON_PRESTATE"),
	}
	CannonL2Flag = &cli.StringFlag{
		Name:    "cannon-l2",
		Usage:   "L2 Address of L2 JSON-RPC endpoint to use (eth and debug namespace required)  (cannon trace type only)",
		EnvVars: prefixEnvVars("CANNON_L2"),
	}
	CannonSnapshotFreqFlag = &cli.UintFlag{
		Name:    "cannon-snapshot-freq",
		Usage:   "Frequency of cannon snapshots to generate in VM steps (cannon trace type only)",
		EnvVars: prefixEnvVars("CANNON_SNAPSHOT_FREQ"),
		Value:   config.DefaultCannonSnapshotFreq,
	}
	GameWindowFlag = &cli.DurationFlag{
		Name:    "game-window",
		Usage:   "The time window which the challenger will look for games to progress.",
		EnvVars: prefixEnvVars("GAME_WINDOW"),
		Value:   config.DefaultGameWindow,
	}
)

// requiredFlags are checked by [CheckRequired]
var requiredFlags = []cli.Flag{
	L1EthRpcFlag,
	FactoryAddressFlag,
	TraceTypeFlag,
	AgreeWithProposedOutputFlag,
	DatadirFlag,
}

// optionalFlags is a list of unchecked cli flags
var optionalFlags = []cli.Flag{
	AlphabetFlag,
	GameAllowlistFlag,
	CannonNetworkFlag,
	CannonRollupConfigFlag,
	CannonL2GenesisFlag,
	CannonBinFlag,
	CannonServerFlag,
	CannonPreStateFlag,
	CannonL2Flag,
	CannonSnapshotFreqFlag,
	GameWindowFlag,
}

func init() {
	optionalFlags = append(optionalFlags, oplog.CLIFlags(envVarPrefix)...)
	optionalFlags = append(optionalFlags, txmgr.CLIFlags(envVarPrefix)...)
	optionalFlags = append(optionalFlags, opmetrics.CLIFlags(envVarPrefix)...)
	optionalFlags = append(optionalFlags, oppprof.CLIFlags(envVarPrefix)...)

	Flags = append(requiredFlags, optionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.Names()[0]) {
			return fmt.Errorf("flag %s is required", f.Names()[0])
		}
	}
	gameType := config.TraceType(strings.ToLower(ctx.String(TraceTypeFlag.Name)))
	switch gameType {
	case config.TraceTypeCannon:
		if !ctx.IsSet(CannonNetworkFlag.Name) &&
			!(ctx.IsSet(CannonRollupConfigFlag.Name) && ctx.IsSet(CannonL2GenesisFlag.Name)) {
			return fmt.Errorf("flag %v or %v and %v is required",
				CannonNetworkFlag.Name, CannonRollupConfigFlag.Name, CannonL2GenesisFlag.Name)
		}
		if ctx.IsSet(CannonNetworkFlag.Name) &&
			(ctx.IsSet(CannonRollupConfigFlag.Name) || ctx.IsSet(CannonL2GenesisFlag.Name)) {
			return fmt.Errorf("flag %v can not be used with %v and %v",
				CannonNetworkFlag.Name, CannonRollupConfigFlag.Name, CannonL2GenesisFlag.Name)
		}
		if !ctx.IsSet(CannonBinFlag.Name) {
			return fmt.Errorf("flag %s is required", CannonBinFlag.Name)
		}
		if !ctx.IsSet(CannonServerFlag.Name) {
			return fmt.Errorf("flag %s is required", CannonServerFlag.Name)
		}
		if !ctx.IsSet(CannonPreStateFlag.Name) {
			return fmt.Errorf("flag %s is required", CannonPreStateFlag.Name)
		}
		if !ctx.IsSet(CannonL2Flag.Name) {
			return fmt.Errorf("flag %s is required", CannonL2Flag.Name)
		}
	case config.TraceTypeAlphabet:
		if !ctx.IsSet(AlphabetFlag.Name) {
			return fmt.Errorf("flag %s is required", "alphabet")
		}
	default:
		return fmt.Errorf("invalid trace type. must be one of %v", config.TraceTypes)
	}
	return nil
}

// NewConfigFromCLI parses the Config from the provided flags or environment variables.
func NewConfigFromCLI(ctx *cli.Context) (*config.Config, error) {
	if err := CheckRequired(ctx); err != nil {
		return nil, err
	}
	gameFactoryAddress, err := opservice.ParseAddress(ctx.String(FactoryAddressFlag.Name))
	if err != nil {
		return nil, err
	}
	var allowedGames []common.Address
	if ctx.StringSlice(GameAllowlistFlag.Name) != nil {
		for _, addr := range ctx.StringSlice(GameAllowlistFlag.Name) {
			gameAddress, err := opservice.ParseAddress(addr)
			if err != nil {
				return nil, err
			}
			allowedGames = append(allowedGames, gameAddress)
		}
	}

	txMgrConfig := txmgr.ReadCLIConfig(ctx)
	metricsConfig := opmetrics.ReadCLIConfig(ctx)
	pprofConfig := oppprof.ReadCLIConfig(ctx)

	traceTypeFlag := config.TraceType(strings.ToLower(ctx.String(TraceTypeFlag.Name)))

	return &config.Config{
		// Required Flags
		L1EthRpc:                ctx.String(L1EthRpcFlag.Name),
		TraceType:               traceTypeFlag,
		GameFactoryAddress:      gameFactoryAddress,
		GameAllowlist:           allowedGames,
		GameWindow:              ctx.Duration(GameWindowFlag.Name),
		AlphabetTrace:           ctx.String(AlphabetFlag.Name),
		CannonNetwork:           ctx.String(CannonNetworkFlag.Name),
		CannonRollupConfigPath:  ctx.String(CannonRollupConfigFlag.Name),
		CannonL2GenesisPath:     ctx.String(CannonL2GenesisFlag.Name),
		CannonBin:               ctx.String(CannonBinFlag.Name),
		CannonServer:            ctx.String(CannonServerFlag.Name),
		CannonAbsolutePreState:  ctx.String(CannonPreStateFlag.Name),
		Datadir:                 ctx.String(DatadirFlag.Name),
		CannonL2:                ctx.String(CannonL2Flag.Name),
		CannonSnapshotFreq:      ctx.Uint(CannonSnapshotFreqFlag.Name),
		AgreeWithProposedOutput: ctx.Bool(AgreeWithProposedOutputFlag.Name),
		TxMgrConfig:             txMgrConfig,
		MetricsConfig:           metricsConfig,
		PprofConfig:             pprofConfig,
	}, nil
}
