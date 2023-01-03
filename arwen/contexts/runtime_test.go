package contexts

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/ElrondNetwork/wasm-vm/arwen"
	"github.com/ElrondNetwork/wasm-vm/arwen/elrondapi"
	"github.com/ElrondNetwork/wasm-vm/config"
	"github.com/ElrondNetwork/wasm-vm/crypto/factory"
	"github.com/ElrondNetwork/wasm-vm/executor"
	contextmock "github.com/ElrondNetwork/wasm-vm/mock/context"
	worldmock "github.com/ElrondNetwork/wasm-vm/mock/world"
	"github.com/ElrondNetwork/wasm-vm/wasmer"
	"github.com/stretchr/testify/require"
)

const counterWasmCode = "./../../test/contracts/counter/output/counter.wasm"

var vmType = []byte("type")

func InitializeArwenAndWasmer() *contextmock.VMHostMock {
	gasSchedule := config.MakeGasMapForTests()
	gasCostConfig, _ := config.CreateGasConfig(gasSchedule)
	wasmer.SetOpcodeCosts(gasCostConfig.WASMOpcodeCost)

	host := &contextmock.VMHostMock{}

	mockMetering := &contextmock.MeteringContextMock{}
	mockMetering.SetGasSchedule(gasSchedule)
	host.MeteringContext = mockMetering
	host.BlockchainContext, _ = NewBlockchainContext(host, worldmock.NewMockWorld())
	host.OutputContext, _ = NewOutputContext(host)
	host.CryptoHook = factory.NewVMCrypto()
	return host
}

func makeDefaultRuntimeContext(t *testing.T, host arwen.VMHost) *runtimeContext {
	exec, err := wasmer.ExecutorFactory().CreateExecutor(executor.ExecutorFactoryArgs{
		VMHooks: elrondapi.NewElrondApi(host),
	})
	require.Nil(t, err)
	runtimeCtx, err := NewRuntimeContext(
		host,
		vmType,
		builtInFunctions.NewBuiltInFunctionContainer(),
		exec,
	)
	require.Nil(t, err)
	require.NotNil(t, runtimeCtx)

	return runtimeCtx
}

func TestNewRuntimeContext(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	require.Equal(t, &vmcommon.ContractCallInput{}, runtimeCtx.vmInput)
	require.Equal(t, []byte{}, runtimeCtx.codeAddress)
	require.Equal(t, "", runtimeCtx.callFunction)
	require.Equal(t, false, runtimeCtx.readOnly)
}

func TestRuntimeContext_InitState(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.vmInput = nil
	runtimeCtx.codeAddress = []byte("some address")
	runtimeCtx.callFunction = "a function"
	runtimeCtx.readOnly = true

	runtimeCtx.InitState()

	require.Equal(t, &vmcommon.ContractCallInput{}, runtimeCtx.vmInput)
	require.Equal(t, []byte{}, runtimeCtx.codeAddress)
	require.Equal(t, "", runtimeCtx.callFunction)
	require.Equal(t, false, runtimeCtx.readOnly)
}

func TestRuntimeContext_NewWasmerInstance(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	var dummy []byte
	err := runtimeCtx.StartWasmerInstance(dummy, gasLimit, false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, wasmer.ErrInvalidBytecode))

	gasLimit = uint64(100000000)
	dummy = []byte("contract")
	err = runtimeCtx.StartWasmerInstance(dummy, gasLimit, false)
	require.NotNil(t, err)

	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err = runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)
	require.Equal(t, arwen.BreakpointNone, runtimeCtx.GetRuntimeBreakpointValue())
}

