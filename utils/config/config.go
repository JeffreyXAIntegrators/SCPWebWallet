package config

import (
	"time"
)

// WebWalletConfig holds the config values needed by the web wallet.
type WebWalletConfig struct {
	CreateGateway                 bool
	CreateConsensusSet            bool
	CreateTransactionPool         bool
	CreateWallet                  bool
	Bootstrap                     bool
	Headless                      bool
	Port                          int
	Dir                           string
	CheckTokenExpirationFrequency time.Duration
}
