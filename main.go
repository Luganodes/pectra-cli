package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/holiman/uint256"
)

// Config represents the JSON input file structure
type Config struct {
	WithdrawalAddressPrivateKey string            `json:"withdrawalAddressPrivateKey"`
	RPCUrl                      string            `json:"rpcUrl"`
	PectraBatchContract         string            `json:"pectraBatchContract"`
	Switch                      SwitchConfig      `json:"switch"`
	Consolidate                 ConsolidateConfig `json:"consolidate"`
	PartialExit                 PartialExitConfig `json:"partialExit"`
}

// SwitchConfig represents the switch configuration
type SwitchConfig struct {
	Validators         []string `json:"validators"`
	AmountPerValidator int64    `json:"amountPerValidator"`
}

// ConsolidateConfig represents the consolidate configuration
type ConsolidateConfig struct {
	SourceValidators   []string `json:"sourceValidators"`
	TargetValidator    string   `json:"targetValidator"`
	AmountPerValidator int64    `json:"amountPerValidator"`
}

// PartialExitConfig represents the partial exit configuration
type PartialExitConfig struct {
	Validators         map[string]float64 `json:"validators"`
	AmountPerValidator int64              `json:"amountPerValidator"`
}

// PartialExitData represents a validator and its exit amount
type PartialExitData struct {
	Pubkey string
	Amount *big.Int
}

func main() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	command := os.Args[1]
	configPath := os.Args[2]

	config, err := loadConfig(configPath)
	if err != nil {
		color.Red("Error loading config: %v\n", err)
		return
	}

	client, err := ethclient.Dial(config.RPCUrl)
	if err != nil {
		color.Red("Failed to connect to the Ethereum client: %v\n", err)
		return
	}
	color.Green("Connected to the Ethereum client")

	privateKey, err := crypto.HexToECDSA(config.WithdrawalAddressPrivateKey)
	if err != nil {
		color.Red("Failed to get the private key: %v\n", err)
		return
	}

	contractAddress := common.HexToAddress(config.PectraBatchContract)

	// Load the ABI only once
	abiPath, err := os.ReadFile("./abi.json")
	if err != nil {
		color.Red("Failed to read the ABI: %v\n", err)
		return
	}

	parsedAbi, err := abi.JSON(strings.NewReader(string(abiPath)))
	if err != nil {
		color.Red("Failed to parse the ABI: %v\n", err)
		return
	}

	switch command {
	case "switch":
		if len(config.Switch.Validators) == 0 {
			color.Red("No validators specified for switch operation")
			return
		}

		// Use provided amount or default to 1
		amountPerValidator := config.Switch.AmountPerValidator
		if amountPerValidator <= 0 {
			color.Yellow("Amount per validator is not set, using default value of 1")
			amountPerValidator = 1
		}

		batchSwitch(client, privateKey, contractAddress, config.Switch.Validators, amountPerValidator, parsedAbi)

	case "consolidate":
		if len(config.Consolidate.SourceValidators) == 0 || config.Consolidate.TargetValidator == "" {
			color.Red("Source or target validators not specified for consolidate operation")
			return
		}

		// Use provided amount or default to 1
		amountPerValidator := config.Consolidate.AmountPerValidator
		if amountPerValidator <= 0 {
			color.Yellow("Amount per validator is not set, using default value of 1")
			amountPerValidator = 1
		}

		// Create a combined validators array with target validator as the last element
		validators := append(
			config.Consolidate.SourceValidators,
			config.Consolidate.TargetValidator,
		)
		batchConsolidate(client, privateKey, contractAddress, validators, amountPerValidator, parsedAbi)

	case "partial-exit":
		if len(config.PartialExit.Validators) == 0 {
			color.Red("No validators specified for partial exit operation")
			return
		}

		// Use provided amount or default to 11
		amountPerValidator := config.PartialExit.AmountPerValidator
		if amountPerValidator <= 0 {
			color.Yellow("Amount per validator is not set, using default value of 1")
			amountPerValidator = 1
		}

		partialExits := []PartialExitData{}
		for pubkey, amountFloat := range config.PartialExit.Validators {
			// Convert to Gwei (divide by 10^9)
			amountGwei := new(big.Int).Div(
				new(big.Int).SetUint64(uint64(amountFloat)),
				new(big.Int).SetUint64(1000000000),
			)
			partialExits = append(partialExits, PartialExitData{
				Pubkey: pubkey,
				Amount: amountGwei,
			})
		}
		batchPartialExit(client, contractAddress, partialExits, privateKey, amountPerValidator, parsedAbi)

	case "unset-code":
		sendTransactionUsingAuthorization(client, privateKey, common.Address{}, nil, nil)

	default:
		color.Red("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	color.White("Usage: pectra-cli [switch|consolidate|partial-exit|unset-code] input.json")
	color.White("Example: pectra-cli switch input.json")
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func batchSwitch(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contract common.Address, validators []string, amountPerValidator int64, parsedAbi abi.ABI) {
	pubkeys := [][]byte{}
	for _, validator := range validators {
		pubkeys = append(pubkeys, common.FromHex(validator))
	}

	data, err := parsedAbi.Pack("batchSwitch", pubkeys)
	if err != nil {
		color.Red("Failed to pack the data: %v\n", err)
		return
	}

	value := new(big.Int)
	value.Mul(big.NewInt(int64(len(validators))), big.NewInt(amountPerValidator))
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)\n",
		value, len(validators), amountPerValidator)

	sendTransactionUsingAuthorization(client, privateKey, contract, data, uint256.NewInt(uint64(value.Int64())))
}