func TestRuntimeContext_IsFunctionImported(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)
	require.Equal(t, arwen.BreakpointNone, runtimeCtx.GetRuntimeBreakpointValue())

	// These API functions exist, and are imported by 'counter'
	require.True(t, runtimeCtx.IsFunctionImported("int64storageLoad"))
	require.True(t, runtimeCtx.IsFunctionImported("int64storageStore"))
	require.True(t, runtimeCtx.IsFunctionImported("int64finish"))

	// These API functions exist, but are not imported by 'counter'
	require.False(t, runtimeCtx.IsFunctionImported("transferValue"))
	require.False(t, runtimeCtx.IsFunctionImported("executeOnSameContext"))
	require.False(t, runtimeCtx.IsFunctionImported("asyncCall"))

	// These API functions don't even exist
	require.False(t, runtimeCtx.IsFunctionImported(""))
	require.False(t, runtimeCtx.IsFunctionImported("*"))
	require.False(t, runtimeCtx.IsFunctionImported("$@%"))
	require.False(t, runtimeCtx.IsFunctionImported("doesNotExist"))
}

func TestRuntimeContext_StateSettersAndGetters(t *testing.T) {
	host := &contextmock.VMHostMock{}

	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	arguments := [][]byte{[]byte("argument 1"), []byte("argument 2")}
	esdtTransfer := &vmcommon.ESDTTransfer{
		ESDTValue:      big.NewInt(4242),
		ESDTTokenName:  []byte("random_token"),
		ESDTTokenType:  uint32(core.NonFungible),
		ESDTTokenNonce: 94,
	}

	vmInput := vmcommon.VMInput{
		CallerAddr:    []byte("caller"),
		Arguments:     arguments,
		CallValue:     big.NewInt(0),
		ESDTTransfers: []*vmcommon.ESDTTransfer{esdtTransfer},
	}
	callInput := &vmcommon.ContractCallInput{
		VMInput:       vmInput,
		RecipientAddr: []byte("recipient"),
		Function:      "test function",
	}

	runtimeCtx.InitStateFromContractCallInput(callInput)
	require.Equal(t, []byte("caller"), runtimeCtx.GetVMInput().CallerAddr)
	require.Equal(t, []byte("recipient"), runtimeCtx.GetContextAddress())
	require.Equal(t, "test function", runtimeCtx.FunctionName())
	require.Equal(t, vmType, runtimeCtx.GetVMType())
	require.Equal(t, arguments, runtimeCtx.Arguments())

	runtimeInput := runtimeCtx.GetVMInput()
	require.Zero(t, big.NewInt(4242).Cmp(runtimeInput.ESDTTransfers[0].ESDTValue))
	require.True(t, bytes.Equal([]byte("random_token"), runtimeInput.ESDTTransfers[0].ESDTTokenName))
	require.Equal(t, uint32(core.NonFungible), runtimeInput.ESDTTransfers[0].ESDTTokenType)
	require.Equal(t, uint64(94), runtimeInput.ESDTTransfers[0].ESDTTokenNonce)

	vmInput2 := vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: []byte("caller2"),
			Arguments:  arguments,
			CallValue:  big.NewInt(0),
		},
	}
	runtimeCtx.SetVMInput(&vmInput2)
	require.Equal(t, []byte("caller2"), runtimeCtx.GetVMInput().CallerAddr)

	runtimeCtx.SetCodeAddress([]byte("smartcontract"))
	require.Equal(t, []byte("smartcontract"), runtimeCtx.codeAddress)
}

func TestRuntimeContext_PushPopInstance(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	instance := runtimeCtx.instance

	runtimeCtx.pushInstance()
	runtimeCtx.instance = nil
	require.Equal(t, 1, len(runtimeCtx.instanceStack))

	runtimeCtx.popInstance()
	require.NotNil(t, runtimeCtx.instance)
	require.Equal(t, instance, runtimeCtx.instance)
	require.Equal(t, 0, len(runtimeCtx.instanceStack))

	runtimeCtx.pushInstance()
	require.Equal(t, 1, len(runtimeCtx.instanceStack))
}

