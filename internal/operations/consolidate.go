package operations

import (
	"fmt"
	"math/big"

	"github.com/Luganodes/Pectra-CLI/internal/transaction"
	"github.com/Luganodes/Pectra-CLI/internal/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	"github.com/holiman/uint256"
)

// ConsolidateOperation represents a batch consolidation operation
type ConsolidateOperation struct {
	BaseOperation
	SourceValidators   []string
	TargetValidator    string
	AmountPerValidator int64
}

// Execute performs the batch consolidation operation
func (op *ConsolidateOperation) Execute() error {
	if len(op.SourceValidators) == 0 || op.TargetValidator == "" {
		return fmt.Errorf("source or target validators not specified for consolidate operation")
	}

	op.SourceValidators = utils.RemoveDuplicateValidators(op.SourceValidators)

	if len(op.SourceValidators) > 63 {
		return fmt.Errorf("a maximum of 63 validators can be consolidated at a time")
	}

	// Validate source validator public keys
	if err := utils.ValidateValidatorPubkeys(op.SourceValidators); err != nil {
		return fmt.Errorf("invalid source validator public key: %w", err)
	}

	// Validate target validator public key
	if err := utils.ValidateValidatorPubkeys([]string{op.TargetValidator}); err != nil {
		return fmt.Errorf("invalid target validator public key: %w", err)
	}

	// Check if target validator is in source validators
	for _, sourceValidator := range op.SourceValidators {
		if sourceValidator == op.TargetValidator {
			return fmt.Errorf("target validator (%s) cannot be in the list of source validators", op.TargetValidator)
		}
	}

	// Use provided amount or default to 1
	amountPerValidator := op.AmountPerValidator
	if amountPerValidator <= 0 {
		color.Yellow("Amount per validator is not set, using default value of 1")
		amountPerValidator = 1
	}

	pubkeys := [][]byte{}
	for _, validator := range op.SourceValidators {
		pubkeys = append(pubkeys, common.FromHex(validator))
	}
	target := common.FromHex(op.TargetValidator)

	data, err := op.ABI.Pack("batchConsolidation", pubkeys, target)
	if err != nil {
		return fmt.Errorf("failed to pack the data: %w", err)
	}

	value := new(big.Int)
	value.Mul(big.NewInt(int64(len(pubkeys))), big.NewInt(amountPerValidator))
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)",
		value, len(pubkeys), amountPerValidator)

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
