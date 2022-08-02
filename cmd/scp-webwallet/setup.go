package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/zenity"
)

func setupBrowser() (bool, error) {
	browserConfig := filepath.Join(webWalletConfig.Dir, "browser", "browser.txt")
	if exists(browserConfig) {
		return true, nil
	}
	browser, err := zenity.List(
		"Select desired browser from the list below:",
		[]string{"Default", "Chrome", "Edge"},
		zenity.Title("ScPrime Web Wallet Setup"),
		zenity.DefaultItems("Default"),
		zenity.DisallowEmpty(),
	)
	if err != nil {
		return false, nil
	}
	err = os.WriteFile(browserConfig, []byte(strings.ToLower(browser)), 0600)
	return err == nil, err
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
