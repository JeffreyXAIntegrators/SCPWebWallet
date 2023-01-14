package daemon

import (
	"fmt"
	"path/filepath"
	"time"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/consensus"
	"gitlab.com/scpcorp/ScPrime/modules/gateway"
	"gitlab.com/scpcorp/ScPrime/modules/transactionpool"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/node"

	"gitlab.com/scpcorp/webwallet/modules/bootstrapper"
	"gitlab.com/scpcorp/webwallet/modules/browserconfig"
	"gitlab.com/scpcorp/webwallet/modules/consensesbuilder"
	"gitlab.com/scpcorp/webwallet/server"
	wwConfig "gitlab.com/scpcorp/webwallet/utils/config"
)

func loadNode(node *node.Node, config *wwConfig.WebWalletConfig) error {
	fmt.Println("Loading modules:")
	// Make sure the path is an absolute one.
	dir, err := filepath.Abs(config.Dir)
	if err != nil {
		return err
	}
	node.Dir = dir
	// Configure Browser
	needsShutdown, err := initializeBrowser(config)
	if err != nil {
		return err
	} else if needsShutdown {
		return nil
	}
	// Bootstrap Consensus Set if necessary
	bootstrapConsensusSet(config)
	// Attach Node To Server
	server.AttachNode(node)
	// Load Gateway.
	err = loadGateway(config, node)
	if err != nil {
		return err
	}
	// Load Consensus Set
	err = loadConsensusSet(config, node)
	if err != nil {
		return err
	}
	// Build Consensus Set if necessary
	buildConsensusSet(config)
	// Load Transaction Pool
	err = loadTransactionPool(config, node)
	if err != nil {
		return err
	}
	return nil
}

func closeNode(node *node.Node, config *wwConfig.WebWalletConfig) error {
	fmt.Println("Closing modules:")
	config.CreateWallet = false
	config.CreateTransactionPool = false
	consensusbuilder.Close()
	config.CreateConsensusSet = false
	config.CreateGateway = false
	err := node.Close()
	bootstrapper.Close()
	browserconfig.Close()
	return err
}

func initializeBrowser(config *wwConfig.WebWalletConfig) (bool, error) {
	loadStart := time.Now()
	fmt.Printf("Initializing browser...")
	time.Sleep(1 * time.Millisecond)
	browserconfig.Start(config.Dir)
	loadTime := time.Since(loadStart).Seconds()
	if browserconfig.Status() == browserconfig.Closed {
		fmt.Println(" closed after", loadTime, "seconds.")
		return true, nil
	}
	if browserconfig.Status() == browserconfig.Failed {
		fmt.Println(" failed after", loadTime, "seconds.")
		return true, nil
	}
	browser, err := browserconfig.Browser(config.Dir)
	if err != nil {
		fmt.Println(" failed after", loadTime, "seconds.")
		return true, err
	}
	if browserconfig.Status() == browserconfig.Initialized {
		fmt.Printf(" browser initialized to %s in %v seconds.\n", browser, loadTime)
		return true, nil
	}
	fmt.Printf(" browser set to %s in %v seconds.\n", browser, loadTime)
	return false, nil
}

func bootstrapConsensusSet(config *wwConfig.WebWalletConfig) {
	loadStart := time.Now()
	fmt.Printf("Bootstrapping consensus...")
	time.Sleep(1 * time.Millisecond)
	bootstrapper.Start(config.Dir)
	loadTime := time.Since(loadStart).Seconds()
	if bootstrapper.Progress() == bootstrapper.Skipped {
		fmt.Println(" skipped after", loadTime, "seconds.")
	} else if bootstrapper.Progress() == bootstrapper.Closed {
		fmt.Println(" closed after", loadTime, "seconds.")
	} else {
		fmt.Println(" done in", loadTime, "seconds.")
	}
}

func loadGateway(config *wwConfig.WebWalletConfig, node *node.Node) error {
	loadStart := time.Now()
	if !config.CreateGateway {
		return nil
	}
	rpcAddress := "localhost:0"
	gatewayDeps := modules.ProdDependencies
	fmt.Printf("Loading gateway...")
	dir := node.Dir
	g, err := gateway.NewCustomGateway(rpcAddress, config.Bootstrap, filepath.Join(dir, modules.GatewayDir), gatewayDeps)
	if err != nil {
		return err
	}
	if g != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.Gateway = g
	return nil
}

func loadConsensusSet(config *wwConfig.WebWalletConfig, node *node.Node) error {
	loadStart := time.Now()
	c := make(chan error, 1)
	defer close(c)
	if !config.CreateConsensusSet {
		return nil
	}
	fmt.Printf("Loading consensus set...")
	consensusSetDeps := modules.ProdDependencies
	g := node.Gateway
	dir := node.Dir
	cs, errChanCS := consensus.NewCustomConsensusSet(g, config.Bootstrap, filepath.Join(dir, modules.ConsensusDir), consensusSetDeps)
	if err := modules.PeekErr(errChanCS); err != nil {
		return err
	}
	if cs != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.ConsensusSet = cs
	return nil
}

func buildConsensusSet(config *wwConfig.WebWalletConfig) {
	loadStart := time.Now()
	fmt.Printf("Building consensus set...")
	time.Sleep(1 * time.Millisecond)
	consensusbuilder.Start(config.Dir)
	loadTime := time.Since(loadStart).Seconds()
	if consensusbuilder.Progress() == consensusbuilder.Closed {
		fmt.Println(" closed after", loadTime, "seconds.")
	} else {
		fmt.Println(" done in", loadTime, "seconds.")
	}
}

func loadTransactionPool(config *wwConfig.WebWalletConfig, node *node.Node) error {
	loadStart := time.Now()
	if !config.CreateTransactionPool {
		return nil
	}
	fmt.Printf("Loading transaction pool...")
	tpoolDeps := modules.ProdDependencies
	cs := node.ConsensusSet
	g := node.Gateway
	dir := node.Dir
	tp, err := transactionpool.NewCustomTPool(cs, g, filepath.Join(dir, modules.TransactionPoolDir), tpoolDeps)
	if err != nil {
		return err
	}
	if tp != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.TransactionPool = tp
	return nil
}

// LoadWallet loads the wallet module
func LoadWallet(config *wwConfig.WebWalletConfig, node *node.Node, walletDirName string) error {
	loadStart := time.Now()
	if !config.CreateWallet {
		return nil
	}
	walletDeps := modules.ProdDependencies
	fmt.Printf("Loading wallet...")
	cs := node.ConsensusSet
	tp := node.TransactionPool
	dir := node.Dir
	w, err := wallet.NewCustomWallet(cs, tp, filepath.Join(dir, "wallets", walletDirName), walletDeps)
	if err != nil {
		return err
	}
	if w != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.Wallet = w
	return nil
}
