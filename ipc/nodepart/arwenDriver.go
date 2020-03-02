package nodepart

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/arwen-wasm-vm/ipc/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ vmcommon.VMExecutionHandler = (*ArwenDriver)(nil)

// ArwenDriver is
type ArwenDriver struct {
	blockchainHook vmcommon.BlockchainHook
	cryptoHook     vmcommon.CryptoHook
	vmType         []byte
	blockGasLimit  uint64
	gasSchedule    map[string]map[string]uint64

	arwenInputRead   *os.File
	arwenInputWrite  *os.File
	arwenOutputRead  *os.File
	arwenOutputWrite *os.File
	command          *exec.Cmd
	part             *NodePart
}

// NewArwenDriver creates
func NewArwenDriver(
	blockchainHook vmcommon.BlockchainHook,
	cryptoHook vmcommon.CryptoHook,
	vmType []byte,
	blockGasLimit uint64,
	gasSchedule map[string]map[string]uint64,
) (*ArwenDriver, error) {
	driver := &ArwenDriver{
		blockchainHook: blockchainHook,
		cryptoHook:     cryptoHook,
		vmType:         vmType,
		blockGasLimit:  blockGasLimit,
		gasSchedule:    gasSchedule,
	}

	err := driver.startArwen()
	return driver, err
}

func (driver *ArwenDriver) startArwen() error {
	err := driver.startArwenWithPipes()
	return err
}

func (driver *ArwenDriver) startArwenWithFiles() error {
	user, _ := user.Current()
	home := user.HomeDir
	folder := path.Join(home, "Arwen")

	nodeToArwen := filepath.Join(folder, fmt.Sprintf("node-to-arwen.bin"))
	arwenToNode := filepath.Join(folder, fmt.Sprintf("arwen-to-node.bin"))

	// Open the files as required
	nodeToArwenFile, err := os.OpenFile(nodeToArwen, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	arwenToNodeFile, err := os.Open(arwenToNode)
	if err != nil {
		return err
	}

	driver.part, err = NewNodePart(arwenToNodeFile, nodeToArwenFile, driver.blockchainHook, driver.cryptoHook)
	if err != nil {
		return err
	}

	return nil
}

func (driver *ArwenDriver) startArwenWithPipes() error {
	driver.resetPipeStreams()

	user, _ := user.Current()
	home := user.HomeDir
	executable := path.Join(home, "Arwen", "arwen")

	driver.command = exec.Command(executable)
	driver.command.Stdout = os.Stdout
	driver.command.Stderr = os.Stderr
	// TODO: pass vmType, blockGasLimit and gasSchedule when starting Arwen

	driver.command.ExtraFiles = []*os.File{driver.arwenInputRead, driver.arwenOutputWrite}
	err := driver.command.Start()
	if err != nil {
		return err
	}

	driver.part, err = NewNodePart(driver.arwenOutputRead, driver.arwenInputWrite, driver.blockchainHook, driver.cryptoHook)
	if err != nil {
		return err
	}

	return nil
}

func (driver *ArwenDriver) resetPipeStreams() error {
	closeFile(driver.arwenInputRead)
	closeFile(driver.arwenInputWrite)
	closeFile(driver.arwenOutputRead)
	closeFile(driver.arwenOutputWrite)

	var err error

	driver.arwenInputRead, driver.arwenInputWrite, err = os.Pipe()
	if err != nil {
		return err
	}

	driver.arwenOutputRead, driver.arwenOutputWrite, err = os.Pipe()
	if err != nil {
		return err
	}

	return nil
}

func closeFile(file *os.File) {
	if file != nil {
		err := file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot close file.\n")
		}
	}
}

func (driver *ArwenDriver) forceRestartArwen() error {
	if !driver.command.ProcessState.Exited() {
		err := driver.command.Process.Kill()
		if err != nil {
			return err
		}
	}

	err := driver.startArwen()
	return err
}

func (driver *ArwenDriver) restartArwenIfNecessary() error {
	if !driver.command.ProcessState.Exited() {
		return nil
	}

	err := driver.startArwen()
	return err
}

// RunSmartContractCreate creates
func (driver *ArwenDriver) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	// TODO: restart if necessary

	request := &common.ContractRequest{
		Action:      "Deploy",
		CreateInput: input,
	}

	response, err := driver.part.StartLoop(request)
	if err != nil {
		// TODO: if critical error, restart.
	}

	return response.VMOutput, err
}

// RunSmartContractCall calls
func (driver *ArwenDriver) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	// TODO: restart if necessary

	request := &common.ContractRequest{
		Action:    "Call",
		CallInput: input,
	}

	response, err := driver.part.StartLoop(request)
	if err != nil {
		// TODO: if critical error, restart.
	}

	return response.VMOutput, err
}

// func OnRoundEnded -> triggers Arwen restart.

// Close stops Arwen
func (driver *ArwenDriver) Close() error {
	err := driver.command.Process.Kill()
	if err != nil {
		return err
	}

	return nil
}