func TestRuntimeContext_PushPopState(t *testing.T) {
	host := &contextmock.VMHostMock{}
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	vmInput := vmcommon.VMInput{
		CallerAddr:    []byte("caller"),
		GasProvided:   1000,
		CallValue:     big.NewInt(0),
		ESDTTransfers: make([]*vmcommon.ESDTTransfer, 0),
	}

	funcName := "test_func"
	scAddress := []byte("smartcontract")
	input := &vmcommon.ContractCallInput{
		VMInput:       vmInput,
		RecipientAddr: scAddress,
		Function:      funcName,
	}
	runtimeCtx.InitStateFromContractCallInput(input)

	runtimeCtx.instance = &wasmer.WasmerInstance{}
	runtimeCtx.PushState()
	require.Equal(t, 1, len(runtimeCtx.stateStack))

	// change state
	runtimeCtx.SetCodeAddress([]byte("dummy"))
	runtimeCtx.SetVMInput(nil)
	runtimeCtx.SetReadOnly(true)

	require.Equal(t, []byte("dummy"), runtimeCtx.codeAddress)
	require.Nil(t, runtimeCtx.GetVMInput())
	require.True(t, runtimeCtx.ReadOnly())

	runtimeCtx.PopSetActiveState()

	// check state was restored correctly
	require.Equal(t, scAddress, runtimeCtx.GetContextAddress())
	require.Equal(t, funcName, runtimeCtx.FunctionName())
	require.Equal(t, input, runtimeCtx.GetVMInput())
	require.False(t, runtimeCtx.ReadOnly())
	require.Nil(t, runtimeCtx.Arguments())

	runtimeCtx.instance = &wasmer.WasmerInstance{}
	runtimeCtx.PushState()
	require.Equal(t, 1, len(runtimeCtx.stateStack))

	runtimeCtx.instance = &wasmer.WasmerInstance{}
	runtimeCtx.PushState()
	require.Equal(t, 2, len(runtimeCtx.stateStack))

	runtimeCtx.PopDiscard()
	require.Equal(t, 1, len(runtimeCtx.stateStack))

	runtimeCtx.ClearStateStack()
	require.Equal(t, 0, len(runtimeCtx.stateStack))
}

func TestRuntimeContext_CountContractInstancesOnStack(t *testing.T) {
	alpha := []byte("alpha")
	beta := []byte("beta")
	gamma := []byte("gamma")

	host := &contextmock.VMHostMock{}

	testVmType := []byte("type")
	exec, err := wasmer.ExecutorFactory().CreateExecutor(executor.ExecutorFactoryArgs{
		VMHooks: elrondapi.NewElrondApi(host),
	})
	require.Nil(t, err)
	runtime, _ := NewRuntimeContext(
		host,
		testVmType,
		builtInFunctions.NewBuiltInFunctionContainer(),
		exec,
	)

	vmInput := vmcommon.VMInput{
		CallerAddr:  []byte("caller"),
		GasProvided: 1000,
		CallValue:   big.NewInt(0),
	}
	input := &vmcommon.ContractCallInput{
		VMInput:  vmInput,
		Function: "function",
	}

	input.RecipientAddr = alpha
	runtime.InitStateFromContractCallInput(input)
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(alpha))
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(beta))
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(gamma))

	runtime.instance = &wasmer.WasmerInstance{}
	runtime.PushState()
	input.RecipientAddr = beta
	runtime.InitStateFromContractCallInput(input)
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(alpha))
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(beta))
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(gamma))

	runtime.instance = &wasmer.WasmerInstance{}
	runtime.PushState()
	input.RecipientAddr = gamma
	runtime.InitStateFromContractCallInput(input)
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(alpha))
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(beta))
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(gamma))

	runtime.instance = &wasmer.WasmerInstance{}
	runtime.PushState()
	input.RecipientAddr = alpha
	runtime.InitStateFromContractCallInput(input)
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(alpha))
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(beta))
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(gamma))

	runtime.PushState()
	input.RecipientAddr = gamma
	runtime.InitStateFromContractCallInput(input)
	require.Equal(t, uint64(2), runtime.CountSameContractInstancesOnStack(alpha))
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(beta))
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(gamma))

	runtime.PopSetActiveState()
	runtime.PopSetActiveState()
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(alpha))
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(beta))
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(gamma))

	runtime.PopDiscard()
	require.Equal(t, uint64(1), runtime.CountSameContractInstancesOnStack(alpha))
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(beta))
	require.Equal(t, uint64(0), runtime.CountSameContractInstancesOnStack(gamma))
}