func batchConsolidate(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contract common.Address, validators []string, amountPerValidator int64, parsedAbi abi.ABI) {
	pubkeys := [][]byte{}
	for _, validator := range validators[:len(validators)-1] {
		pubkeys = append(pubkeys, common.FromHex(validator))
	}
	target := common.FromHex(validators[len(validators)-1])

	data, err := parsedAbi.Pack("batchConsolidation", pubkeys, target)
	if err != nil {
		color.Red("Failed to pack the data: %v\n", err)
		return
	}

	value := new(big.Int)
	value.Mul(big.NewInt(int64(len(pubkeys))), big.NewInt(amountPerValidator))
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)\n",
		value, len(pubkeys), amountPerValidator)

	sendTransactionUsingAuthorization(client, privateKey, contract, data, uint256.NewInt(uint64(value.Int64())))
}

func batchPartialExit(client *ethclient.Client, contract common.Address, validators []PartialExitData, privateKey *ecdsa.PrivateKey, amountPerValidator int64, parsedAbi abi.ABI) {
	pubkeys := [][][]byte{}
	for _, validator := range validators {
		hexAmount := validator.Amount.Text(16)
		paddedAmount := strings.Repeat("0", 16-len(hexAmount)) + hexAmount
		pubkeys = append(pubkeys, [][]byte{common.FromHex(validator.Pubkey), common.FromHex(paddedAmount)})
	}

	data, err := parsedAbi.Pack("batchELExit", pubkeys)
	if err != nil {
		color.Red("Failed to pack the data: %v\n", err)
		return
	}

	value := new(big.Int)
	value.Mul(big.NewInt(int64(len(validators))), big.NewInt(amountPerValidator))
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)\n",
		value, len(validators), amountPerValidator)

	sendTransactionUsingAuthorization(client, privateKey, contract, data, uint256.NewInt(uint64(value.Int64())))
	// sendTransactionUsingAuthorization(client, privateKey, contract, data, uint256.NewInt(9))

}

func sendTransactionUsingAuthorization(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contract common.Address, data []byte, value *uint256.Int) {
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		color.Red("Failed to get the chain ID: %v\n", err)
		return
	}
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		color.Red("Failed to get nonce: %v", err)
		return
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		color.Red("Failed to get gas price: %v", err)
		return
	}
	tipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		color.Red("Failed to get the gas tip cap: %v\n", err)
		return
	}

	authorization := types.SetCodeAuthorization{
		ChainID: *uint256.NewInt(chainID.Uint64()),
		Address: contract,
		Nonce:   nonce + 1,
	}

	signedAuthorization, err := types.SignSetCode(privateKey, authorization)
	if err != nil {
		color.Red("Failed to sign the authorization: %v\n", err)
		return
	}

	tx := types.NewTx(&types.SetCodeTx{
		ChainID:   uint256.NewInt(chainID.Uint64()),
		Nonce:     nonce,
		GasTipCap: uint256.NewInt(tipCap.Uint64()),
		GasFeeCap: uint256.NewInt(gasPrice.Uint64()),
		Gas:       uint64(30000000),
		To:        fromAddress,
		Value:     value,
		Data:      data,
		AuthList:  []types.SetCodeAuthorization{signedAuthorization},
	})

	tx, err = types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		color.Red("Failed to sign the transaction: %v\n", err)
		return
	}

	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		color.Red("Failed to send the transaction: %v\n", err)
		return
	}

	color.Cyan("Transaction sent: https://hoodi.etherscan.io/tx/%s\n", tx.Hash().Hex())
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		color.Red("Failed to wait for the transaction to be mined: %v\n", err)
		return
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		color.Red("Transaction failed")
		return
	}
	color.Green("Transaction successful")
}
