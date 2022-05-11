package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func buildHTTPRoutes() *httprouter.Router {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(notFoundHandler)
	router.RedirectTrailingSlash = false

	//GUI Calls
	router.GET("/favicon.ico", faviconHandler)
	router.GET("/gui/bootstrapperProgress", bootstrapperProgressHandler)
	router.GET("/gui/consensusBuilderProgress", consensusBuilderProgressHandler)
	router.GET("/gui/logo.png", logoHandler)
	router.GET("/gui/scripts.js", scriptHandler)
	router.GET("/gui/wasm_exec.js", wasmExecHandler)
	router.GET("/gui/wallet.wasm", walletWasmHandler)
	router.GET("/gui/styles.css", styleHandler)
	router.GET("/gui/fonts/open-sans-v27-latin-regular.woff2", openSansLatinRegularWoff2Handler)
	router.GET("/gui/fonts/open-sans-v27-latin-700.woff2", openSansLatin700Woff2Handler)
	router.GET("/initializeColdWallet", coldWalletHandler)
	router.POST("/gui/heartbeat", heartbeatHandler)
	if n == nil {
		router.GET("/", initializingNodeHandler)
		router.GET("/initializeBootstrapper", initializeBootstrapperHandler)
		router.GET("/skipBootstrapper", skipBootstrapperHandler)
		router.GET("/initializeConsensusBuilder", initializeConsensusBuilderHandler)
		router.GET("/configureBrowser", redirect)
		router.POST("/configureBrowser", configureBrowser)
	} else {
		router.GET("/", guiHandler)
		router.GET("/gui", guiHandler)
		router.GET("/gui/export", redirect)
		router.GET("/gui/alert/changeLock", redirect)
		router.GET("/gui/alert/initializeSeed", redirect)
		router.GET("/gui/alert/sendCoins", redirect)
		router.GET("/gui/alert/receiveCoins", redirect)
		router.GET("/gui/alert/recoverSeed", redirect)
		router.GET("/gui/alert/restoreFromSeed", redirect)
		router.GET("/gui/changeLock", redirect)
		router.GET("/gui/collapseMenu", redirect)
		router.GET("/gui/expandMenu", redirect)
		router.GET("/gui/explainWhale", redirect)
		router.GET("/gui/initializeSeed", redirect)
		router.GET("/gui/lockWallet", redirect)
		router.GET("/gui/privacy", redirect)
		router.GET("/gui/restoreSeed", redirect)
		router.GET("/gui/scanning", redirect)
		router.GET("/gui/sendCoins", redirect)
		router.GET("/gui/setTxHistoryPage", redirect)
		router.GET("/gui/unlockWallet", redirect)
		router.GET("/gui/unlockWalletForm", redirect)
		router.GET("/gui/explorer", redirect)
		router.POST("/gui", guiHandler)
		router.POST("/gui/export", transactionHistoryCsvExport)
		router.POST("/gui/alert/changeLock", alertChangeLockHandler)
		router.POST("/gui/alert/initializeSeed", alertInitializeSeedHandler)
		router.POST("/gui/alert/sendCoins", alertSendCoinsHandler)
		router.POST("/gui/alert/receiveCoins", alertReceiveCoinsHandler)
		router.POST("/gui/alert/recoverSeed", alertRecoverSeedHandler)
		router.POST("/gui/alert/restoreFromSeed", alertRestoreFromSeedHandler)
		router.POST("/gui/changeLock", changeLockHandler)
		router.POST("/gui/collapseMenu", collapseMenuHandler)
		router.POST("/gui/expandMenu", expandMenuHandler)
		router.POST("/gui/explainWhale", explainWhaleHandler)
		router.POST("/gui/initializeSeed", initializeSeedHandler)
		router.POST("/gui/lockWallet", lockWalletHandler)
		router.POST("/gui/privacy", privacyHandler)
		router.POST("/gui/restoreSeed", restoreSeedHandler)
		router.POST("/gui/scanning", scanningHandler)
		router.POST("/gui/sendCoins", sendCoinsHandler)
		router.POST("/gui/setTxHistoryPage", setTxHistoyPage)
		router.POST("/gui/unlockWallet", unlockWalletHandler)
		router.POST("/gui/unlockWalletForm", unlockWalletFormHandler)
		router.POST("/gui/explorer", explorerHandler)
		router.POST("/gui/balance", balanceHandler)
		router.POST("/gui/blockHeight", blockHeightHandler)
	}
	return router
}