func TestRuntimeContext_Instance(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	gasPoints := uint64(100)
	runtimeCtx.SetPointsUsed(gasPoints)
	require.Equal(t, gasPoints, runtimeCtx.GetPointsUsed())

	funcName := "increment"
	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
		},
		RecipientAddr: []byte("addr"),
		Function:      funcName,
	}
	runtimeCtx.InitStateFromContractCallInput(input)

	functionName, err := runtimeCtx.FunctionNameChecked()
	require.Nil(t, err)
	require.NotEmpty(t, functionName)

	input.Function = "func"
	runtimeCtx.InitStateFromContractCallInput(input)
	functionName, err = runtimeCtx.FunctionNameChecked()
	require.Equal(t, executor.ErrFuncNotFound, err)
	require.Empty(t, functionName)

	hasInitFunction := runtimeCtx.HasFunction(arwen.InitFunctionName)
	require.True(t, hasInitFunction)

	runtimeCtx.ClearWarmInstanceCache()
	require.Nil(t, runtimeCtx.instance)
}

func TestRuntimeContext_Breakpoints(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	mockOutput := &contextmock.OutputContextMock{
		OutputAccountMock: NewVMOutputAccount([]byte("address")),
	}
	mockOutput.OutputAccountMock.Code = []byte("code")
	mockOutput.SetReturnMessage("")

	host.OutputContext = mockOutput

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	// Set and get curent breakpoint value
	require.Equal(t, arwen.BreakpointNone, runtimeCtx.GetRuntimeBreakpointValue())
	runtimeCtx.SetRuntimeBreakpointValue(arwen.BreakpointOutOfGas)
	require.Equal(t, arwen.BreakpointOutOfGas, runtimeCtx.GetRuntimeBreakpointValue())

	runtimeCtx.SetRuntimeBreakpointValue(arwen.BreakpointNone)
	require.Equal(t, arwen.BreakpointNone, runtimeCtx.GetRuntimeBreakpointValue())

	// Signal user error
	mockOutput.SetReturnCode(vmcommon.Ok)
	mockOutput.SetReturnMessage("")
	runtimeCtx.SetRuntimeBreakpointValue(arwen.BreakpointNone)

	runtimeCtx.SignalUserError("something happened")
	require.Equal(t, arwen.BreakpointSignalError, runtimeCtx.GetRuntimeBreakpointValue())
	require.Equal(t, vmcommon.UserError, mockOutput.ReturnCode())
	require.Equal(t, "something happened", mockOutput.ReturnMessage())

	// Fail execution
	mockOutput.SetReturnCode(vmcommon.Ok)
	mockOutput.SetReturnMessage("")
	runtimeCtx.SetRuntimeBreakpointValue(arwen.BreakpointNone)

	runtimeCtx.FailExecution(nil)
	require.Equal(t, arwen.BreakpointExecutionFailed, runtimeCtx.GetRuntimeBreakpointValue())
	require.Equal(t, vmcommon.ExecutionFailed, mockOutput.ReturnCode())
	require.Equal(t, "execution failed", mockOutput.ReturnMessage())

	mockOutput.SetReturnCode(vmcommon.Ok)
	mockOutput.SetReturnMessage("")
	runtimeCtx.SetRuntimeBreakpointValue(arwen.BreakpointNone)
	require.Equal(t, arwen.BreakpointNone, runtimeCtx.GetRuntimeBreakpointValue())

	runtimeError := errors.New("runtime error")
	runtimeCtx.FailExecution(runtimeError)
	require.Equal(t, arwen.BreakpointExecutionFailed, runtimeCtx.GetRuntimeBreakpointValue())
	require.Equal(t, vmcommon.ExecutionFailed, mockOutput.ReturnCode())
	require.Equal(t, runtimeError.Error(), mockOutput.ReturnMessage())
}

