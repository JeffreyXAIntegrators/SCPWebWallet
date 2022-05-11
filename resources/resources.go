package resources

import (
	_ "embed" // blank import is a compile-time dependency
)

//go:embed resources/logo.png
var logo []byte

//go:embed resources/favicon.ico
var favicon []byte

//go:embed resources/styles.css
var cssStyleSheet []byte

//go:embed resources/scripts.js
var javascript []byte

//nolint:typecheck // resources/wasm_exec.js is generated dynamically by the release script.
//go:embed resources/wasm_exec.js
var wasmExec []byte

//nolint:typecheck // resources/wallet.wasm is generated dynamically by the release script.
//go:embed resources/wallet.wasm
var walletwasm []byte

//go:embed resources/bootstrapping.html
var bootstrappingHTML string

//go:embed resources/consensus_set_building.html
var consensusSetBuildingHTML string

//go:embed resources/cold_wallet.html
var coldWalletHTML string

//go:embed resources/wallet_template.html
var walletHTMLTemplate string

//go:embed resources/alert_template.html
var alertHTMLTemplate string

//go:embed resources/error_template.html
var errorHTMLTemplate string

//go:embed resources/privacy_template.html
var privacyHTMLTemplate string

//go:embed resources/transaction_templates/history_line_template.html
var transactionHistoryLineHTMLTemplate string

//go:embed resources/transaction_templates/history_template.html
var transactionsHistoryHTMLTemplate string

//go:embed resources/transaction_templates/info_template.html
var transactionInfoTemplate string

//go:embed resources/transaction_templates/input_template.html
var transactionInputTemplate string

//go:embed resources/transaction_templates/output_template.html
var transactionOutputTemplate string

//go:embed resources/forms/starting_wallet.html
var startingWalletForm string

//go:embed resources/forms/close_alert.html
var closeAlertForm string

//go:embed resources/initialize_consensus_set.html
var initializeConsensusSetForm string

//go:embed resources/forms/initialize_seed.html
var initializeSeedForm string

//go:embed resources/forms/initialize_wallet.html
var initializeWalletForm string

//go:embed resources/forms/initialize_browser.html
var initializeBrowserForm string

//go:embed resources/forms/browser_configured.html
var browserConfigured string

//go:embed resources/forms/restore_from_seed.html
var restoreFromSeedForm string

//go:embed resources/forms/scanning_wallet.html
var scanningWalletForm string

//go:embed resources/forms/send_coins.html
var sendCoinsForm string

//go:embed resources/forms/receive_coins_form.html
var receiveCoinsForm string

//go:embed resources/forms/unlock_wallet.html
var unlockWalletForm string

//go:embed resources/forms/change_lock.html
var changeLockForm string

//go:embed resources/forms/collapsed_menu.html
var collapsedMenuForm string

//go:embed resources/forms/expanded_menu.html
var expandedMenuForm string

//go:embed resources/forms/explain_whale.html
var explainWhaleForm string

//go:embed resources/fonts/open-sans-v27-latin/open-sans-v27-latin-regular.woff2
var openSansLatinRegularWoff2 []byte

//go:embed resources/fonts/open-sans-v27-latin/open-sans-v27-latin-700.woff2
var openSansLatin700Woff2 []byte

// Logo returns the Logo.
func Logo() []byte {
	return logo
}

// Favicon returns the favicon.
func Favicon() []byte {
	return favicon
}

// CSSStyleSheet returns the css style sheet.
func CSSStyleSheet() []byte {
	return cssStyleSheet
}

// WasmExec returns the wasm_exec.js binding library.
func WasmExec() []byte {
	return wasmExec
}

// WalletWasm returns the wallet wasm.
func WalletWasm() []byte {
	return walletwasm
}

// Javascript returns the javascript.
func Javascript() []byte {
	return javascript
}

// BootstrappingHTML returns an html page
func BootstrappingHTML() string {
	return bootstrappingHTML
}

// ConsensusSetBuildingHTML returns an html page
func ConsensusSetBuildingHTML() string {
	return consensusSetBuildingHTML
}

// ColdWalletHTML returns an html page
func ColdWalletHTML() string {
	return coldWalletHTML
}

// WalletHTMLTemplate returns the wallet html template
func WalletHTMLTemplate() string {
	return walletHTMLTemplate
}

// AlertHTMLTemplate returns the alert html template
func AlertHTMLTemplate() string {
	return alertHTMLTemplate
}

// ErrorHTMLTemplate returns the error html template
func ErrorHTMLTemplate() string {
	return errorHTMLTemplate
}

// PrivacyHTMLTemplate returns the privacy html template
func PrivacyHTMLTemplate() string {
	return privacyHTMLTemplate
}

// TransactionHistoryLineHTMLTemplate returns an HTML template
func TransactionHistoryLineHTMLTemplate() string {
	return transactionHistoryLineHTMLTemplate
}

// TransactionsHistoryHTMLTemplate returns an HTML template
func TransactionsHistoryHTMLTemplate() string {
	return transactionsHistoryHTMLTemplate
}

// TransactionInfoTemplate returns an HTML template
func TransactionInfoTemplate() string {
	return transactionInfoTemplate
}

// TransactionInputTemplate returns an HTML template
func TransactionInputTemplate() string {
	return transactionInputTemplate
}

// TransactionOutputTemplate returns an HTML template
func TransactionOutputTemplate() string {
	return transactionOutputTemplate
}

// StartingWalletForm returns the starting wallet form
func StartingWalletForm() string {
	return startingWalletForm
}

// CloseAlertForm returns the close alert form
func CloseAlertForm() string {
	return closeAlertForm
}

// InitializeConsensusSetForm returns the initialize consensus set form
func InitializeConsensusSetForm() string {
	return initializeConsensusSetForm
}

// InitializeSeedForm returns the initialize seed form
func InitializeSeedForm() string {
	return initializeSeedForm
}

// InitializeWalletForm returns the initialize wallet form
func InitializeWalletForm() string {
	return initializeWalletForm
}

// InitializeBrowserForm returns the Initialize browser form
func InitializeBrowserForm() string {
	return initializeBrowserForm
}

// BrowserConfigured returns the browser configured alert
func BrowserConfigured() string {
	return browserConfigured
}

// RestoreFromSeedForm returns the restore from seed form
func RestoreFromSeedForm() string {
	return restoreFromSeedForm
}

// ScanningWalletForm returns the scanning wallet alert form
func ScanningWalletForm() string {
	return scanningWalletForm
}

// SendCoinsForm returns the send coins form
func SendCoinsForm() string {
	return sendCoinsForm
}

// ReceiveCoinsForm returns the receive coins form
func ReceiveCoinsForm() string {
	return receiveCoinsForm
}

// UnlockWalletForm returns the unlock wallet form
func UnlockWalletForm() string {
	return unlockWalletForm
}

// ChangeLockForm returns the change lock form
func ChangeLockForm() string {
	return changeLockForm
}

// ExpandedMenuForm returns the HTML form
func ExpandedMenuForm() string {
	return expandedMenuForm
}

// ExplainWhaleForm returns the HTML form
func ExplainWhaleForm() string {
	return explainWhaleForm
}

// CollapsedMenuForm returns the HTML form
func CollapsedMenuForm() string {
	return collapsedMenuForm
}

// OpenSansLatinRegularWoff2 returns the open sans font
func OpenSansLatinRegularWoff2() []byte {
	return openSansLatinRegularWoff2
}

// OpenSansLatin700Woff2 returns the open sans font
func OpenSansLatin700Woff2() []byte {
	return openSansLatin700Woff2
}
