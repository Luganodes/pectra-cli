package transaction

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/holiman/uint256"
)

// SendTransactionUsingAuthorization sends a transaction with authorization
func SendTransactionUsingAuthorization(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contract common.Address, data []byte, value *uint256.Int) error {
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get the chain ID: %w", err)
	}
	
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
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

	color.Cyan("Transaction sent: https://hoodi.etherscan.io/tx/%s", tx.Hash().Hex())
	
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for the transaction to be mined: %w", err)
	}
	
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transaction failed")
	}
	
	color.Green("Transaction successful")
	return nil
}