func TestRuntimeContext_MemLoadStoreOk(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	memory := runtimeCtx.instance.GetMemory()

	memContents, err := runtimeCtx.MemLoad(10, 10)
	require.Nil(t, err)
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, memContents)

	pageSize := uint32(65536)
	require.Equal(t, 2*pageSize, memory.Length())

	memContents = []byte("test data")
	err = runtimeCtx.MemStore(10, memContents)
	require.Nil(t, err)
	require.Equal(t, 2*pageSize, memory.Length())

	memContents, err = runtimeCtx.MemLoad(10, 10)
	require.Nil(t, err)
	require.Equal(t, []byte{'t', 'e', 's', 't', ' ', 'd', 'a', 't', 'a', 0}, memContents)
}

func TestRuntimeContext_MemoryIsBlank(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	path := "./../../test/contracts/init-simple/output/init-simple.wasm"
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	memory := runtimeCtx.instance.GetMemory()
	err = memory.Grow(30)
	require.Nil(t, err)

	totalPages := 32
	memoryContents := memory.Data()
	require.Equal(t, memory.Length(), uint32(len(memoryContents)))
	require.Equal(t, totalPages*arwen.WASMPageSize, len(memoryContents))

	for i, value := range memoryContents {
		if value != byte(0) {
			msg := fmt.Sprintf("Non-zero value found at %d in Wasmer memory: 0x%X", i, value)
			require.Fail(t, msg)
		}
	}
}

func TestRuntimeContext_MemLoadCases(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	memory := runtimeCtx.instance.GetMemory()

	var offset int32
	var length int32
	// Offset too small
	offset = -3
	length = 10
	memContents, err := runtimeCtx.MemLoad(offset, length)
	require.True(t, errors.Is(err, arwen.ErrBadBounds))
	require.Nil(t, memContents)

	// Offset too larget
	offset = int32(memory.Length() + 1)
	length = 10
	memContents, err = runtimeCtx.MemLoad(offset, length)
	require.True(t, errors.Is(err, arwen.ErrBadBounds))
	require.Nil(t, memContents)

	// Negative length
	offset = 10
	length = -2
	memContents, err = runtimeCtx.MemLoad(offset, length)
	require.True(t, errors.Is(err, arwen.ErrNegativeLength))
	require.Nil(t, memContents)

	// Requested end too large
	memContents = []byte("test data")
	offset = int32(memory.Length() - 9)
	err = runtimeCtx.MemStore(offset, memContents)
	require.Nil(t, err)

	offset = int32(memory.Length() - 9)
	length = 9
	memContents, err = runtimeCtx.MemLoad(offset, length)
	require.Nil(t, err)
	require.Equal(t, []byte("test data"), memContents)

	offset = int32(memory.Length() - 8)
	length = 9
	memContents, err = runtimeCtx.MemLoad(offset, length)
	require.Nil(t, err)
	require.Equal(t, []byte{'e', 's', 't', ' ', 'd', 'a', 't', 'a', 0}, memContents)

	// Zero length
	offset = int32(memory.Length() - 8)
	length = 0
	memContents, err = runtimeCtx.MemLoad(offset, length)
	require.Nil(t, err)
	require.Equal(t, []byte{}, memContents)
}

func TestRuntimeContext_MemStoreCases(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(1)

	gasLimit := uint64(100000000)
	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	pageSize := uint32(65536)
	memory := runtimeCtx.instance.GetMemory()
	require.Equal(t, 2*pageSize, memory.Length())

	// Bad lower bounds
	memContents := []byte("test data")
	offset := int32(-2)
	err = runtimeCtx.MemStore(offset, memContents)
	require.True(t, errors.Is(err, arwen.ErrBadLowerBounds))

	// Memory growth
	require.Equal(t, 2*pageSize, memory.Length())
	offset = int32(memory.Length() - 4)
	err = runtimeCtx.MemStore(offset, memContents)
	require.Nil(t, err)
	require.Equal(t, 3*pageSize, memory.Length())

	// Bad upper bounds - forcing the Wasmer memory to grow more than a page at a
	// time is not allowed
	memContents = make([]byte, pageSize+100)
	offset = int32(memory.Length() - 50)
	err = runtimeCtx.MemStore(offset, memContents)
	require.True(t, errors.Is(err, arwen.ErrBadUpperBounds))
	require.Equal(t, 4*pageSize, memory.Length())

	// Write something, then overwrite, then overwrite with empty byte slice
	memContents = []byte("this is a message")
	offset = int32(memory.Length() - 100)
	err = runtimeCtx.MemStore(offset, memContents)
	require.Nil(t, err)

	memContents, err = runtimeCtx.MemLoad(offset, 17)
	require.Nil(t, err)
	require.Equal(t, []byte("this is a message"), memContents)

	memContents = []byte("this is something")
	err = runtimeCtx.MemStore(offset, memContents)
	require.Nil(t, err)

	memContents, err = runtimeCtx.MemLoad(offset, 17)
	require.Nil(t, err)
	require.Equal(t, []byte("this is something"), memContents)

	memContents = []byte{}
	err = runtimeCtx.MemStore(offset, memContents)
	require.Nil(t, err)

	memContents, err = runtimeCtx.MemLoad(offset, 17)
	require.Nil(t, err)
	require.Equal(t, []byte("this is something"), memContents)
}

