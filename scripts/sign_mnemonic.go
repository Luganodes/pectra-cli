package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/fatih/color"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/term"
)

const (
	// Default derivation path for Ethereum accounts 
	defaultDerivationPath = "m/44'/60'/0'/0/0"
)

func main() {
	color.Green("Pectra CLI - Mnemonic-based Transaction Signer")
	color.Cyan("This tool signs transactions using a mnemonic phrase (seed phrase)")
	fmt.Println()

	// Get mnemonic from user
	mnemonic, err := getMnemonic()
	if err != nil {
		log.Fatalf("Failed to get mnemonic: %v", err)
	}

	// Get derivation path from user
	derivationPath, err := getDerivationPath()
	if err != nil {
		log.Fatalf("Failed to get derivation path: %v", err)
	}

	// Derive private key from mnemonic
	privateKey, address, err := derivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		log.Fatalf("Failed to derive private key from mnemonic: %v", err)
	}

	color.Green("Derived address: %s", address)
	fmt.Println()

	// Determine input filename
	inputFile := "unsigned_txn.json"
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
	}

	color.Cyan("Reading unsigned transaction from: %s", inputFile)

	// Read the unsigned transaction from file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", inputFile, err)
	}

	var txData struct {
		UnsignedTransaction string `json:"unsignedTransaction"`
		ChainId             string `json:"chainId"`
	}

	if err := json.Unmarshal(data, &txData); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	hexTx := txData.UnsignedTransaction

	// Remove 0x prefix if present
	if len(hexTx) > 2 && hexTx[0:2] == "0x" {
		hexTx = hexTx[2:]
	}

	// Decode hex to bytes
	txBytes, err := hex.DecodeString(hexTx)
	if err != nil {
		log.Fatalf("Failed to decode hex string: %v", err)
	}

	// Decode transaction
	tx := new(types.Transaction)
	err = rlp.DecodeBytes(txBytes, tx)
	if err != nil {
		log.Fatalf("Failed to decode transaction: %v", err)
	}

	color.Cyan("Transaction details:")
	color.White("  Hash: %s", tx.Hash().Hex())
	color.White("  Chain ID: %s", tx.ChainId().String())
	color.White("  To: %s", tx.To().Hex())
	color.White("  Value: %s Wei", tx.Value().String())
	color.White("  Gas: %d", tx.Gas())
	fmt.Println()

	// Confirm signing
	if !confirmSigning() {
		color.Yellow("Transaction signing cancelled by user")
		return
	}

	// Sign the authorization
	color.Cyan("Signing authorization...")
	signedAuthorization, err := types.SignSetCode(privateKey, tx.SetCodeAuthorizations()[0])
	if err != nil {
		log.Fatalf("Failed to sign the authorization: %v", err)
	}

	// Update the transaction with signed authorization
	tx.SetCodeAuthorizations()[0] = signedAuthorization

	// Sign the transaction
	color.Cyan("Signing transaction...")
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(tx.ChainId()), privateKey)
	if err != nil {
		log.Fatalf("Failed to sign the transaction: %v", err)
	}

	// Encode the signed transaction
	signedTxBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		log.Fatalf("Failed to encode signed transaction: %v", err)
	}

	// Write the signed transaction to a file
	signedData := map[string]string{
		"signedTransaction": hex.EncodeToString(signedTxBytes),
	}

	jsonData, err := json.MarshalIndent(signedData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal to JSON: %v", err)
	}

	outputFile := "signed_txn.json"
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		log.Fatalf("Failed to write to %s: %v", outputFile, err)
	}

	color.Green("Transaction successfully signed!")
	color.Green("Signed transaction written to: %s", outputFile)
	color.Cyan("You can now broadcast this transaction using: ./pectra-cli broadcast -f %s -c config.json", outputFile)
}

// getMnemonic securely prompts the user for their mnemonic phrase
func getMnemonic() (string, error) {
	color.Cyan("Please enter your mnemonic phrase (12, 15, 18, 21, or 24 words):")
	color.Yellow("Note: The mnemonic will not be displayed for security. Just type/paste and press Enter.")
	fmt.Print("> ")

	// Read mnemonic without echoing to terminal
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read mnemonic: %w", err)
	}
	fmt.Println() 

	mnemonic := strings.TrimSpace(string(bytePassword))
	
	// Basic validation
	words := strings.Fields(mnemonic)
	if len(words) < 12 || len(words) > 24 {
		return "", fmt.Errorf("invalid mnemonic length: expected 12-24 words, got %d", len(words))
	}

	// Validate mnemonic using BIP-39
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", fmt.Errorf("invalid mnemonic phrase")
	}

	return mnemonic, nil
}

// getDerivationPath prompts the user for a derivation path or uses the default
func getDerivationPath() (string, error) {
	color.Cyan("Enter derivation path (or press Enter for default: %s):", defaultDerivationPath)
	color.Yellow("Common paths:")
	color.Yellow("  Ethereum (default): m/44'/60'/0'/0/0")
	color.Yellow("  Ledger Live: m/44'/60'/0'/0/0")
	color.Yellow("  MEW/MyCrypto: m/44'/60'/0'/0")
	color.Yellow("  Custom account index: m/44'/60'/0'/0/[index]")
	fmt.Print("> ")

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil && err.Error() != "unexpected newline" {
		return "", fmt.Errorf("failed to read derivation path: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultDerivationPath, nil
	}

	// Basic validation for derivation path format
	if !strings.HasPrefix(input, "m/") {
		return "", fmt.Errorf("derivation path must start with 'm/'")
	}

	return input, nil
}

// derivePrivateKeyFromMnemonic derives a private key from a mnemonic using the specified derivation path
func derivePrivateKeyFromMnemonic(mnemonic, derivationPath string) (*ecdsa.PrivateKey, string, error) {
	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Create master key
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create master key: %w", err)
	}

	// Parse and apply derivation path
	path, err := parseDerivationPath(derivationPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse derivation path: %w", err)
	}

	key := masterKey
	for _, index := range path {
		key, err = key.NewChildKey(index)
		if err != nil {
			return nil, "", fmt.Errorf("failed to derive child key: %w", err)
		}
	}

	// Convert to ECDSA private key
	privateKey, err := crypto.ToECDSA(key.Key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to convert to ECDSA private key: %w", err)
	}

	// Get the Ethereum address
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	return privateKey, address, nil
}

// parseDerivationPath parses a derivation path string into uint32 indexes
func parseDerivationPath(path string) ([]uint32, error) {
	if !strings.HasPrefix(path, "m/") {
		return nil, fmt.Errorf("derivation path must start with 'm/'")
	}

	path = path[2:]
	if path == "" {
		return []uint32{}, nil
	}

	parts := strings.Split(path, "/")
	indexes := make([]uint32, len(parts))

	for i, part := range parts {
		var index uint64
		var err error

		if strings.HasSuffix(part, "'") {
			part = part[:len(part)-1]
			index, err = strconv.ParseUint(part, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid derivation path component: %s", part)
			}
			index += 0x80000000
		} else {
			index, err = strconv.ParseUint(part, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid derivation path component: %s", part)
			}
		}

		indexes[i] = uint32(index)
	}

	return indexes, nil
}

// confirmSigning asks the user to confirm they want to sign the transaction
func confirmSigning() bool {
	color.Yellow("Are you sure you want to sign this transaction? (y/N): ")
	fmt.Print("> ")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}