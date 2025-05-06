package operations

import (
	"fmt"
	"math/big"
	"strings"

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

	elExits := []ELExitData{}
	for pubkey, details := range op.Validators {
		// The amount is already in Gwei in the JSON
		amountGwei := big.NewInt(int64(details.Amount))
		
		// If amount is 0, we need the confirmFullExit flag
		isZeroAmount := details.Amount == 0 || amountGwei.Cmp(big.NewInt(0)) == 0
		if isZeroAmount && !details.ConfirmFullExit {
			color.Yellow("Warning: Validator %s has zero amount but confirmFullExit is not set. This exit will fail.", pubkey)
		}
		
		elExits = append(elExits, ELExitData{
			Pubkey:          pubkey,
			Amount:          amountGwei,
			ConfirmFullExit: details.ConfirmFullExit,
		})
	}

	// Data structure is now bytes[3][] to include the confirmation flag
	exitData := [][][]byte{}
	
	for _, validator := range elExits {
		hexAmount := validator.Amount.Text(16)
		paddedAmount := strings.Repeat("0", 16-len(hexAmount)) + hexAmount
		
		// Create confirmation flag byte - 0x01 for true, 0x00 for false
		var confirmFlag byte
		if validator.ConfirmFullExit {
			confirmFlag = 0x01
		} else {
			confirmFlag = 0x00
		}
		
		exitData = append(exitData, [][]byte{
			common.FromHex(validator.Pubkey),           // Validator pubkey
			common.FromHex(paddedAmount),               // Amount
			[]byte{confirmFlag},                        // Confirmation flag
		})
	}

	data, err := op.ABI.Pack("batchELExit", exitData)
	if err != nil {
		return fmt.Errorf("failed to pack the data: %w", err)
	}

	value := new(big.Int)
	value.Mul(big.NewInt(int64(len(elExits))), big.NewInt(amountPerValidator))
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)",
		value, len(elExits), amountPerValidator)

	return transaction.SendTransactionUsingAuthorization(
		op.Client, 
		op.PrivateKey, 
		op.ContractAddress, 
		data, 
		uint256.NewInt(uint64(value.Int64())),
	)
}