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

// SwitchOperation represents a batch switch operation
type SwitchOperation struct {
	BaseOperation
	Validators         []string
	AmountPerValidator *big.Int
}

// Execute performs the batch switch operation
func (op *SwitchOperation) Execute() error {
	if len(op.Validators) == 0 {
		return fmt.Errorf("no validators specified for switch operation")
	}

	op.Validators = utils.RemoveDuplicateValidators(op.Validators)

	if len(op.Validators) > 200 {
		return fmt.Errorf("a maximum of 200 validators can be switched at a time")
	}

	// Validate source validator public keys
	if err := utils.ValidateValidatorPubkeys(op.Validators); err != nil {
		return fmt.Errorf("invalid source validator public key: %w", err)
	}

	// Use provided amount or default to 1
	amountPerValidator := op.AmountPerValidator
	if amountPerValidator == nil {
		color.Yellow("Amount per validator is not set, using default value of 1")
		amountPerValidator = big.NewInt(1)
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
	value.Mul(big.NewInt(int64(len(op.Validators))), amountPerValidator)
	color.Cyan("Sending transaction with value: %v (for %d validators at %d each)",
		value, len(op.Validators), amountPerValidator)

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
