package browserconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BrowserConfigDir defined the directory that the browser config is stored in
const BrowserConfigDir = "browser"

// Closed is the value that the browser config is set to after it has been closed
const Closed = "Closed"

// Failed is the value that the browser config is set to if configuration fails
const Failed = "Failed"

// Waiting is the value that the browser config is set to when it is waiting for input
const Waiting = "Waiting"

// Done is the value that the browser config is set to after it has been loaded
const Done = "Done"

// Initialized is the value that the browser config is set to after it has been set
const Initialized = "Initialized"

var status = ""

// Close the consensus builder module
func Close() {
	fmt.Println("Closing browser config...")
	status = Closed
}

// Initialize the browser config module
func Initialize() {
	if status == "" {
		status = Waiting
	}
}

// Configure the browser config module
func Configure(dataDir string, browser string) error {
	browserConfigDir := filepath.Join(dataDir, BrowserConfigDir)
	browserConfig := filepath.Join(browserConfigDir, BrowserConfigDir+".txt")
	_, err := os.Stat(browserConfigDir)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(browserConfigDir, os.ModePerm)
	}
	if err != nil {
		status = Failed
		return err
	}
	err = os.WriteFile(browserConfig, []byte(browser), 0600)
	if err != nil {
		status = Failed
		return err
	}
	if browser == "default" {
		status = Done
	} else {
		status = Initialized
	}
	return nil
}

// Browser returns the configured browser
func Browser(dataDir string) (string, error) {
	browserConfig := filepath.Join(dataDir, BrowserConfigDir, BrowserConfigDir+".txt")
	browser, err := os.ReadFile(browserConfig)
	if err != nil {
		return "", err
	}
	return string(browser), nil
}

// Status returns the status
func Status() string {
	return status
}

// Start begins the process of buildinng the consensus set from peers.
func Start(dataDir string) {
	browserConfig := filepath.Join(dataDir, BrowserConfigDir, BrowserConfigDir+".txt")
	if exists(browserConfig) {
		status = Done
		return
	}
	status = Waiting
	for status == Waiting {
		time.Sleep(25 * time.Millisecond)
	}
	if status == Waiting {
		status = Done
	}
	return
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
