package operations

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	"github.com/holiman/uint256"
	"github.com/mannan-goyal/0x04/internal/config"
	"github.com/mannan-goyal/0x04/internal/transaction"
)

// elExitOperation represents a batch EL exit operation
type ELExitOperation struct {
	BaseOperation
	Validators         map[string]config.ELExitDetails
	AmountPerValidator int64
}

// Execute performs the batch EL exit operation
func (op *ELExitOperation) Execute() error {
	if len(op.Validators) == 0 {
		return fmt.Errorf("no validators specified for EL exit operation")
	}

	// Use provided amount or default to 1
	amountPerValidator := op.AmountPerValidator
	if amountPerValidator <= 0 {
		color.Yellow("Amount per validator is not set, using default value of 1")
		amountPerValidator = 1
	}

	// Create a slice of ExitData structs to match the contract's expected input
	exitData := []struct {
		Pubkey     []byte
		Amount     uint64
		IsFullExit bool
	}{}

	for pubkey, details := range op.Validators {
		// Convert amount to uint64 (contract expects uint64)
		amountUint64 := uint64(details.Amount)

		// If amount is 0, we need the confirmFullExit flag
		isZeroAmount := details.Amount == 0
		if isZeroAmount && !details.ConfirmFullExit {
			return fmt.Errorf(color.RedString("validator %s has zero amount but confirmFullExit is not set. This exit will fail"), pubkey)
		}

		if !isZeroAmount && details.ConfirmFullExit {
			return fmt.Errorf(color.RedString("validator %s doesn't have a zero amount but confirmFullExit is set. This exit will fail"), pubkey)
		}

		exitData = append(exitData, struct {
			Pubkey     []byte
			Amount     uint64
			IsFullExit bool
		}{
			Pubkey:     common.FromHex(pubkey),
			Amount:     amountUint64,
			IsFullExit: details.ConfirmFullExit,
		})
	}

	// Pack the data for the contract call
	data, err := op.ABI.Pack("batchELExit", exitData)
	if err != nil {
		return fmt.Errorf("failed to pack the data: %w", err)
	}

	value := new(big.Int)
	value.Mul(big.NewInt(int64(len(exitData))), big.NewInt(amountPerValidator))
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)",
		value, len(exitData), amountPerValidator)

	return transaction.SendTransactionUsingAuthorization(
		op.Client,
		op.PrivateKey,
		op.ContractAddress,
		data,
		uint256.NewInt(uint64(value.Int64())),
		op.ExplorerUrl,
	)
}
