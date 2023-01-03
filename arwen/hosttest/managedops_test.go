package hosttest

import (
	"testing"

	"github.com/ElrondNetwork/wasm-vm/arwen"
	"github.com/ElrondNetwork/wasm-vm/arwen/elrondapi"
	mock "github.com/ElrondNetwork/wasm-vm/mock/context"
	worldmock "github.com/ElrondNetwork/wasm-vm/mock/world"
	test "github.com/ElrondNetwork/wasm-vm/testcommon"
	"github.com/stretchr/testify/assert"
)

// wasm memory ~~~> managed buffer
func TestManaged_SetByteSlice(t *testing.T) {
	prefix := "ABCD"
	slice := "EFGHIJKLMN"
	suffix := "OPR"
	data := prefix + slice + suffix
	_, err := test.BuildMockInstanceCallTest(t).
		WithContracts(
			test.CreateMockContract(test.ParentAddress).
				WithBalance(1000).
				WithMethods(func(parentInstance *mock.InstanceMock, config interface{}) {
					parentInstance.AddMockMethod("testFunction", func() *mock.InstanceMock {
						host := parentInstance.Host
						managedType := host.ManagedTypes()
						mBuffer := managedType.NewManagedBufferFromBytes(
							make([]byte, len(data)))
						result := elrondapi.ManagedBufferSetByteSliceWithTypedArgs(
							host, mBuffer, int32(len(prefix)), int32(len(slice)), []byte(data))
						if result != 0 {
							elrondapi.WithFaultAndHost(host, arwen.ErrSignalError, true)
						}
						bufferBytes, err := managedType.GetBytes(mBuffer)
						if err != nil {
							elrondapi.WithFaultAndHost(host, err, true)
						}
						host.Output().Finish(bufferBytes)
						return parentInstance
					})
				}),
		).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(1000).
			WithFunction("testFunction").
			Build()).
		AndAssertResults(func(world *worldmock.MockWorld, verify *test.VMOutputVerifier) {
			verify.Ok().
				// |....ABCDEFGHIJ...|
				ReturnData(append(make([]byte, len(prefix)),
					append([]byte(data[0:len(slice)]), make([]byte, len(suffix))...)...))
		})
	assert.Nil(t, err)
}

// managed buffer ~~~> managed buffer
func TestManaged_CopyByteSlice_DifferentBuffer(t *testing.T) {
	prefix := "ABCD"
	slice := "EFGHIJKLMN"
	suffix := "OPR"
	sourceData := prefix + slice + suffix
	destinationData := "01234567890123456789"
	_, err := test.BuildMockInstanceCallTest(t).
		WithContracts(
			test.CreateMockContract(test.ParentAddress).
				WithBalance(1000).
				WithMethods(func(parentInstance *mock.InstanceMock, config interface{}) {
					parentInstance.AddMockMethod("testFunction", func() *mock.InstanceMock {
						host := parentInstance.Host
						managedType := host.ManagedTypes()
						sourceMBuffer := managedType.NewManagedBufferFromBytes(
							[]byte(sourceData))
						destMBuffer := managedType.NewManagedBufferFromBytes(
							[]byte(destinationData))
						result := elrondapi.ManagedBufferCopyByteSliceWithHost(
							host, sourceMBuffer, int32(len(prefix)), int32(len(slice)), destMBuffer)
						if result != 0 {
							elrondapi.WithFaultAndHost(host, arwen.ErrSignalError, true)
						}
						destBytes, err := managedType.GetBytes(destMBuffer)
						if err != nil {
							elrondapi.WithFaultAndHost(host, err, true)
						}
						host.Output().Finish(destBytes)
						return parentInstance
					})
				}),
		).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(1000).
			WithFunction("testFunction").
			Build()).
		AndAssertResults(func(world *worldmock.MockWorld, verify *test.VMOutputVerifier) {
			verify.Ok().
				ReturnData([]byte(slice))
		})
	assert.Nil(t, err)
}

func TestManaged_CopyByteSlice_SameBuffer(t *testing.T) {
	prefix := "ABCD"
	slice := "EFGHIJKLMN"
	suffix := "OPR"
	sourceData := prefix + slice + suffix
	deltaForSlice := int32(2)
	_, err := test.BuildMockInstanceCallTest(t).
		WithContracts(
			test.CreateMockContract(test.ParentAddress).
				WithBalance(1000).
				WithMethods(func(parentInstance *mock.InstanceMock, config interface{}) {
					parentInstance.AddMockMethod("testFunction", func() *mock.InstanceMock {
						host := parentInstance.Host
						managedType := host.ManagedTypes()
						sourceMBuffer := managedType.NewManagedBufferFromBytes(
							[]byte(sourceData))
						result := elrondapi.ManagedBufferCopyByteSliceWithHost(
							host, sourceMBuffer, int32(len(prefix))-deltaForSlice, int32(len(slice)), sourceMBuffer)
						if result != 0 {
							elrondapi.WithFaultAndHost(host, arwen.ErrSignalError, true)
						}
						destBytes, err := managedType.GetBytes(sourceMBuffer)
						if err != nil {
							elrondapi.WithFaultAndHost(host, err, true)
						}
						host.Output().Finish(destBytes)
						return parentInstance
					})
				}),
		).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(1000).
			WithFunction("testFunction").
			Build()).
		AndAssertResults(func(world *worldmock.MockWorld, verify *test.VMOutputVerifier) {
			prefixLen := int32(len(prefix))
			sliceLen := int32(len(slice))
			verify.Ok().
				// |CDEFGHIJKL|
				ReturnData(
					append([]byte(prefix)[prefixLen-deltaForSlice:prefixLen],
						[]byte(slice)[:sliceLen-deltaForSlice]...))
		})
	assert.Nil(t, err)
}