func TestRuntimeContext_MemLoadStoreVsInstanceStack(t *testing.T) {
	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.SetMaxInstanceStackSize(2)

	gasLimit := uint64(100000000)
	path := counterWasmCode
	contractCode := arwen.GetSCCode(path)
	err := runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	// Write "test data1" to the WASM memory of the current instance
	memContents := []byte("test data1")
	err = runtimeCtx.MemStore(10, memContents)
	require.Nil(t, err)

	memContents, err = runtimeCtx.MemLoad(10, 10)
	require.Nil(t, err)
	require.Equal(t, []byte("test data1"), memContents)

	// Push the current instance down the instance stack
	runtimeCtx.pushInstance()
	require.Equal(t, 1, len(runtimeCtx.instanceStack))

	// Create a new Wasmer instance
	contractCode = arwen.GetSCCode(path)
	err = runtimeCtx.StartWasmerInstance(contractCode, gasLimit, false)
	require.Nil(t, err)

	// Write "test data2" to the WASM memory of the new instance
	memContents = []byte("test data2")
	err = runtimeCtx.MemStore(10, memContents)
	require.Nil(t, err)

	memContents, err = runtimeCtx.MemLoad(10, 10)
	require.Nil(t, err)
	require.Equal(t, []byte("test data2"), memContents)

	// Pop the initial instance from the stack, making it the 'current instance'
	runtimeCtx.popInstance()
	require.Equal(t, 0, len(runtimeCtx.instanceStack))

	// Check whether the previously-written string "test data1" is still in the
	// memory of the initial instance
	memContents, err = runtimeCtx.MemLoad(10, 10)
	require.Nil(t, err)
	require.Equal(t, []byte("test data1"), memContents)

	// Write "test data3" to the WASM memory of the initial instance (now current)
	memContents = []byte("test data3")
	err = runtimeCtx.MemStore(10, memContents)
	require.Nil(t, err)

	memContents, err = runtimeCtx.MemLoad(10, 10)
	require.Nil(t, err)
	require.Equal(t, []byte("test data3"), memContents)
}

func TestRuntimeContext_PopSetActiveStateIfStackIsEmptyShouldNotPanic(t *testing.T) {
	t.Parallel()

	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.PopSetActiveState()

	require.Equal(t, 0, len(runtimeCtx.stateStack))
}

func TestRuntimeContext_PopDiscardIfStackIsEmptyShouldNotPanic(t *testing.T) {
	t.Parallel()

	host := InitializeArwenAndWasmer()
	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()

	runtimeCtx.PopDiscard()

	require.Equal(t, 0, len(runtimeCtx.stateStack))
}

func TestRuntimeContext_PopInstanceIfStackIsEmptyShouldNotPanic(t *testing.T) {
	t.Parallel()

	host := InitializeArwenAndWasmer()

	runtimeCtx := makeDefaultRuntimeContext(t, host)
	defer runtimeCtx.ClearWarmInstanceCache()
	runtimeCtx.popInstance()

	require.Equal(t, 0, len(runtimeCtx.stateStack))
}
