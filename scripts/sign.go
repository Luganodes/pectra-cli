package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Luganodes/Pectra-CLI/internal/config"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func main() {
	privateKey, err := config.GetPrivateKey()
	if err != nil {
		log.Fatalf("Failed to get private key: %v", err)
	}

	// Determine input filename
	inputFile := "unsigned_txn.json"
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
	}

	// Read the unsigned transaction from file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", inputFile, err)
	}

	// Parse the JSON
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

	signedAuthorization, err := types.SignSetCode(privateKey, tx.SetCodeAuthorizations()[0])
	if err != nil {
		log.Fatalf("failed to sign the authorization: %v", err)
	}

	tx.SetCodeAuthorizations()[0] = signedAuthorization

	tx, err = types.SignTx(tx, types.LatestSignerForChainID(tx.ChainId()), privateKey)
	if err != nil {
		log.Fatalf("failed to sign the transaction: %v", err)
	}

	txBytes, err = rlp.EncodeToBytes(tx)
	if err != nil {
		log.Fatalf("failed to encode transaction: %v", err)
	}

	// Write the signed transaction to a file
	signedData := map[string]string{
		"signedTransaction": hex.EncodeToString(txBytes),
	}

	jsonData, err := json.MarshalIndent(signedData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal to JSON: %v", err)
	}

	outputFile := "signed_txn.json"
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		log.Fatalf("Failed to write to %s: %v", outputFile, err)
	}

	fmt.Printf("Signed transaction written to %s\n", outputFile)
}
