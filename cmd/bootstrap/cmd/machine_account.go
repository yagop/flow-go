package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-go/crypto"
	model "github.com/onflow/flow-go/model/bootstrap"
	ioutils "github.com/onflow/flow-go/utils/io"
)

var (
	flagMachineAccountAddress string
)

// machineAccountCmd represents the `machine-account` command which generates required machine account file
// for existing and new operators. New operators would have run the `bootstrap keys` cmd get all three keys
// before running this command.
var machineAccountCmd = &cobra.Command{
	Use:   "machine-account",
	Short: "",
	Run:   machineAccountRun,
}

func init() {
	rootCmd.AddCommand(machineAccountCmd)

	machineAccountCmd.Flags().StringVar(&flagMachineAccountAddress, "address", "", "the node's machine account address")
	_ = machineAccountCmd.MarkFlagRequired("address")
}

// keyCmdRun generate the node staking key, networking key and node information
func machineAccountRun(_ *cobra.Command, _ []string) {

	// read nodeID written to boostrap dir by `bootstrap key`
	nodeID, err := readNodeID()
	if err != nil {
		log.Fatal().Err(err).Msg("could not read node id")
	}

	// check if node-machine-account-key.priv.json path exists
	machineAccountKeyPath := fmt.Sprintf(model.PathNodeMachineAccountPrivateKey, nodeID)
	keyExists, err := pathExists(filepath.Join(flagOutdir, machineAccountKeyPath))
	if err != nil {
		log.Fatal().Err(err).Msg("could not check if node-machine-account-key.priv.json exists")
	}
	if !keyExists {
		log.Fatal().Msg("could not read machine account private key file: run `bootstrap machine-account-key` to create one")
	}

	// check if node-machine-account-info.priv.json file exists in boostrap dir
	machineAccountInfoPath := fmt.Sprintf(model.PathNodeMachineAccountInfoPriv, nodeID)
	infoExists, err := pathExists(filepath.Join(flagOutdir, machineAccountInfoPath))
	if err != nil {
		log.Fatal().Err(err).Msg("could not check if node-machine-account-info.priv.json exists")
	}
	if infoExists {
		log.Info().Str("path", machineAccountInfoPath).Msg("node maching account info file already exists")
		return
	}

	// read in machine account private key
	machinePrivKey := readMachineAccountPriv(nodeID)
	log.Info().Msg("read machine account private key json")

	// create node-machine-account-info.priv.json file
	machineAccountInfo := assembleNodeMachineAccountInfo(flagMachineAccountAddress, 0, machinePrivKey)

	// write machine account info
	writeJSON(fmt.Sprintf(model.PathNodeMachineAccountInfoPriv, nodeID), machineAccountInfo)
}

// readMachineAccountPriv reads the machine account private key files in the bootstrap dir
func readMachineAccountPriv(nodeID string) crypto.PrivateKey {
	var machineAccountPriv model.NodeMachineAccountPriv

	path := filepath.Join(flagOutdir, fmt.Sprintf(model.PathNodeMachineAccountPrivateKey, nodeID))
	readJSON(path, &machineAccountPriv)

	return machineAccountPriv.PrivateKey.PrivateKey
}

// readNodeID reads the NodeID file written by `bootstrap key` command
func readNodeID() (string, error) {
	path := filepath.Join(flagOutdir, model.PathNodeID)

	data, err := ioutils.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %w", path, err)
	}

	return strings.TrimSpace(string(data)), nil
}
