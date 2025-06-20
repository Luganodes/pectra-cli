package config

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fatih/color"
	"golang.org/x/term"
)

//go:embed abi.json
var abiFile []byte

// Config represents the JSON input file structure
type Config struct {
	RPCUrl              string            `json:"rpcUrl"`
	BlockExplorerUrl    string            `json:"blockExplorerUrl"`
	PectraBatchContract string            `json:"pectraBatchContract"`
	Switch              SwitchConfig      `json:"switch"`
	Consolidate         ConsolidateConfig `json:"consolidate"`
	ELExit              ELExitConfig      `json:"elExit"`
}

// SwitchConfig represents the switch configuration
type SwitchConfig struct {
	Validators []string `json:"validators"`
}

// ConsolidateConfig represents the consolidate configuration
type ConsolidateConfig struct {
	SourceValidators []string `json:"sourceValidators"`
	TargetValidator  string   `json:"targetValidator"`
}

// ELExitConfig represents the EL exit configuration
type ELExitConfig struct {
	Validators map[string]ELExitDetails `json:"validators"`
}

// elExitDetails represents a validator's exit details
type ELExitDetails struct {
	Amount          float64 `json:"amount"`
	ConfirmFullExit bool    `json:"confirmFullExit"` // New field to confirm full exit when amount is 0
}

// LoadConfig loads and validates the configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if config.RPCUrl == "" {
		return nil, fmt.Errorf("rpcUrl is required in the configuration")
	}

	if config.PectraBatchContract == "" {
		return nil, fmt.Errorf("pectraBatchContract is required in the configuration")
	}

	return &config, nil
}

// LoadABI loads the ABI from a file
func LoadABI() (abi.ABI, error) {
	contractABI, err := abi.JSON(strings.NewReader(string(abiFile)))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI: %v", err)
	}
	return contractABI, nil
}

// GetPrivateKey securely gets the private key from the config or prompts the user
func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	var privateKeyHex string

	// If private key is not in config, prompt for it securely
	color.Cyan("Please enter your private key (without 0x prefix):")
	color.Yellow("Note: For security, the key will not be displayed when pasted. Just paste and press Enter.")
	fmt.Print("> ")

	// Read password without echoing to terminal
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	fmt.Println() // Add a newline after the password input

	privateKeyHex = strings.TrimSpace(string(bytePassword))

	// Validate private key format
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %w", err)
	}

	return privateKey, nil
}

// GetPublicKey gets a withdrawal address (EOA) public key from user input
func GetPublicKey() (string, error) {
	var publicKeyHex string

	// Prompt user for input
	color.Cyan("Please enter the withdrawal address (0x... format):")
	fmt.Print("> ")

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return "", fmt.Errorf("failed to read address: %w", err)
	}

	publicKeyHex = strings.TrimSpace(input)

	// Check format
	if !strings.HasPrefix(publicKeyHex, "0x") {
		return "", fmt.Errorf("address must start with 0x prefix")
	}

	// Should be 42 characters (0x + 40 hex chars)
	if len(publicKeyHex) != 42 {
		return "", fmt.Errorf("invalid address length: expected 42 characters (20 bytes), got %d", len(publicKeyHex))
	}

	// Check if it's a valid hex string
	_, decodeErr := hex.DecodeString(publicKeyHex[2:])
	if decodeErr != nil {
		return "", fmt.Errorf("invalid address format: %w", decodeErr)
	}

	// Validate checksum if it's a mixed-case address
	if !common.IsHexAddress(publicKeyHex) {
		return "", fmt.Errorf("invalid Ethereum address")
	}

	// Return checksum address for consistency
	return common.HexToAddress(publicKeyHex).Hex(), nil
}
