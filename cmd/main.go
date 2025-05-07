package main

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/mannan-goyal/0x04/internal/config"
	"github.com/mannan-goyal/0x04/internal/operations"
	"github.com/mannan-goyal/0x04/internal/transaction"
	"github.com/mannan-goyal/0x04/internal/utils"
)

func main() {
	if len(os.Args) < 3 {
		utils.PrintUsage()
		return
	}

	command := os.Args[1]
	configPath := os.Args[2]

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		color.Red("Error loading config: %v", err)
		return
	}

	// Connect to Ethereum client
	client, err := ethclient.Dial(cfg.RPCUrl)
	if err != nil {
		color.Red("Failed to connect to the Ethereum client: %v", err)
		return
	}
	color.Green("Connected to the Ethereum client")

	// Get private key securely
	privateKey, err := config.GetPrivateKey(cfg)
	if err != nil {
		color.Red("Failed to get the private key: %v", err)
		return
	}

	contractAddress := common.HexToAddress(cfg.PectraBatchContract)

	// Load the ABI
	parsedAbi, err := config.LoadABI("./abi.json")
	if err != nil {
		color.Red("%v", err)
		return
	}

	// Create base operation
	baseOp := operations.BaseOperation{
		Client:          client,
		PrivateKey:      privateKey,
		ContractAddress: contractAddress,
		ABI:             parsedAbi,
		ExplorerUrl:     cfg.BlockExplorerUrl,
	}

	var op operations.Operation

	// Helper function to get fee for a contract
	getFeeForContract := func(functionName string) (int64, error) {
		fee, err := utils.GetFee(client, contractAddress, parsedAbi, functionName)
		if err != nil {
			return 0, err
		}
		color.Green("Fee Amount per Validator: %v", fee)
		return fee.Int64(), nil
	}

	switch command {
	case "switch":
		feeAmount, err := getFeeForContract("getConsolidationFee")
		if err != nil {
			color.Red("Failed to get the fee: %v", err)
			return
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
			return
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
			return
		}

		op = &operations.ELExitOperation{
			BaseOperation:      baseOp,
			Validators:         cfg.ELExit.Validators,
			AmountPerValidator: feeAmount,
		}

	case "unset-code":
		err = transaction.SendTransactionUsingAuthorization(client, privateKey, common.Address{}, nil, nil, baseOp.ExplorerUrl)
		if err != nil {
			color.Red("Failed to execute unset-code: %v", err)
		}
		return

	default:
		color.Red("Unknown command: %s", command)
		utils.PrintUsage()
		return
	}

	// Execute the operation
	if err := op.Execute(); err != nil {
		color.Red("Operation failed: %v", err)
		return
	}
}
