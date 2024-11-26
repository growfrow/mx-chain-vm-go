package vmjsonintegrationtest

import (
	"testing"
)

const expectedAdderLog = `starting log:
GetFunctionNames: [add callBack getSum init upgrade]
ValidateFunctionArities: true
GetFunctionNames: [add callBack getSum init upgrade]
HasFunction(init): true
CallFunction(init):
VM hook begin: CheckNoPayment()
GetPointsUsed: 3
GetPointsUsed: 3
SetPointsUsed: 103
VM hook end:   CheckNoPayment()
VM hook begin: GetNumArguments()
GetPointsUsed: 110
GetPointsUsed: 110
SetPointsUsed: 210
VM hook end:   GetNumArguments()
VM hook begin: BigIntGetUnsignedArgument(0, -201)
GetPointsUsed: 249
GetPointsUsed: 249
SetPointsUsed: 1249
VM hook end:   BigIntGetUnsignedArgument(0, -201)
VM hook begin: MBufferSetBytes(-202, 131097, 3)
GetPointsUsed: 1289
GetPointsUsed: 1289
SetPointsUsed: 3289
GetPointsUsed: 3289
GetPointsUsed: 3289
SetPointsUsed: 6289
VM hook end:   MBufferSetBytes(-202, 131097, 3)
VM hook begin: MBufferFromBigIntUnsigned(-203, -201)
GetPointsUsed: 6333
GetPointsUsed: 6333
SetPointsUsed: 10333
VM hook end:   MBufferFromBigIntUnsigned(-203, -201)
VM hook begin: MBufferStorageStore(-202, -203)
GetPointsUsed: 10345
GetPointsUsed: 10345
SetPointsUsed: 85345
GetPointsUsed: 85345
GetPointsUsed: 85345
SetPointsUsed: 85345
GetPointsUsed: 85345
GetPointsUsed: 85345
SetPointsUsed: 135345
VM hook end:   MBufferStorageStore(-202, -203)
GetPointsUsed: 135352
GetPointsUsed: 135352
GetPointsUsed: 135352
GetPointsUsed: 135352
GetPointsUsed: 135352
GetPointsUsed: 135352
GetPointsUsed: 135352
Reset: true
SetPointsUsed: 0
SetGasLimit: 9223372036853567007
SetBreakpointValue: 0
HasFunction(getSum): true
CallFunction(getSum):
VM hook begin: CheckNoPayment()
GetPointsUsed: 3
GetPointsUsed: 3
SetPointsUsed: 103
VM hook end:   CheckNoPayment()
VM hook begin: GetNumArguments()
GetPointsUsed: 110
GetPointsUsed: 110
SetPointsUsed: 210
VM hook end:   GetNumArguments()
VM hook begin: MBufferSetBytes(-201, 131097, 3)
GetPointsUsed: 250
GetPointsUsed: 250
SetPointsUsed: 2250
GetPointsUsed: 2250
GetPointsUsed: 2250
SetPointsUsed: 5250
VM hook end:   MBufferSetBytes(-201, 131097, 3)
VM hook begin: MBufferStorageLoad(-201, -202)
GetPointsUsed: 5291
GetPointsUsed: 5291
SetPointsUsed: 6291
GetPointsUsed: 6291
GetPointsUsed: 6291
SetPointsUsed: 21578
VM hook end:   MBufferStorageLoad(-201, -202)
VM hook begin: MBufferToBigIntUnsigned(-202, -203)
GetPointsUsed: 21611
GetPointsUsed: 21611
SetPointsUsed: 25611
VM hook end:   MBufferToBigIntUnsigned(-202, -203)
VM hook begin: BigIntFinishUnsigned(-203)
GetPointsUsed: 25622
GetPointsUsed: 25622
SetPointsUsed: 26622
GetPointsUsed: 26622
GetPointsUsed: 26622
SetPointsUsed: 36622
VM hook end:   BigIntFinishUnsigned(-203)
GetPointsUsed: 36624
GetPointsUsed: 36624
GetPointsUsed: 36624
GetPointsUsed: 36624
GetPointsUsed: 36624
GetPointsUsed: 36624
GetPointsUsed: 36624
GetPointsUsed: 36624
Reset: true
SetPointsUsed: 0
SetGasLimit: 3791200
SetBreakpointValue: 0
HasFunction(add): true
CallFunction(add):
VM hook begin: CheckNoPayment()
GetPointsUsed: 3
GetPointsUsed: 3
SetPointsUsed: 103
VM hook end:   CheckNoPayment()
VM hook begin: GetNumArguments()
GetPointsUsed: 110
GetPointsUsed: 110
SetPointsUsed: 210
VM hook end:   GetNumArguments()
VM hook begin: BigIntGetUnsignedArgument(0, -201)
GetPointsUsed: 249
GetPointsUsed: 249
SetPointsUsed: 1249
VM hook end:   BigIntGetUnsignedArgument(0, -201)
VM hook begin: MBufferSetBytes(-202, 131097, 3)
GetPointsUsed: 1289
GetPointsUsed: 1289
SetPointsUsed: 3289
GetPointsUsed: 3289
GetPointsUsed: 3289
SetPointsUsed: 6289
VM hook end:   MBufferSetBytes(-202, 131097, 3)
VM hook begin: MBufferStorageLoad(-202, -203)
GetPointsUsed: 6333
GetPointsUsed: 6333
SetPointsUsed: 7333
GetPointsUsed: 7333
GetPointsUsed: 7333
SetPointsUsed: 22620
VM hook end:   MBufferStorageLoad(-202, -203)
VM hook begin: MBufferToBigIntUnsigned(-203, -204)
GetPointsUsed: 22653
GetPointsUsed: 22653
SetPointsUsed: 26653
VM hook end:   MBufferToBigIntUnsigned(-203, -204)
VM hook begin: BigIntAdd(-204, -204, -201)
GetPointsUsed: 26673
GetPointsUsed: 26673
SetPointsUsed: 28673
VM hook end:   BigIntAdd(-204, -204, -201)
VM hook begin: MBufferFromBigIntUnsigned(-205, -204)
GetPointsUsed: 28712
GetPointsUsed: 28712
SetPointsUsed: 32712
VM hook end:   MBufferFromBigIntUnsigned(-205, -204)
VM hook begin: MBufferStorageStore(-202, -205)
GetPointsUsed: 32724
GetPointsUsed: 32724
SetPointsUsed: 107724
GetPointsUsed: 107724
GetPointsUsed: 107724
SetPointsUsed: 107724
GetPointsUsed: 107724
GetPointsUsed: 107724
SetPointsUsed: 107724
VM hook end:   MBufferStorageStore(-202, -205)
GetPointsUsed: 107731
GetPointsUsed: 107731
GetPointsUsed: 107731
GetPointsUsed: 107731
GetPointsUsed: 107731
GetPointsUsed: 107731
GetPointsUsed: 107731
GetPointsUsed: 107731
Clean: true
`

func TestRustAdderLog(t *testing.T) {
	ScenariosTest(t).
		Folder("adder/scenarios/adder.scen.json").
		WithExecutorLogs().
		Run().
		CheckNoError().
		CheckLog(expectedAdderLog)
}
