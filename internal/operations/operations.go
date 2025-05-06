package operations

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
)

// ELExitData represents a validator and its exit data
type ELExitData struct {
	Pubkey          string
	Amount          *big.Int
	ConfirmFullExit bool
}

// Operation defines the interface for all validator operations
type Operation interface {
	Execute() error
}

// BaseOperation contains common fields for all operations
type BaseOperation struct {
	Client          *ethclient.Client
	PrivateKey      *ecdsa.PrivateKey
	ContractAddress common.Address
	ABI             abi.ABI
}

// SendTransaction sends a transaction with the given data and value
func SendTransaction(client *ethclient.Client, privateKey *ecdsa.PrivateKey,
	contract common.Address, data []byte, value *uint256.Int) error {
	// Implementation will be in transaction package
	return nil
}
