package daemon

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ncruces/zenity"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/consensus"
	"gitlab.com/scpcorp/ScPrime/modules/gateway"
	"gitlab.com/scpcorp/ScPrime/modules/transactionpool"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/node"

	"gitlab.com/scpcorp/webwallet/modules/bootstrapper"
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

func closeNode(node *node.Node, config *wwConfig.WebWalletConfig, shutdownGui zenity.ProgressDialog) {
	fmt.Println("Closing modules:")
	config.CreateWallet = false
	config.CreateTransactionPool = false
	closeConsensusSetBuilder(shutdownGui)
	config.CreateConsensusSet = false
	config.CreateGateway = false
	closeMiningPool(node, shutdownGui)
	closeStratumMiner(node, shutdownGui)
	closeRenter(node, shutdownGui)
	closeHost(node, shutdownGui)
	closeMiner(node, shutdownGui)
	closeWallet(node, shutdownGui)
	closeTransactionPool(node, shutdownGui)
	closeExplorer(node, shutdownGui)
	closeConsensusSet(node, shutdownGui)
	closeGateway(node, shutdownGui)
	closeMux(node, shutdownGui)
	closeBootstrapper(shutdownGui)
}

func closeConsensusSetBuilder(shutdownGui zenity.ProgressDialog) {
	fmt.Println("Closing consensusset builder...")
	if shutdownGui != nil {
		shutdownGui.Text("Closing consensusset builder...")
	}
	consensusbuilder.Close()
}

func closeMiningPool(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.MiningPool != nil {
		fmt.Println("Closing mining pool...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing mining pool...")
		}
		node.MiningPool.Close()
	}
}

func closeStratumMiner(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.StratumMiner != nil {
		fmt.Println("Closing stratum miner...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing stratum miner...")
		}
		node.StratumMiner.Close()
	}
}

func closeRenter(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.Renter != nil {
		fmt.Println("Closing renter...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing renter...")
		}
		node.Renter.Close()
	}
}

func closeHost(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.Host != nil {
		fmt.Println("Closing host...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing host...")
		}
		node.Host.Close()
	}
}

func closeMiner(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.Miner != nil {
		fmt.Println("Closing miner...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing miner...")
		}
		node.Miner.Close()
	}
}

func closeWallet(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.Wallet != nil {
		fmt.Println("Closing wallet...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing wallet...")
		}
		node.Wallet.Close()
	}
}

func closeTransactionPool(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.TransactionPool != nil {
		fmt.Println("Closing transactionpool...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing transactionpool...")
		}
		node.TransactionPool.Close()
	}
}

func closeExplorer(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.Explorer != nil {
		fmt.Println("Closing explorer...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing explorer...")
		}
		node.Explorer.Close()
	}
}

func closeConsensusSet(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.ConsensusSet != nil {
		fmt.Println("Closing consensusset...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing consensusset...")
		}
		node.ConsensusSet.Close()
	}
}

func closeGateway(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.Gateway != nil {
		fmt.Println("Closing gateway...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing gateway...")
		}
		node.Gateway.Close()
	}
}

func closeMux(node *node.Node, shutdownGui zenity.ProgressDialog) {
	if node.Mux != nil {
		fmt.Println("Closing mux...")
		if shutdownGui != nil {
			shutdownGui.Text("Closing mux...")
		}
		node.Mux.Close()
	}
}

func closeBootstrapper(shutdownGui zenity.ProgressDialog) {
	fmt.Println("Closing bootstrapper...")
	if shutdownGui != nil {
		shutdownGui.Text("Closing bootstrapper...")
	}
	bootstrapper.Close()
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
