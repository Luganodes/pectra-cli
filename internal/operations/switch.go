package operations

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	"github.com/holiman/uint256"
	"github.com/mannan-goyal/0x04/internal/transaction"
)

// SwitchOperation represents a batch switch operation
type SwitchOperation struct {
	BaseOperation
	Validators         []string
	AmountPerValidator int64
}

// Execute performs the batch switch operation
func (op *SwitchOperation) Execute() error {
	if len(op.Validators) == 0 {
		return fmt.Errorf("no validators specified for switch operation")
	}

	// Use provided amount or default to 1
	amountPerValidator := op.AmountPerValidator
	if amountPerValidator <= 0 {
		color.Yellow("Amount per validator is not set, using default value of 1")
		amountPerValidator = 1
	}

	pubkeys := [][]byte{}
	for _, validator := range op.Validators {
		pubkeys = append(pubkeys, common.FromHex(validator))
	}

	data, err := op.ABI.Pack("batchSwitch", pubkeys)
	if err != nil {
		return fmt.Errorf("failed to pack the data: %w", err)
	}

	value := new(big.Int)
	value.Mul(big.NewInt(int64(len(op.Validators))), big.NewInt(amountPerValidator))
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)",
		value, len(op.Validators), amountPerValidator)

	return transaction.SendTransactionUsingAuthorization(
		op.Client,
		op.PrivateKey,
		op.ContractAddress,
		data,
		uint256.NewInt(uint64(value.Int64())),
		op.ExplorerUrl,
	)
}
