package transaction

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Luganodes/Pectra-CLI/internal/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/fatih/color"

	// "github.com/fatih/color"
	"github.com/holiman/uint256"
)

// SendTransactionUsingAuthorization sends a transaction with authorization
func SendTransactionUsingAuthorization(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contract common.Address, data []byte, value *uint256.Int, explorerURL string, airgapped bool) error {
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get the chain ID: %w", err)
	}

	var fromAddress common.Address
	var nonce uint64

	if privateKey != nil {
		fromAddress = crypto.PubkeyToAddress(privateKey.PublicKey)
		nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			return fmt.Errorf("failed to get nonce: %w", err)
		}
	} else if airgapped {
		// In airgapped mode without privateKey, use a placeholder address and nonce
		// The actual values will be provided during signing
		addressStr, err := config.GetPublicKey()
		if err != nil {
			return fmt.Errorf("failed to get public key: %w", err)
		}
		fromAddress = common.HexToAddress(addressStr)
		nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			return fmt.Errorf("failed to get nonce: %w", err)
		}
	} else {
		return fmt.Errorf("private key is required for non-airgapped mode")
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	tipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get the gas tip cap: %w", err)
	}

	authorization := types.SetCodeAuthorization{
		ChainID: *uint256.NewInt(chainID.Uint64()),
		Address: contract,
		Nonce:   nonce + 1,
	}

	if airgapped {
		tx := types.NewTx(&types.SetCodeTx{
			ChainID:   uint256.NewInt(chainID.Uint64()),
			Nonce:     nonce,
			GasTipCap: uint256.NewInt(tipCap.Uint64()),
			GasFeeCap: uint256.NewInt(gasPrice.Uint64()),
			Gas:       uint64(30000000),
			To:        fromAddress,
			Value:     value,
			Data:      data,
			AuthList:  []types.SetCodeAuthorization{authorization},
		})

		// serialize the transaction to hex
		txBytes, err := rlp.EncodeToBytes(tx)
		if err != nil {
			return fmt.Errorf("failed to serialize the transaction: %w", err)
		}

		// Create JSON structure for the transaction
		txData := map[string]string{
			"unsignedTransaction": hex.EncodeToString(txBytes),
			"chainId":             chainID.String(),
		}

		// Marshal to JSON
		jsonData, err := json.MarshalIndent(txData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal transaction to JSON: %w", err)
		}

		// Write to file
		err = os.WriteFile("unsigned_txn.json", jsonData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write transaction to file: %w", err)
		}

		color.Green("Transaction data written to unsigned_txn.json")
	} else {
		signedAuthorization, err := types.SignSetCode(privateKey, authorization)
		if err != nil {
			return fmt.Errorf("failed to sign the authorization: %w", err)
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
			return fmt.Errorf("failed to sign the transaction: %w", err)
		}

		err = client.SendTransaction(context.Background(), tx)
		if err != nil {
			return fmt.Errorf("failed to send the transaction: %w", err)
		}

		color.Cyan("Transaction sent: %s/tx/%s", explorerURL, tx.Hash().Hex())

		color.Cyan("Waiting for transaction to be included in a block...")

		receipt, err := bind.WaitMined(context.Background(), client, tx)
		if err != nil {
			return fmt.Errorf("failed to wait for the transaction to be included in a block: %w", err)
		}

		if receipt.Status != types.ReceiptStatusSuccessful {
			return fmt.Errorf("transaction failed")
		}

		color.Green("Transaction successful")
	}

	return nil
}

// BroadcastTransactionFromFile broadcasts a signed transaction from the specified file
func BroadcastTransactionFromFile(filePath string, configPath string) error {
	// Read signed transaction from specified file
	color.Cyan("Reading transaction from file: %s", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read transaction file %s: %w", filePath, err)
	}

	// Parse the JSON
	var signedData struct {
		SignedTransaction string `json:"signedTransaction"`
	}

	if err := json.Unmarshal(data, &signedData); err != nil {
		return fmt.Errorf("failed to parse JSON from %s: %w", filePath, err)
	}

	// Check if SignedTransaction field exists
	if signedData.SignedTransaction == "" {
		return fmt.Errorf("invalid JSON format: missing 'signedTransaction' field in %s", filePath)
	}

	hexTx := signedData.SignedTransaction

	// Remove 0x prefix if present
	if len(hexTx) > 2 && hexTx[0:2] == "0x" {
		hexTx = hexTx[2:]
	}

	// Decode hex to bytes
	txBytes, err := hex.DecodeString(hexTx)
	if err != nil {
		return fmt.Errorf("failed to decode hex string: %w", err)
	}

	// Decode transaction
	tx := new(types.Transaction)
	err = rlp.DecodeBytes(txBytes, tx)
	if err != nil {
		return fmt.Errorf("failed to decode transaction: %w", err)
	}

	// Get chain ID from transaction
	chainID := tx.ChainId()
	color.Green("Transaction decoded - hash: %s, Chain ID: %s", tx.Hash().Hex(), chainID.String())

	// Connect to Ethereum client
	// Use the RPC URL from the config or chain ID mapping

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	rpcURL := cfg.RPCUrl
	if rpcURL == "" {
		return fmt.Errorf("RPC URL is not set in the config")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}
	color.Cyan("Connected to the Ethereum client")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Send transaction
	err = client.SendTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	color.Cyan("Transaction hash: %s", tx.Hash().Hex())

	color.Cyan("Waiting for transaction to be included in a block...")

	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for the transaction to be included in a block: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transaction failed")
	}

	color.Green("Transaction successful")


	return nil
}