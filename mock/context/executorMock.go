package mock

import (
	"testing"

	"github.com/ElrondNetwork/wasm-vm/arwen"
	"github.com/ElrondNetwork/wasm-vm/executor"
	worldmock "github.com/ElrondNetwork/wasm-vm/mock/world"
	"github.com/ElrondNetwork/wasm-vm/wasmer"
)

// ExecutorMock can be passed to RuntimeContext as an InstanceBuilder to
// create mocked Wasmer instances.
type ExecutorMock struct {
	InstanceMap map[string]InstanceMock
	World       *worldmock.MockWorld
}

// NewExecutorMock constructs a new InstanceBuilderMock
func NewExecutorMock(world *worldmock.MockWorld) *ExecutorMock {
	return &ExecutorMock{
		InstanceMap: make(map[string]InstanceMock),
		World:       world,
	}
}

// SetOpcodeCosts should set gas costs, but it does nothing in the case of this mock.
func (builder *ExecutorMock) SetOpcodeCosts(opcodeCosts *[executor.OpcodeCount]uint32) {
}

// SetRkyvSerializationEnabled controls a Wasmer flag, but it does nothing in the case of this mock.
func (builder *ExecutorMock) SetRkyvSerializationEnabled(enabled bool) {
}

// SetSIGSEGVPassthrough controls a Wasmer flag, but it does nothing in the case of this mock.
func (builder *ExecutorMock) SetSIGSEGVPassthrough() {
}

// CreateAndStoreInstanceMock creates a new InstanceMock and registers it as a
// smart contract account in the World, using `code` as the address of the account
func (builder *ExecutorMock) CreateAndStoreInstanceMock(t testing.TB, host arwen.VMHost, code []byte, codeHash []byte, codeMetadata []byte, ownerAddress []byte, shardID uint32, balance int64, createAccount bool) *InstanceMock {
	instance := NewInstanceMock(code)
	instance.Address = code
	instance.T = t
	instance.Host = host
	builder.InstanceMap[string(code)] = *instance

	if createAccount {
		account := builder.World.AcctMap.CreateSmartContractAccountWithCodeHash(nil, code, code, codeHash, builder.World)
		account.SetBalance(balance)
		account.ShardID = shardID
		account.CodeMetadata = codeMetadata
		account.OwnerAddress = ownerAddress
	}

	return instance
}

// getNewCopyOfStoredInstance retrieves and initializes a stored Wasmer instance, or
// nil if it doesn't exist
func (builder *ExecutorMock) getNewCopyOfStoredInstance(code []byte, gasLimit uint64) (executor.InstanceHandler, bool) {
	// this is a map to InstanceMock(s), and copies of these instances will be returned (as the method name indicates)
	instance, ok := builder.InstanceMap[string(code)]
	if ok {
		instance.SetPointsUsed(0)
		instance.SetGasLimit(gasLimit)
		return &instance, true
	}
	return nil, false
}

// NewInstanceWithOptions attempts to load a prepared instance using
// GetStoredInstance; if it doesn't exist, it creates a true Wasmer
// instance with the provided contract code.
func (builder *ExecutorMock) NewInstanceWithOptions(
	contractCode []byte,
	options executor.CompilationOptions,
) (executor.InstanceHandler, error) {

	instance, ok := builder.getNewCopyOfStoredInstance(contractCode, options.GasLimit)
	if ok {
		return instance, nil
	}
	return wasmer.NewInstanceWithOptions(contractCode, options)
}

// NewInstanceFromCompiledCodeWithOptions attempts to load a prepared instance
// using GetStoredInstance; if it doesn't exist, it creates a true Wasmer
// instance with the provided precompiled code.
func (builder *ExecutorMock) NewInstanceFromCompiledCodeWithOptions(
	compiledCode []byte,
	options executor.CompilationOptions,
) (executor.InstanceHandler, error) {
	instance, ok := builder.getNewCopyOfStoredInstance(compiledCode, options.GasLimit)
	if ok {
		return instance, nil
	}
	return wasmer.NewInstanceFromCompiledCodeWithOptions(compiledCode, options)
}
