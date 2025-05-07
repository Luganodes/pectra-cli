package utils

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
)

// PrintUsage prints the usage information
func PrintUsage() {
	// Title
	color.New(color.FgHiCyan, color.Bold).Println("\n╔════════════════════════════════════════════════════════╗")
	color.New(color.FgHiCyan, color.Bold).Println("║                  PECTRA CLI TOOL                       ║")
	color.New(color.FgHiCyan, color.Bold).Println("╚════════════════════════════════════════════════════════╝")

	// Basic usage
	color.New(color.FgHiWhite, color.Bold).Println("\n📋 USAGE:")
	color.New(color.FgWhite).Println("  pectra-cli <command> <config-file>")

	// Commands
	color.New(color.FgHiWhite, color.Bold).Println("\n🔧 COMMANDS:")
	color.New(color.FgGreen).Print("  switch        ")
	color.White("Execute batch switch operation for validators")
	color.New(color.FgGreen).Print("  consolidate   ")
	color.White("Consolidate multiple validators into a target validator")
	color.New(color.FgGreen).Print("  el-exit       ")
	color.White("Execute partial or full exits for validators")
	color.New(color.FgGreen).Print("  unset-code    ")
	color.White("Unset code for the contract")

	// Examples
	color.New(color.FgHiWhite, color.Bold).Println("\n📝 EXAMPLES:")
	color.White("  pectra-cli switch config.json")
	color.White("  pectra-cli consolidate config.json")
	color.White("  pectra-cli el-exit config.json")

	// Configuration details
	color.New(color.FgHiWhite, color.Bold).Println("\n⚙️  CONFIGURATION FORMAT:")
	color.White("  The config file should be a JSON file with the following structure:")

	// Switch config
	color.New(color.FgYellow).Println("\n  Switch Operation:")
	color.White(`  {
    "switch": {
      "validators": ["Validator1", "Validator2", ...],
      "amountPerValidator": 1
    }
  }`)

	// Consolidate config
	color.New(color.FgYellow).Println("\n  Consolidate Operation:")
	color.White(`  {
    "consolidate": {
      "sourceValidators": ["SourceValidator1", "SourceValidator2", ...],
      "targetValidator": "TargetValidator",
      "amountPerValidator": 1
    }
  }`)

	// EL exit config
	color.New(color.FgYellow).Println("\n  EL Exit Operation:")
	color.White(`  {
    "elExit": {
      "validators": {
        "Validator1": {
          "amount": 1000000000,  // Amount in Gwei (1 ETH = 1,000,000,000 Gwei)
          "confirmFullExit": false
        },
        "Validator2": {
          "amount": 0,
          "confirmFullExit": true  // Must be true for full exits
        }
      },
      "amountPerValidator": 1
    }
  }`)

	// Notes
	color.New(color.FgHiWhite, color.Bold).Println("\n📌 NOTES:")
	color.White("  • Private keys can be provided in the config file or entered securely at runtime")
	color.White("  • ALl validator addresses must be in hex format, without 0x prefix")
	color.White("  • To execute a full exit the amount should be 0 & confirmFullExit must be set to true")
	color.White("  • All amounts are specified in Gwei (1 ETH = 1,000,000,000 Gwei)")

	// Footer
	color.New(color.FgHiCyan).Println("\n═════════════════════════════════════════════════════════════════════")
	color.New(color.FgHiCyan).Println("For more information, visit: https://github.com/Luganodes/pectra-cli")
	color.New(color.FgHiCyan).Println("═════════════════════════════════════════════════════════════════════")
}

// GetFee calls the getFee function on the contract and returns the fee value
func GetFee(client *ethclient.Client, contractAddress common.Address, parsedABI abi.ABI, functionName string) (*big.Int, error) {
	// Pack the function call data
	data, err := parsedABI.Pack(functionName)
	if err != nil {
		return nil, fmt.Errorf("failed to pack data for getFee call: %v", err)
	}

	// Create a message call
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	// Execute the call
	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("contract call failed: %v", err)
	}

	// Unpack the result
	var fee *big.Int
	err = parsedABI.UnpackIntoInterface(&fee, functionName, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	return fee, nil
}
