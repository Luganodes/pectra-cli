package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"

	"github.com/Luganodes/Pectra-CLI/internal/config"
	"github.com/Luganodes/Pectra-CLI/internal/operations"
	"github.com/Luganodes/Pectra-CLI/internal/transaction"
	"github.com/Luganodes/Pectra-CLI/internal/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	version = "1.0.0"
)

func main() {
	app := &cli.App{
		Name:     "pectra-cli",
		Usage:    "CLI tool for Ethereum validator operations",
		Version:  version,
		HelpName: "pectra-cli",
		Authors: []*cli.Author{
			{
				Name:  "Luganodes",
				Email: "hello@luganodes.com",
			},
		},
		Description: "Pectra CLI is a tool for executing operations on Ethereum validators including switching, consolidation, and execution layer exits",
		Commands: []*cli.Command{
			{
				Name:        "switch",
				Usage:       "Execute batch switch operation for validators",
				Description: "Switch validators to a new setup based on configuration",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Path to config file (required)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "airgapped",
						Aliases: []string{"a"},
						Usage:   "Run in airgapped mode",
					},
				},
				Action: func(c *cli.Context) error {
					return runCommand("switch", c.String("config"), c.Bool("airgapped"))
				},
			},
			{
				Name:        "consolidate",
				Usage:       "Consolidate multiple validators into a target validator",
				Description: "Consolidate funds from multiple source validators into a single target validator",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Path to config file (required)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "airgapped",
						Aliases: []string{"a"},
						Usage:   "Run in airgapped mode",
					},
				},
				Action: func(c *cli.Context) error {
					return runCommand("consolidate", c.String("config"), c.Bool("airgapped"))
				},
			},
			{
				Name:        "el-exit",
				Usage:       "Execute partial or full exits for validators",
				Description: "Execute execution layer exits for validators, either partially or fully",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Path to config file (required)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "airgapped",
						Aliases: []string{"a"},
						Usage:   "Run in airgapped mode",
					},
				},
				Action: func(c *cli.Context) error {
					return runCommand("el-exit", c.String("config"), c.Bool("airgapped"))
				},
			},
			{
				Name:        "unset-code",
				Usage:       "Unset code for the contract",
				Description: "Remove the contract code (for emergency situations only)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Path to config file (required)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "airgapped",
						Aliases: []string{"a"},
						Usage:   "Run in airgapped mode",
					},
				},
				Action: func(c *cli.Context) error {
					return runCommand("unset-code", c.String("config"), c.Bool("airgapped"))
				},
			},
			{
				Name:        "broadcast",
				Usage:       "Broadcast a signed transaction",
				Description: "Broadcast a previously signed transaction from a JSON file containing a 'signedTransaction' field with the hex-encoded transaction data. Chain-specific explorer and RPC URLs are automatically determined from the transaction's chain ID.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "Path to JSON file containing the signed transaction (required)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Path to config file (required)",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return broadcastTransaction(c.String("file"), c.String("config"))
				},
			},
		},
		// Use the custom help template from utils.PrintUsage when showing app help
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			utils.PrintUsage()
			return err
		},
		CommandNotFound: func(c *cli.Context, command string) {
			color.Red("Unknown command: %s", command)
			utils.PrintUsage()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runCommand(command, configPath string, airgapped bool) error {
	color.Green("Airgapped: %v", airgapped)

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		color.Red("Error loading config: %v", err)
		return err
	}

	// Connect to Ethereum client
	client, err := ethclient.Dial(cfg.RPCUrl)
	if err != nil {
		color.Red("Failed to connect to the Ethereum client: %v", err)
		return err
	}
	color.Green("Connected to the Ethereum client")

	var privateKey *ecdsa.PrivateKey
	if !airgapped {
		// Get private key securely
		privateKey, err = config.GetPrivateKey()
		if err != nil {
			color.Red("Failed to get the private key: %v", err)
			return err
		}
	}

	contractAddress := common.HexToAddress(cfg.PectraBatchContract)

	// Load the ABI
	parsedAbi, err := config.LoadABI()
	if err != nil {
		color.Red("%v", err)
		return err
	}

	// Create base operation
	baseOp := operations.BaseOperation{
		Client:          client,
		ContractAddress: contractAddress,
		ABI:             parsedAbi,
		ExplorerUrl:     cfg.BlockExplorerUrl,
		Airgapped:       airgapped,
	}

	if !airgapped {
		baseOp.PrivateKey = privateKey
	}

	var op operations.Operation

	// Helper function to get fee for a contract
	getFeeForContract := func(functionName string) (int64, error) {
		fee, err := utils.GetFee(client, contractAddress, parsedAbi, functionName)
		if err != nil {
			return 0, err
		}
		color.Green("Fee Amount per Validator: %v wei", fee)
		return fee.Int64(), nil
	}

	switch command {
	case "switch":
		feeAmount, err := getFeeForContract("getConsolidationFee")
		if err != nil {
			color.Red("Failed to get the fee: %v", err)
			return err
		}

		op = &operations.SwitchOperation{
			BaseOperation:      baseOp,
			Validators:         cfg.Switch.Validators,
			AmountPerValidator: feeAmount,
		}

	case "consolidate":
		feeAmount, err := getFeeForContract("getConsolidationFee")
		if err != nil {
			color.Red("Failed to get the fee: %v", err)
			return err
		}

		op = &operations.ConsolidateOperation{
			BaseOperation:      baseOp,
			SourceValidators:   cfg.Consolidate.SourceValidators,
			TargetValidator:    cfg.Consolidate.TargetValidator,
			AmountPerValidator: feeAmount,
		}

	case "el-exit":
		feeAmount, err := getFeeForContract("getExitFee")
		if err != nil {
			color.Red("Failed to get the fee: %v", err)
			return err
		}

		op = &operations.ELExitOperation{
			BaseOperation:      baseOp,
			Validators:         cfg.ELExit.Validators,
			AmountPerValidator: feeAmount,
		}

	case "unset-code":
		if !airgapped && privateKey == nil {
			color.Red("Private key is required for unset-code operation in non-airgapped mode")
			return fmt.Errorf("private key required")
		}
		err = transaction.SendTransactionUsingAuthorization(client, privateKey, common.Address{}, nil, nil, baseOp.ExplorerUrl, airgapped)
		if err != nil {
			color.Red("Failed to execute unset-code: %v", err)
			return err
		}
		return nil

	default:
		color.Red("Unknown command: %s", command)
		return fmt.Errorf("unknown command: %s", command)
	}

	// Execute the operation
	if err := op.Execute(); err != nil {
		color.Red("Operation failed: %v", err)
		return err
	}

	return nil
}

// Add this new function for broadcasting transactions
func broadcastTransaction(txFilePath string, configPath string) error {
	color.Green("Broadcasting transaction from file: %s", txFilePath)

	// Call the broadcast function directly with the file
	err := transaction.BroadcastTransactionFromFile(txFilePath, configPath)
	if err != nil {
		color.Red("Failed to broadcast transaction: %v", err)
		return err
	}

	color.Green("Transaction broadcast process completed")
	return nil
}
