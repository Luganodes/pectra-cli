package operations

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	"github.com/holiman/uint256"
	"github.com/mannan-goyal/0x04/internal/transaction"
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
	)
}