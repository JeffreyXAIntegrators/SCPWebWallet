package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ncruces/zenity"
	"gitlab.com/scpcorp/webwallet/build"
	"gitlab.com/scpcorp/webwallet/daemon"
	"gitlab.com/scpcorp/webwallet/utils/config"
)

// exit codes
// inspired by sysexits.h
const (
	exitCodeGeneral = 1  // Not in sysexits.h, but is standard practice.
	exitCodeUsage   = 64 // EX_USAGE in sysexits.h
)

var (
	// configure the the web wallet
	webWalletConfig = config.WebWalletConfig{
		CreateGateway:                 true,
		CreateConsensusSet:            true,
		CreateTransactionPool:         true,
		CreateWallet:                  true,
		Bootstrap:                     true, // set to true when the gateway should use the bootstrap peer list
		Headless:                      false,
		Port:                          4300,
		Dir:                           build.ScPrimeWebWalletDir(),
		CheckTokenExpirationFrequency: 1 * time.Hour, // default
	}
)

// die prints its arguments to stderr, then exits the program with the default
// error code.
func die(err error) {
	fmt.Println(err)
	zenity.Error(fmt.Sprint(err),
		zenity.Title("Error"),
		zenity.ErrorIcon)
	os.Exit(exitCodeGeneral)
}

// main starts the daemon.
func main() {
	// Configure Browser settings.
	done, err := setupBrowser()
	if err != nil {
		die(err)
		return
	}
	if !done {
		return
	}
	// Start the ScPrime web wallet daemon.
	// the startDaemon method will only return when it is shutting down.
	err = daemon.StartDaemon(&webWalletConfig)
	if err != nil {
		die(err)
	}
	// Daemon seems to have closed cleanly. Print a 'closed' message.
	fmt.Println("Shutdown complete.")
}
