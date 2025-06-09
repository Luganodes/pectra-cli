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
	color.New(color.FgWhite).Println("  pectra-cli [command] [options]")

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
	color.New(color.FgGreen).Print("  broadcast     ")
	color.White("Broadcast a signed transaction")

	// Global options
	color.New(color.FgHiWhite, color.Bold).Println("\n🛠️  OPTIONS:")
	color.New(color.FgYellow).Print("  -c, --config    ")
	color.White("Path to configuration file (required for most commands)")
	color.New(color.FgYellow).Print("  -f, --file      ")
	color.White("Path to transaction file (required for broadcast command)")
	color.New(color.FgYellow).Print("  -a, --airgapped ")
	color.White("Run in airgapped mode")
	color.New(color.FgYellow).Print("  -h, --help      ")
	color.White("Show help for command")

	// Examples
	color.New(color.FgHiWhite, color.Bold).Println("\n📝 EXAMPLES:")
	color.White("  pectra-cli switch --config config.json")
	color.White("  pectra-cli consolidate -c config.json -a")
	color.White("  pectra-cli el-exit --config config.json")
	color.White("  pectra-cli broadcast --file signed_txn.json")

	// Configuration details
	color.New(color.FgHiWhite, color.Bold).Println("\n⚙️  CONFIGURATION FORMAT:")
	color.White("  The config file should be a JSON file with the following structure: (See `sample_config.json` for reference)")

	// Switch config
	color.New(color.FgYellow).Println("\n  Switch Operation:")
	color.White(`  {
    "switch": {
      "validators": ["Validator1", "Validator2", ...]
    }
  }`)

	// Consolidate config
	color.New(color.FgYellow).Println("\n  Consolidate Operation:")
	color.White(`  {
    "consolidate": {
      "sourceValidators": ["SourceValidator1", "SourceValidator2", ...],
      "targetValidator": "TargetValidator"
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
      }
    }
  }`)

	// Transaction file format
	color.New(color.FgYellow).Println("\n  Transaction File Format (for broadcast command):")
	color.White(`  {
    "signedTransaction": "0x...hex encoded signed transaction data..."
  }`)

	// Notes
	color.New(color.FgHiWhite, color.Bold).Println("\n📌 NOTES:")
	color.White("  • Private keys can be entered securely at runtime")
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

// RemoveDuplicateValidators takes a slice of validator strings and returns a new slice
// containing only unique validators.
func RemoveDuplicateValidators(validators []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, validator := range validators {
		if _, ok := seen[validator]; !ok {
			seen[validator] = true
			result = append(result, validator)
		}
	}
	return result
}
