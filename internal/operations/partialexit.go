package operations

import (
	"fmt"
	"math/big"

	"github.com/Luganodes/Pectra-CLI/internal/config"
	"github.com/Luganodes/Pectra-CLI/internal/transaction"
	"github.com/Luganodes/Pectra-CLI/internal/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	"github.com/holiman/uint256"
)

// elExitOperation represents a batch EL exit operation
type ELExitOperation struct {
	BaseOperation
	Validators         map[string]config.ELExitDetails
	AmountPerValidator *big.Int
}

// Execute performs the batch EL exit operation
func (op *ELExitOperation) Execute() error {
	if len(op.Validators) == 0 {
		return fmt.Errorf("no validators specified for EL exit operation")
	}

	if len(op.Validators) > 200 {
		return fmt.Errorf("a maximum of 200 validators can be exited at a time")
	}

	// Validate public keys
	pubkeysToValidate := make([]string, 0, len(op.Validators))
	for pubkey := range op.Validators {
		pubkeysToValidate = append(pubkeysToValidate, pubkey)
	}
	if err := utils.ValidateValidatorPubkeys(pubkeysToValidate); err != nil {
		return fmt.Errorf("validator public key validation failed: %w", err)
	}

	// Use provided amount or default to 1
	amountPerValidator := op.AmountPerValidator
	if amountPerValidator == nil {
		color.Yellow("Amount per validator is not set, using default value of 1")
		amountPerValidator = big.NewInt(1)
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
	value.Mul(big.NewInt(int64(len(exitData))), amountPerValidator)
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)",
		value, len(exitData), amountPerValidator)

	return transaction.SendTransactionUsingAuthorization(
		op.Client,
		op.PrivateKey,
		op.ContractAddress,
		data,
		uint256.NewInt(uint64(value.Int64())),
		op.ExplorerUrl,
		op.Airgapped,
	)
}
