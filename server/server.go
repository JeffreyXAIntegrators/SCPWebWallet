package server

import (
	"crypto/rand"
	"encoding/hex"
	checkErrors "errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/NebulousLabs/errors"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/node"

	wwConfig "gitlab.com/scpcorp/webwallet/utils/config"
)

var (
	n        *node.Node
	config   *wwConfig.WebWalletConfig
	srv      *http.Server
	status   string
	sessions []*Session
	waitCh   chan struct{}
)

// Session is a struct that tracks session settings
type Session struct {
	id            string
	alert         string
	collapseMenu  bool
	txHistoryPage int
	cachedPage    string
	wallet        modules.Wallet
	name          string
}

// StartHTTPServer starts the HTTP server to serve the GUI.
func StartHTTPServer(webWalletConfig *wwConfig.WebWalletConfig) {
	config = webWalletConfig
	wg := &sync.WaitGroup{}
	wg.Add(1)
	srv = &http.Server{Addr: ":4300", Handler: buildHTTPRoutes()}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("Unable to start server: %v\n", err)
			srv = nil
		}
	}()
	waitCh = make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()
}

// IsRunning returns true when the server is running
func IsRunning() bool {
	if srv == nil {
		return false
	}
	for i := 0; i < 20; i++ {
		time.Sleep(5 * time.Millisecond)
		if srv == nil {
			return false
		}
	}
	return srv != nil
}

// Wait returns the servers wait channel
func Wait() chan struct{} {
	return waitCh
}

// AttachNode attaches the node to the HTTP server.
func AttachNode(node *node.Node) {
	n = node
	if srv != nil {
		srv.Handler = buildHTTPRoutes()
	}
}

// newWallet attaches a newly created wallet module to the session.
func newWallet(walletDirName string, sessionID string) (modules.Wallet, error) {
	loadStart := time.Now()
	walletDeps := modules.ProdDependencies
	fmt.Printf("Loading wallet...")
	walletDir := filepath.Join(n.Dir, "wallets", walletDirName)
	_, err := os.Stat(walletDir)
	if err == nil {
		return nil, fmt.Errorf("%s already exists", walletDirName)
	}
	cs := n.ConsensusSet
	tp := n.TransactionPool
	session, err := getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session.wallet != nil {
		errors.New("session already has a wallet loaded")
	}
	w, err := wallet.NewCustomWallet(cs, tp, walletDir, walletDeps)
	if err != nil {
		return nil, err
	}
	session.wallet = w
	session.name = walletDirName
	fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	return w, nil
}

// existingWallet attaches an existing wallet module to the session.
func existingWallet(walletDirName string, sessionID string) (modules.Wallet, error) {
	loadStart := time.Now()
	walletDeps := modules.ProdDependencies
	fmt.Printf("Loading wallet...")
	walletDir := filepath.Join(n.Dir, "wallets", walletDirName)
	_, err := os.Stat(walletDir)
	if checkErrors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%s does not exist", walletDirName)
	}
	cs := n.ConsensusSet
	tp := n.TransactionPool
	session, err := getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session.wallet != nil {
		errors.New("session already has a wallet loaded")
	}
	w, err := wallet.NewCustomWallet(cs, tp, walletDir, walletDeps)
	if err != nil {
		return nil, err
	}
	session.wallet = w
	session.name = walletDirName
	fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	return w, nil
}

// closeWallet closes the wallet and detaches it from the node.
func closeWallet(sessionID string) error {
	session, err := getSession(sessionID)
	if err != nil {
		return err
	}
	wallet := session.wallet
	if wallet != nil {
		session.wallet = nil
		session.name = ""
		fmt.Println("Closing wallet...")
		err = wallet.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// CloseAllWallets closes all wallets and detaches them from the node.
func CloseAllWallets() (err error) {
	for _, session := range sessions {
		wallet := session.wallet
		if wallet != nil {
			session.wallet = nil
			session.name = ""
			fmt.Println("Closing wallet...")
			err = errors.Compose(wallet.Close())
		}
	}
	return err
}

func getWallet(sessionID string) (modules.Wallet, error) {
	session, err := getSession(sessionID)
	if err != nil {
		return nil, err
	} else if session.wallet == nil {
		return nil, errors.New("no wallet is attached to the session")
	}
	return session.wallet, nil
}

// setStatus sets the status.
func setStatus(s string) {
	status = s
}

// addSessionId adds a new session ID to memory.
func addSessionID() string {
	b := make([]byte, 16) //32 characters long
	rand.Read(b)
	session := &Session{}
	session.id = hex.EncodeToString(b)
	session.collapseMenu = true
	session.txHistoryPage = 1
	session.cachedPage = ""
	sessions = append(sessions, session)
	return session.id
}

// getSession returns the session.
func getSession(sessionID string) (*Session, error) {
	for _, session := range sessions {
		if session.id == sessionID {
			return session, nil
		}
	}
	return nil, errors.New("session ID was not found")
}

// sessionIDExists returns true when the supplied session ID exists in memory.
func sessionIDExists(sessionID string) bool {
	session, _ := getSession(sessionID)
	return session != nil
}

// setAlert sets an alert on the session.
func setAlert(alert string, sessionID string) {
	session, _ := getSession(sessionID)
	if session != nil {
		session.alert = alert
	}
}

// hasAlert returns true when the session has an alert.
func hasAlert(sessionID string) bool {
	session, _ := getSession(sessionID)
	if session != nil {
		return session.alert != ""
	}
	return false
}

// popAlert gets the alert from the session and then clears it from the session.
func popAlert(sessionID string) string {
	session, _ := getSession(sessionID)
	if session != nil {
		alert := session.alert
		session.alert = ""
		return alert
	}
	return ""
}

// collapseMenu sets the menu state to collapsed and returns true
func collapseMenu(sessionID string) bool {
	session, _ := getSession(sessionID)
	if session != nil {
		session.collapseMenu = true
	}
	return true
}

// expandMenu sets the menu state to expanded and returns true
func expandMenu(sessionID string) bool {
	session, _ := getSession(sessionID)
	if session != nil {
		session.collapseMenu = false
	}
	return true
}

// menuIsCollapsed returns true when the menu state is collapsed
func menuIsCollapsed(sessionID string) bool {
	session, _ := getSession(sessionID)
	if session != nil {
		return session.collapseMenu
	}
	// default to the menu being expanded just in case
	return false
}

// setTxHistoryPage sets the session's transaction history page and returns true.
func setTxHistoryPage(txHistoryPage int, sessionID string) bool {
	session, _ := getSession(sessionID)
	if session != nil {
		session.txHistoryPage = txHistoryPage
	}
	return true
}

// getTxHistoryPage returns the session's transaction history page or -1 when no session is found.
func getTxHistoryPage(sessionID string) int {
	session, _ := getSession(sessionID)
	if session != nil {
		return session.txHistoryPage
	}
	return -1
}

// cachedPage caches the page without the menu and returns true.
func cachedPage(cachedPage string, sessionID string) bool {
	session, _ := getSession(sessionID)
	if session != nil {
		session.cachedPage = cachedPage
	}
	return true
}

// getCachedPage returns the session's cached page.
func getCachedPage(sessionID string) string {
	session, _ := getSession(sessionID)
	if session != nil {
		return session.cachedPage
	}
	return ""
}
