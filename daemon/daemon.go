package daemon

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/ncruces/zenity"

	"gitlab.com/scpcorp/webwallet/build"
	"gitlab.com/scpcorp/webwallet/modules/browserconfig"
	"gitlab.com/scpcorp/webwallet/modules/launcher"
	"gitlab.com/scpcorp/webwallet/server"
	wwConfig "gitlab.com/scpcorp/webwallet/utils/config"

	spdBuild "gitlab.com/scpcorp/ScPrime/build"
	"gitlab.com/scpcorp/ScPrime/node"
)

// printVersionAndRevision prints the daemon's version and revision numbers.
func printVersionAndRevision() {
	if build.Version == "" {
		fmt.Println("WARN: compiled ScPrime web wallet without version.")
	} else {
		fmt.Println("ScPrime web wallet v" + build.Version)
	}
	if build.GitRevision == "" {
		fmt.Println("WARN: compiled ScPrime web wallet without version.")
	} else {
		fmt.Println("ScPrime web wallet Git revision " + build.GitRevision)
	}
	if spdBuild.DEBUG {
		fmt.Println("Running ScPrime daemon with debugging enabled")
	}
	if spdBuild.Version == "" {
		fmt.Println("WARN: compiled ScPrime daemon without version.")
	} else {
		fmt.Println("ScPrime daemon v" + spdBuild.Version)
	}
}

// installMmapSignalHandler installs a signal handler for Mmap related signals
// and exits when such a signal is received.
func installMmapSignalHandler() {
	// NOTE: ideally we would catch SIGSEGV here too, since that signal can
	// also be thrown by an mmap I/O error. However, SIGSEGV can occur under
	// other circumstances as well, and in those cases, we will want a full
	// stack trace.
	mmapChan := make(chan os.Signal, 1)
	signal.Notify(mmapChan, syscall.SIGBUS)
	go func() {
		<-mmapChan
		fmt.Println("A fatal I/O exception (SIGBUS) has occurred.")
		fmt.Println("Please check your disk for errors.")
		os.Exit(1)
	}()
}

// installKillSignalHandler installs a signal handler for os.Interrupt, os.Kill
// and syscall.SIGTERM and returns a channel that is closed when one of them is
// caught.
func installKillSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	return sigChan
}

func startNode(node *node.Node, config *wwConfig.WebWalletConfig, loadStart time.Time) {
	err := loadNode(node, config)
	if err != nil {
		fmt.Println("Server is unable to create the ScPrime node.")
		fmt.Println(err)
		return
	}
	// Print a 'startup complete' message.
	startupTime := time.Since(loadStart)
	fmt.Printf("Finished full startup in %.3f seconds\n", startupTime.Seconds())
	return
}

func launchGui(config *wwConfig.WebWalletConfig) chan struct{} {
	dir, err := filepath.Abs(config.Dir)
	if err != nil {
		fmt.Printf("unable to launch GUI: %v\n", err)
		return nil
	}
	browser, _ := browserconfig.Browser(dir)
	return launcher.Launch(browser)
}

func isPortAvailabile(config *wwConfig.WebWalletConfig) (bool, error) {
	port := config.Port
	url := "http://" + net.JoinHostPort("localhost", strconv.Itoa(port))
	client := http.Client{Timeout: time.Duration(50) * time.Millisecond}
	optReq, err := http.NewRequest("OPTIONS", url, nil)
	if err != nil {
		return false, err
	}
	res, err := client.Do(optReq)
	if err == nil {
		res.Body.Close()
		for _, app := range res.Header.Values("Application") {
			return false, fmt.Errorf("port %d already in use by %s", port, app)
		}
		return false, fmt.Errorf("port %d already in use", port)
	}
	res, err = http.Get(url)
	if err == nil {
		res.Body.Close()
		for _, app := range res.Header.Values("Application") {
			return false, fmt.Errorf("port %d already in use by %s", port, app)
		}
		return false, fmt.Errorf("port %d already in use", port)
	}
	return true, nil
}

// StartDaemon uses the config parameters to initialize modules and start the web wallet.
func StartDaemon(config *wwConfig.WebWalletConfig) error {

	// record startup time
	loadStart := time.Now()

	// listen for kill signals
	sigChan := installKillSignalHandler()

	// print the Version and GitRevision
	printVersionAndRevision()

	// install a signal handler that will catch exceptions thrown by mmap'd files
	installMmapSignalHandler()

	// verify port is open
	resp, err := isPortAvailabile(config)
	if err != nil || !resp {
		fmt.Printf("Port %d is not available, quitting...\n", config.Port)
		return err
	}

	// start server
	fmt.Println("Starting ScPrime Web Wallet server...")
	server.StartHTTPServer(config)

	// start a node
	node := &node.Node{}
	if server.IsRunning() {
		go startNode(node, config, loadStart)
	}

	if server.IsRunning() {
		// block until node is started or 500 milliseconds has passed.
		for i := 0; i < 100; i++ {
			if node.TransactionPool == nil {
				time.Sleep(5 * time.Millisecond)
			}
		}
	}

	// launch GUI
	uiDone := make(chan struct{})
	if !config.Headless {
		uiDone = launchGui(config)
	}

	if !server.IsRunning() {
		fmt.Println("Unable to start server, quitting...")
		return nil
	}

	if config.Headless {
		fmt.Printf("SCP Web Wallet is running at http://localhost:%d\n", config.Port)
	}

	select {
	case <-server.Wait():
		fmt.Println("Server was stopped, quitting...")
	case <-sigChan:
		fmt.Println("\rCaught stop signal, quitting...")
	case <-uiDone:
		fmt.Println("Browser was closed, quitting...")
	}

	// Close
	var shutdownGui zenity.ProgressDialog
	if !config.Headless {
		title := "Shutting Down ScPrime Web Wallet"
		shutdownGui, _ = zenity.Progress(zenity.Title(title), zenity.Pulsate())
		shutdownGui.Text("Closing node...")
	}
	server.CloseAllWallets()
	if node != nil {
		closeNode(node, config)
	}
	if shutdownGui != nil {
		shutdownGui.Complete()
		shutdownGui.Close()
	}
	return nil
}
