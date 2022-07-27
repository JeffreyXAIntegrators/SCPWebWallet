package consensusbuilder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/consensus"

	"gitlab.com/scpcorp/webwallet/build"
)

// Closed is the value that the consensus builder's progress is set to after it has been closed.
const Closed = "Closed"

// LocalConsensusSize is the size in bytes of the consensus file that is stored to disk.
var LocalConsensusSize = int64(0)

var status = ""

// Close the consensus builder module
func Close() {
	status = Closed
}

// Initialize building the consensus set from peers.
func Initialize() {
	if status == "" {
		status = "0"
	}
}

// Progress returns the consensus builder's progress as a percentage.
func Progress() string {
	_, err := strconv.Atoi(status)
	if err != nil {
		return status
	}
	return status + `%`
}

// Start begins the process of buildinng the consensus set from peers.
func Start(dataDir string) {
	if status == Closed {
		return
	}
	consensusDir := filepath.Join(dataDir, modules.ConsensusDir)
	consensusDb := filepath.Join(consensusDir, consensus.DatabaseFilename)
	fi, err := os.Stat(consensusDb)
	if !errors.Is(err, os.ErrNotExist) {
		LocalConsensusSize = fi.Size()
		if LocalConsensusSize > build.ConsensusSizeByteCheck() {
			// There is no need to block loading the wallet because the on-disk
			// consensus size is larger than the consensus size byte check.
			status = `100`
			return
		}
	}
	status = `0`
	// Updates the status.
	for updateStatus(consensusDb, build.ConsensusSizeByteCheck()) < 100 {
		for i := 0; i < 40; i++ {
			time.Sleep(25 * time.Millisecond)
			if status == Closed {
				return
			}
		}
	}
	status = `100`
}

// updates the status.
func updateStatus(filepath string, size int64) int {
	if status == Closed {
		return 0
	}
	_, err := strconv.Atoi(status)
	if err != nil {
		fmt.Printf("%v", err)
		return 0
	}
	fi, err := os.Stat(filepath)
	if err != nil {
		status = `0`
	}
	progress := int(float64(fi.Size()) / float64(size) * float64(100))
	if progress > 0 && progress < 99 {
		status = fmt.Sprintf("%d", progress)
	}
	return progress
}
