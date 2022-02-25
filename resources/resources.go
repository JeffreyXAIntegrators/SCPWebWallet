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

//go:embed resources/downloading.html
var downloadingHtml string

//go:embed resources/loading.html
var loadingHtml []byte

//go:embed resources/wallet_template.html
var walletHtmlTemplate string

//go:embed resources/alert_template.html
var alertHtmlTemplate string

//go:embed resources/transaction_templates/history_line_template.html
var transactionHistoryLineHtmlTemplate string

//go:embed resources/transaction_templates/history_template.html
var transactionsHistoryHtmlTemplate string

//go:embed resources/transaction_templates/info_template.html
var transactionInfoTemplate string

//go:embed resources/transaction_templates/input_template.html
var transactionInputTemplate string

//go:embed resources/transaction_templates/output_template.html
var transactionOutputTemplate string

//go:embed resources/transaction_templates/transaction_pagination_template.html
var transactionPaginationTemplate string

//go:embed resources/forms/close_alert.html
var closeAlertForm string

//go:embed resources/forms/initialize_seed.html
var intializeSeedForm string

//go:embed resources/forms/initialize_wallet.html
var initializeWalletForm string

//go:embed resources/forms/restore_from_seed.html
var restoreFromSeedForm string

//go:embed resources/forms/scanning_wallet.html
var scanningWalletForm string

//go:embed resources/forms/send_coins.html
var sendCoinsForm string

//go:embed resources/forms/unlock_wallet.html
var unlockWalletForm string

//go:embed resources/forms/change_lock.html
var changeLockForm string

//go:embed resources/forms/collapsed_menu.html
var collapsedMenuForm string

//go:embed resources/forms/expanded_menu.html
var expandedMenuForm string

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

// CssStyleSheet returns the css style sheet.
func CssStyleSheet() []byte {
	return cssStyleSheet
}

// Javascript returns the javascript.
func Javascript() []byte {
	return javascript
}

// DownloadingHtml returns an html page
func DownloadingHtml() string {
	return downloadingHtml
}

// LoadingHtml returns an html page
func LoadingHtml() []byte {
	return loadingHtml
}

// WalletHtmlTemplate returns the wallet html template
func WalletHtmlTemplate() string {
	return walletHtmlTemplate
}

// AlertHtmlTemplate returns the alert html template
func AlertHtmlTemplate() string {
	return alertHtmlTemplate
}

// TransactionHistoryLineHtmlTemplate returns an HTML template
func TransactionHistoryLineHtmlTemplate() string {
	return transactionHistoryLineHtmlTemplate
}

// TransactionsHistoryHtmlTemplate returns an HTML template
func TransactionsHistoryHtmlTemplate() string {
	return transactionsHistoryHtmlTemplate
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

// TransactionPaginationTemplate returns an HTML template
func TransactionPaginationTemplate() string {
	return transactionPaginationTemplate
}

// CloseAlertForm returns the close alert form
func CloseAlertForm() string {
	return closeAlertForm
}

// IntializeSeedForm returns the initialize seed form
func IntializeSeedForm() string {
	return intializeSeedForm
}

// InitializeWalletForm returns the initialize wallet form
func InitializeWalletForm() string {
	return initializeWalletForm
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