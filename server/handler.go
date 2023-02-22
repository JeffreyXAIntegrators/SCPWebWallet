package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gitlab.com/scpcorp/webwallet/build"
	"gitlab.com/scpcorp/webwallet/modules/bootstrapper"
	"gitlab.com/scpcorp/webwallet/modules/browserconfig"
	consensusbuilder "gitlab.com/scpcorp/webwallet/modules/consensesbuilder"
	"gitlab.com/scpcorp/webwallet/resources"

	nebErrors "gitlab.com/NebulousLabs/errors"

	spdBuild "gitlab.com/scpcorp/ScPrime/build"
	"gitlab.com/scpcorp/ScPrime/crypto"
	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/consensus"
	"gitlab.com/scpcorp/ScPrime/types"

	"github.com/julienschmidt/httprouter"
	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
)

func notFoundHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "404 not found.", http.StatusNotFound)
}

func redirect(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	http.Redirect(w, req, "/", http.StatusMovedPermanently)
}

func optionsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.Header().Set("Content-Length", "0")
	w.Header().Set("Application", "ScPrime Web Wallet")
}

func faviconHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var favicon = resources.Favicon()
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Content-Length", strconv.Itoa(len(favicon))) //len(dec)
	w.Write(favicon)
}

func balanceHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	fmtScpBal, fmtUncBal, fmtSpfaBal, fmtSpfbBal, fmtClmBal, fmtWhale := balancesHelper(sessionID)
	writeArray(w, []string{fmtScpBal, fmtUncBal, fmtSpfaBal, fmtSpfbBal, fmtClmBal, fmtWhale})
}

func blockHeightHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	fmtHeight, fmtStatus, fmtStatCo := blockHeightHelper(sessionID)
	writeArray(w, []string{fmtHeight, fmtStatus, fmtStatCo})
}

func bootstrapperProgressHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	writeArray(w, []string{bootstrapper.Progress()})
}

func consensusBuilderProgressHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	writeArray(w, []string{consensusbuilder.Progress()})
}

func logoHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var logo = resources.Logo()
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(logo))) //len(dec)
	w.Write(logo)
}

func scriptHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var javascript = resources.Javascript()
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(javascript))) //len(dec)
	w.Write(javascript)
}

func wasmExecHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var wasmExec = resources.WasmExec()
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(wasmExec))) //len(dec)
	w.Write(wasmExec)
}

func walletWasmHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var walletwasm = resources.WalletWasm()
	w.Header().Set("Content-Type", "application/wasm")
	w.Header().Set("Content-Length", strconv.Itoa(len(walletwasm))) //len(dec)
	w.Write(walletwasm)
}

func styleHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var cssStyleSheet = resources.CSSStyleSheet()
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(cssStyleSheet))) //len(dec)
	w.Write(cssStyleSheet)
}

func coldWalletHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	style := resources.CSSStyleSheet()
	script := resources.Javascript()
	logo := resources.Logo()
	wasmExec := resources.WasmExec()
	walletWasm := resources.WalletWasm()
	html := resources.ColdWalletHTML()
	html = strings.Replace(html, "&STYLE;", string(style), -1)
	html = strings.Replace(html, "&SCRIPT;", string(script), -1)
	html = strings.Replace(html, "&LOGO;", base64.StdEncoding.EncodeToString(logo), -1)
	html = strings.Replace(html, "&WASM_EXEC;", string(wasmExec), -1)
	html = strings.Replace(html, "&WALLET_WASM;", base64.StdEncoding.EncodeToString(walletWasm), -1)
	htmlBytes := []byte(html)
	w.Header().Set("Content-Disposition", "attachment; filename=scprime-cold-wallet.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(htmlBytes))) //len(dec)
	w.Write(htmlBytes)
}

func transactionHistoryCsvExport(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	msgPrefix := "Unable to export transaction history: "
	sessionID := req.FormValue("session_id")
	if sessionID == "" {
		msg := fmt.Sprintf("%s%v", msgPrefix, "Session ID was not supplied.")
		writeError(w, msg, "")
		return
	} else if !sessionIDExists(sessionID) {
		msg := fmt.Sprintf("%s%v", msgPrefix, "Session ID does not exist.")
		writeError(w, msg, "")
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	history, err := transctionHistoryCsvExportHelper(wallet)
	if err != nil {
		history = "failed"
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-disposition", "attachment;filename=history.csv")
	w.Header().Set("Content-Length", strconv.Itoa(len(history))) //len(dec)
	w.Write([]byte(history))
}

func transctionHistoryCsvExportHelper(wallet modules.Wallet) (string, error) {
	csv := `"Transaction ID","Type","Amount SCP","Amount SPF-A","Amount SPF-B","Fee SCP", "Confirmed","DateTime"` + "\n"
	heightMin := 0
	confirmedTxns, err := wallet.Transactions(types.BlockHeight(heightMin), n.ConsensusSet.Height())
	if err != nil {
		return "", err
	}
	unconfirmedTxns, err := wallet.UnconfirmedTransactions()
	if err != nil {
		return "", err
	}
	sts, err := ComputeSummarizedTransactions(append(confirmedTxns, unconfirmedTxns...), n.ConsensusSet.Height(), wallet)
	if err != nil {
		return "", err
	}
	for _, txn := range sts {
		// Format transaction type
		if txn.Type != "SETUP" {
			csv = csv + fmt.Sprintf(`"%s","%s","%f","%f","%f","%f","%s","%s\n"`, txn.TxnID, txn.Type, txn.Scp, txn.SpfA, txn.SpfB, txn.ScpFee, txn.Confirmed, txn.Time)
		}
	}
	return csv, nil
}

func privacyHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	html := resources.WalletHTMLTemplate()
	html = strings.Replace(html, "&TRANSACTION_PORTAL;", resources.PrivacyHTMLTemplate(), -1)
	writeHTML(w, html, sessionID)
}

func alertChangeLockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	title := "CHANGE LOCK"
	form := resources.ChangeLockForm()
	writeForm(w, title, form, sessionID)
}

func alertInitializeSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	writeStaticHTML(w, resources.InitializeSeedForm(), "")
}

func alertSendCoinsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	title := "SEND"
	form := resources.SendCoinsForm()
	writeForm(w, title, form, sessionID)
}

func alertReceiveCoinsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	var msgPrefix = "Unable to retrieve address: "
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	addresses, err := wallet.LastAddresses(1)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	if len(addresses) == 0 {
		_, err := wallet.NextAddress()
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
	}
	addresses, err = wallet.LastAddresses(1)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	address := strings.ToUpper(fmt.Sprintf("%s", addresses[0]))
	title := "RECEIVE"
	formHTML := resources.ReceiveCoinsForm()
	formHTML = strings.Replace(formHTML, "&ADDRESS;", address, -1)
	writeForm(w, title, formHTML, sessionID)
}

func alertRecoverSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
		return
	}
	cancel := req.FormValue("cancel")
	var msgPrefix = "Unable to recover seed: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	unlocked, err := wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if !unlocked {
		msg := msgPrefix + "Wallet is locked."
		writeError(w, msg, "")
		return
	}
	// Get the primary seed information.
	dictionary := mnemonics.DictionaryID(req.FormValue("dictionary"))
	if dictionary == "" {
		dictionary = mnemonics.English
	}
	primarySeed, _, err := wallet.PrimarySeed()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	primarySeedStr, err := modules.SeedToString(primarySeed, dictionary)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	title := "RECOVER SEED"
	msg := fmt.Sprintf("%s", primarySeedStr)
	writeMsg(w, title, msg, sessionID)
}

func alertRestoreFromSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	writeStaticHTML(w, resources.RestoreFromSeedForm(), "")
}

func unlockWalletFormHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	writeStaticHTML(w, resources.UnlockWalletForm(), "")
}

func changeLockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	cancel := req.FormValue("cancel")
	origPassword := req.FormValue("orig_password")
	newPassword := req.FormValue("new_password")
	confirmPassword := req.FormValue("confirm_password")
	var msgPrefix = "Unable to change lock: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	if origPassword == "" {
		msg := msgPrefix + "The original password must be provided."
		writeError(w, msg, sessionID)
		return
	}
	if newPassword == "" {
		msg := msgPrefix + "A new password must be provided."
		writeError(w, msg, sessionID)
		return
	}
	if len(newPassword) < 8 {
		msg := msgPrefix + "Password must be at least eight characters long."
		writeError(w, msg, "")
		return
	}
	if confirmPassword == "" {
		msg := msgPrefix + "A confirmation password must be provided."
		writeError(w, msg, sessionID)
		return
	}
	if newPassword != confirmPassword {
		msg := msgPrefix + "New password does not match confirmation password."
		writeError(w, msg, sessionID)
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	validPass, err := isPasswordValid(wallet, origPassword)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	} else if !validPass {
		msg := msgPrefix + "The original password is not valid."
		writeError(w, msg, sessionID)
		return
	}
	var newKey crypto.CipherKey
	newKey = crypto.NewWalletKey(crypto.HashObject(newPassword))
	primarySeed, _, _ := wallet.PrimarySeed()
	err = wallet.ChangeKeyWithSeed(primarySeed, newKey)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	guiHandler(w, req, nil)
}

func initializeSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	cancel := req.FormValue("cancel")
	walletDirName := req.FormValue("wallet_dir_name")
	if walletDirName == "" {
		walletDirName = "wallet"
	}
	newPassword := req.FormValue("new_password")
	confirmPassword := req.FormValue("confirm_password")
	var msgPrefix = "Unable to initialize new wallet seed: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	if newPassword == "" {
		msg := msgPrefix + "A new password must be provided."
		writeError(w, msg, "")
		return
	}
	if len(newPassword) < 8 {
		msg := msgPrefix + "Password must be at least eight characters long."
		writeError(w, msg, "")
		return
	}
	if confirmPassword == "" {
		msg := msgPrefix + "A confirmation password must be provided."
		writeError(w, msg, "")
		return
	}
	if newPassword != confirmPassword {
		msg := msgPrefix + "New password does not match confirmation password."
		writeError(w, msg, "")
		return
	}
	sessionID := addSessionID()
	wallet, err := newWallet(walletDirName, sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	encrypted, err := wallet.Encrypted()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if encrypted {
		msg := msgPrefix + "Seed was already initialized."
		writeError(w, msg, "")
		return
	}
	go initializeSeedHelper(newPassword, sessionID)
	title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
	form := resources.ScanningWalletForm()
	writeForm(w, title, form, sessionID)
}

func lockWalletHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	cancel := req.FormValue("cancel")
	var msgPrefix = "Unable to lock wallet: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	unlocked, err := wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if !unlocked {
		msg := msgPrefix + "Wallet was already locked."
		writeError(w, msg, "")
		return
	}
	wallet.Lock()
	closeWallet(sessionID)
	redirect(w, req, nil)
}

func restoreSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	cancel := req.FormValue("cancel")
	walletDirName := req.FormValue("wallet_dir_name")
	if walletDirName == "" {
		walletDirName = "wallet"
	}
	newPassword := req.FormValue("new_password")
	confirmPassword := req.FormValue("confirm_password")
	seedStr := req.FormValue("seed_str")
	var msgPrefix = "Unable to restore wallet from seed: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	if newPassword == "" {
		msg := msgPrefix + "A new password must be provided."
		writeError(w, msg, "")
		return
	}
	if confirmPassword == "" {
		msg := msgPrefix + "A confirmation password must be provided."
		writeError(w, msg, "")
		return
	}
	if newPassword != confirmPassword {
		msg := msgPrefix + "New password does not match confirmation password."
		writeError(w, msg, "")
		return
	}
	if seedStr == "" {
		msg := msgPrefix + "A seed must be provided."
		writeError(w, msg, "")
		return
	}
	sessionID := addSessionID()
	wallet, err := newWallet(walletDirName, sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	encrypted, err := wallet.Encrypted()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if encrypted {
		msg := msgPrefix + "Seed is already initialized."
		writeError(w, msg, "")
		return
	}
	seed, err := modules.StringToSeed(seedStr, "english")
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	go restoreSeedHelper(newPassword, seed, sessionID)
	title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
	form := resources.ScanningWalletForm()
	writeForm(w, title, form, sessionID)
}

func sendCoinsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	cancel := req.FormValue("cancel")
	var msgPrefix = "Unable to send coins: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	unlocked, err := wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if !unlocked {
		msg := msgPrefix + "Wallet is locked."
		writeError(w, msg, "")
		return
	}
	// Verify destination address was supplied.
	dest, err := scanAddress(req.FormValue("destination"))
	if err != nil {
		msg := msgPrefix + "Destination is not valid."
		writeError(w, msg, sessionID)
		return
	}
	coinType := req.FormValue("coin_type")
	if coinType == "SCP" {
		amount, err := NewCurrencyStr(req.FormValue("amount") + "SCP")
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
		_, err = wallet.SendSiacoins(amount, dest)
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
	} else if coinType == "SPF-A" {
		amount, err := NewCurrencyStr(req.FormValue("amount") + "SPF")
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
		_, err = wallet.SendSiafunds(amount, dest)
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
	} else if coinType == "SPF-B" {
		amount, err := NewCurrencyStr(req.FormValue("amount") + "SPF")
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
		_, err = wallet.SendSiafundbs(amount, dest)
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
	} else {
		msg := msgPrefix + "Coin type was not supplied."
		writeError(w, msg, sessionID)
		return
	}
	guiHandler(w, req, nil)
}

func unlockWalletHelper(wallet modules.Wallet, password string, sessionID string) {
	var msgPrefix = "Unable to unlock wallet: "
	if password == "" {
		msg := "A password must be provided."
		setAlert(msgPrefix+msg, sessionID)
		if status == "Scanning" {
			status = ""
		}
		return
	}
	potentialKeys, _ := encryptionKeys(password)
	for _, key := range potentialKeys {
		unlocked, err := wallet.Unlocked()
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			setAlert(msg, sessionID)
			if status == "Scanning" {
				status = ""
			}
			return
		}
		if !unlocked {
			wallet.Unlock(key)
		}
	}
	unlocked, err := wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		setAlert(msg, sessionID)
		if status == "Scanning" {
			status = ""
		}
		return
	}
	if !unlocked {
		msg := msgPrefix + "Password is not valid."
		setAlert(msg, sessionID)
	}
	status = ""
}

func unlockWalletHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	cancel := req.FormValue("cancel")
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	password := req.FormValue("password")
	walletDirName := req.FormValue("wallet_dir_name")
	if walletDirName == "" {
		walletDirName = "wallet"
	}
	sessionID := addSessionID()
	wallet, err := existingWallet(walletDirName, sessionID)
	if err != nil {
		msg := fmt.Sprintf("Unable to unlock wallet: %v", err)
		writeError(w, msg, sessionID)
		return
	}
	status = "Scanning"
	go unlockWalletHelper(wallet, password, sessionID)
	time.Sleep(300 * time.Millisecond)
	if status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionID)
		return
	}
	writeWallet(w, wallet, sessionID)
}

func explainWhaleHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	title := "WHAT WHALE ARE YOU?"
	form := resources.ExplainWhaleForm()
	writeForm(w, title, form, sessionID)
}

func explorerHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		redirect(w, req, nil)
		return
	}
	var msgPrefix = "Unable to retrieve the transaction: "
	if req.FormValue("transaction_id") == "" {
		msg := msgPrefix + "No transaction ID was provided."
		writeError(w, msg, sessionID)
		return
	}
	var transactionID types.TransactionID
	jsonID := "\"" + req.FormValue("transaction_id") + "\""
	err := transactionID.UnmarshalJSON([]byte(jsonID))
	if err != nil {
		msg := msgPrefix + "Unable to parse transaction ID."
		writeError(w, msg, sessionID)
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	txn, ok, err := wallet.Transaction(transactionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	if !ok {
		msg := msgPrefix + "Transaction was not found."
		writeError(w, msg, sessionID)
		return
	}
	transactionDetails, _ := transactionExplorerHelper(txn)
	html := resources.WalletHTMLTemplate()
	html = strings.Replace(html, "&TRANSACTION_PORTAL;", transactionDetails, -1)
	writeHTML(w, html, sessionID)
}

func configureBrowser(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	browser := req.FormValue("browser")
	browserconfig.Configure(build.ScPrimeWebWalletDir(), browser)
	if browser != "default" {
		html := resources.BrowserConfigured()
		writeStaticHTML(w, html, "")
		return
	}
	for i := 0; i < 10; i++ {
		if n != nil {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	redirect(w, req, nil)
}

func initializingNodeHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	browserconfig.Initialize()
	if browserconfig.Status() == browserconfig.Waiting {
		writeStaticHTML(w, resources.InitializeBrowserForm(), "")
	} else if consensusbuilder.Progress() != "" {
		buildingConsensusSetHandler(w, req, nil)
	} else if bootstrapper.Progress() != "" {
		bootstrappingHandler(w, req, nil)
	} else {
		initializeConsensusSetFormHandler(w, req, nil)
	}
}

func initializeConsensusSetFormHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	message := "Consensus set was not found"
	if bootstrapper.LocalConsensusSize > 0 {
		message = "Consensus set is out of date"
	}
	html := strings.Replace(resources.InitializeConsensusSetForm(), "&CONSENSUS_MESSAGE;", message, -1)
	writeStaticHTML(w, html, "")
}

func initializeBootstrapperHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	bootstrapper.Initialize()
	bootstrappingHandler(w, req, nil)
}

func skipBootstrapperHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	bootstrapper.Skip()
	time.Sleep(50 * time.Millisecond)
	consensusbuilder.Initialize()
	buildingConsensusSetHandler(w, req, nil)
}

func initializeConsensusBuilderHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	bootstrapper.Skip()
	time.Sleep(50 * time.Millisecond)
	consensusbuilder.Initialize()
	buildingConsensusSetHandler(w, req, nil)
}

func bootstrappingHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	progress := bootstrapper.Progress()
	html := strings.Replace(resources.BootstrappingHTML(), "&BOOTSTRAPPER_PROGRESS;", progress, -1)
	writeStaticHTML(w, html, "")
}

func buildingConsensusSetHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	progress := consensusbuilder.Progress()
	html := strings.Replace(resources.ConsensusSetBuildingHTML(), "&CONSENSUS_BUILDER_PROGRESS;", progress, -1)
	writeStaticHTML(w, html, "")
}

func uploadMultispendCsvFormHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	title := "MULTISEND FROM CSV"
	form := resources.MultiSendCoinsForm()
	writeForm(w, title, form, sessionID)
}

func uploadMultispendCsvHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	cancel := req.FormValue("cancel")
	var msgPrefix = "Unable to send coins: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	file, _, err := req.FormFile("file")
	if err != nil {
		msg := fmt.Sprintf("%s%s%v", msgPrefix, "Unable to upload multispend csv file: ", err)
		writeError(w, msg, sessionID)
		return
	}
	lines, err := csv.NewReader(file).ReadAll()
	err = file.Close()
	if err != nil {
		msg := fmt.Sprintf("%s%s%v", msgPrefix, "Failed to close the multispend csv upload stream: ", err)
		writeError(w, msg, sessionID)
		return
	}
	var coinOutputs []types.SiacoinOutput
	var fundAOutputs []types.SiafundOutput
	var fundBOutputs []types.SiafundOutput
	for _, line := range lines {
		if line == nil {
			continue
		}
		amount := line[0]
		if amount == "" {
			continue
		}
		dest := line[1]
		if dest == "" {
			continue
		}

		// SPF-A and SPF-B are not recognised as currency units, so just "SPF" should be supplied
		// and different methods called for transfer
		value, err := types.NewCurrencyStr(
			strings.ReplaceAll(
				strings.ReplaceAll(amount, "SPF-A", "SPF"),
				"SPF-B", "SPF"))
		if err != nil {
			msg := fmt.Sprintf("%sCould not parse amount: %s: %v", msgPrefix, amount, err)
			writeError(w, msg, sessionID)
			return
		}
		var hash types.UnlockHash
		if _, err := fmt.Sscan(dest, &hash); err != nil {
			msg := fmt.Sprintf("%s%s%v", msgPrefix, "Failed to parse destination address: ", err)
			writeError(w, msg, sessionID)
			return
		}
		if strings.HasSuffix(amount, "SPF") || strings.HasSuffix(amount, "SPF-A") {
			var output types.SiafundOutput
			output.Value = value
			output.UnlockHash = hash
			fundAOutputs = append(fundAOutputs, output)
		} else if strings.HasSuffix(amount, "SPF-B") {
			var output types.SiafundOutput
			output.Value = value
			output.UnlockHash = hash
			fundBOutputs = append(fundBOutputs, output)
		} else {
			var output types.SiacoinOutput
			output.Value = value
			output.UnlockHash = hash
			coinOutputs = append(coinOutputs, output)
		}
	}
	if len(coinOutputs) == 0 && len(fundAOutputs) == 0 && len(fundBOutputs) == 0 {
		msg := fmt.Sprintf("%s%s%v", msgPrefix, "No ScPrime outputs were supplied: ", err)
		writeError(w, msg, sessionID)
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	unlocked, err := wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	if !unlocked {
		msg := msgPrefix + "Wallet is locked."
		writeError(w, msg, sessionID)
		return
	}
	// Mock the transaction to verify that the wallet is able to fund all transactions.
	// SPF-A and SPF-B cannot be sent in same transaction
	txnBuilder, err := wallet.BuildUnsignedBatchTransaction(coinOutputs, fundAOutputs, nil)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	txnBuilder.Drop()
	// test SCP+SPF-B
	txnBuilder, err = wallet.BuildUnsignedBatchTransaction(coinOutputs, nil, fundBOutputs)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionID)
		return
	}
	txnBuilder.Drop()
	// Send the transactions
	if len(coinOutputs) != 0 {
		_, err = wallet.SendBatchTransaction(coinOutputs, nil, nil)
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
	}
	if len(fundAOutputs) != 0 {
		_, err = wallet.SendBatchTransaction(nil, fundAOutputs, nil)
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
	}
	if len(fundBOutputs) != 0 {
		_, err = wallet.SendBatchTransaction(nil, nil, fundBOutputs)
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionID)
			return
		}
	}
	guiHandler(w, req, nil)
}

func uploadConsensusSetFormHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	writeStaticHTML(w, resources.ConsensusSetUploadingHTML(), "")
}

func uploadConsensusSetHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	file, _, err := req.FormFile("file")
	if err != nil {
		msg := fmt.Sprintf("Unable to upload consensus set: %v", err)
		writeError(w, msg, "")
		return
	}
	consensusDir := filepath.Join(config.Dir, modules.ConsensusDir)
	consensusDb := filepath.Join(consensusDir, consensus.DatabaseFilename)
	_, err = os.Stat(consensusDir)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(consensusDir, os.ModePerm)
	}
	if err != nil {
		msg := fmt.Sprintf("Unable to create the consensus directory: %v", err)
		writeError(w, msg, "")
		return
	}
	out, err := os.Create(consensusDb)
	if err != nil {
		msg := fmt.Sprintf("Failed to open the consensus set file for writing: %v", err)
		writeError(w, msg, "")
		return
	}
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	err = out.Close()
	if err != nil {
		msg := fmt.Sprintf("Failed to close the consensus set file after writing: %v", err)
		writeError(w, msg, "")
		return
	}
	err = file.Close()
	if err != nil {
		msg := fmt.Sprintf("Failed to close the consensus set upload stream after writing: %v", err)
		writeError(w, msg, "")
		return
	}
	bootstrapper.Skip()
	time.Sleep(50 * time.Millisecond)
	consensusbuilder.Initialize()
	initializingNodeHandler(w, req, nil)
}

func expandMenuHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		redirect(w, req, nil)
		return
	}
	expandMenu(sessionID)
	writeHTML(w, getCachedPage(sessionID), sessionID)
}

func collapseMenuHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		redirect(w, req, nil)
		return
	}
	collapseMenu(sessionID)
	writeHTML(w, getCachedPage(sessionID), sessionID)
}

func scanningHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		redirect(w, req, nil)
		return
	}
	_, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		writeError(w, msg, sessionID)
		return
	}
	height, _, _ := blockHeightHelper(sessionID)
	if height == "0" && status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionID)
		return
	}
	if status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionID)
		return
	}
	guiHandler(w, req, nil)
}

func setTxHistoyPage(w http.ResponseWriter, req *http.Request, resp httprouter.Params) {
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		redirect(w, req, nil)
		return
	}
	page, _ := strconv.Atoi(req.FormValue("page"))
	setTxHistoryPage(page, sessionID)
	guiHandler(w, req, nil)
}

func guiHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	for i := 0; i < 10; i++ {
		if n.TransactionPool != nil {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if n.TransactionPool == nil {
		writeStaticHTML(w, resources.StartingWalletForm(), "")
		return
	}
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		writeStaticHTML(w, resources.InitializeWalletForm(), "")
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		writeStaticHTML(w, resources.InitializeWalletForm(), "")
		return
	}
	height, _, _ := blockHeightHelper(sessionID)
	if height == "0" && status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionID)
		return
	}
	if status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionID)
		return
	}
	unlocked, err := wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("Unable to determine if wallet is unlocked: %v", err)
		writeError(w, msg, sessionID)
		return
	}
	if unlocked {
		writeWallet(w, wallet, sessionID)
		return
	}
	closeWallet(sessionID)
	redirect(w, req, nil)
}

func writeWallet(w http.ResponseWriter, wallet modules.Wallet, sessionID string) {
	html := resources.WalletHTMLTemplate()
	html = strings.Replace(html, "&TRANSACTION_PORTAL;", resources.TransactionsHistoryHTMLTemplate(), -1)
	writeHTML(w, html, sessionID)
}

func writeArray(w http.ResponseWriter, arr []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encjson, _ := json.Marshal(arr)
	fmt.Fprint(w, string(encjson))
}

func writeError(w http.ResponseWriter, msg string, sessionID string) {
	html := resources.ErrorHTMLTemplate()
	html = strings.Replace(html, "&POPUP_TITLE;", "ERROR", -1)
	html = strings.Replace(html, "&POPUP_CONTENT;", msg, -1)
	html = strings.Replace(html, "&POPUP_CLOSE;", resources.CloseAlertForm(), -1)
	fmt.Println(msg)
	writeStaticHTML(w, html, sessionID)
}

func writeMsg(w http.ResponseWriter, title string, msg string, sessionID string) {
	html := resources.AlertHTMLTemplate()
	html = strings.Replace(html, "&POPUP_TITLE;", title, -1)
	html = strings.Replace(html, "&POPUP_CONTENT;", msg, -1)
	html = strings.Replace(html, "&POPUP_CLOSE;", resources.CloseAlertForm(), -1)
	writeHTML(w, html, sessionID)
}

func writeForm(w http.ResponseWriter, title string, form string, sessionID string) {
	html := resources.AlertHTMLTemplate()
	html = strings.Replace(html, "&POPUP_TITLE;", title, -1)
	html = strings.Replace(html, "&POPUP_CONTENT;", form, -1)
	html = strings.Replace(html, "&POPUP_CLOSE;", "", -1)
	writeHTML(w, html, sessionID)
}

func writeStaticHTML(w http.ResponseWriter, html string, sessionID string) {
	// add random data to links to act as a cache buster.
	// must be done last in case a cache buster is added in from a template.
	b := make([]byte, 16) //32 characters long
	rand.Read(b)
	cacheBuster := hex.EncodeToString(b)
	html = strings.Replace(html, "&CACHE_BUSTER;", cacheBuster, -1)
	html = strings.Replace(html, "&SESSION_ID;", sessionID, -1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func writeHTML(w http.ResponseWriter, html string, sessionID string) {
	if hasAlert(sessionID) {
		writeError(w, popAlert(sessionID), sessionID)
		return
	}
	cachedPage(html, sessionID)
	html = strings.Replace(html, "&WEB_WALLET_VERSION;", build.Version, -1)
	html = strings.Replace(html, "&SPD_VERSION;", spdBuild.Version, -1)
	session, _ := getSession(sessionID)
	if session != nil {
		html = strings.Replace(html, "&SESSION_NAME;", session.name, -1)
	}
	fmtHeight, fmtStatus, fmtStatCo := blockHeightHelper(sessionID)
	html = strings.Replace(html, "&STATUS_COLOR;", fmtStatCo, -1)
	html = strings.Replace(html, "&STATUS;", fmtStatus, -1)
	html = strings.Replace(html, "&BLOCK_HEIGHT;", fmtHeight, -1)
	fmtScpBal, fmtUncBal, fmtSpfaBal, fmtSpfbBal, fmtClmBal, fmtWhale := balancesHelper(sessionID)
	html = strings.Replace(html, "&SCP_BALANCE;", fmtScpBal, -1)
	html = strings.Replace(html, "&UNCONFIRMED_DELTA;", fmtUncBal, -1)
	html = strings.Replace(html, "&SPFA_BALANCE;", fmtSpfaBal, -1)
	html = strings.Replace(html, "&SPFB_BALANCE;", fmtSpfbBal, -1)
	html = strings.Replace(html, "&SCP_CLAIM_BALANCE;", fmtClmBal, -1)
	html = strings.Replace(html, "&WHALE_SIZE;", fmtWhale, -1)
	if menuIsCollapsed(sessionID) {
		html = strings.Replace(html, "&MENU;", resources.CollapsedMenuForm(), -1)
	} else {
		html = strings.Replace(html, "&MENU;", resources.ExpandedMenuForm(), -1)
	}
	writeStaticHTML(w, html, sessionID)
}

func whaleHelper(scpBal float64) string {
	if scpBal < 50 {
		return "🦐"
	}
	if scpBal < 100 {
		return "🐟"
	}
	if scpBal < 1000 {
		return "🦀"
	}
	if scpBal < 5000 {
		return "🐢"
	}
	if scpBal < 10000 {
		return "⚔️🐠"
	}
	if scpBal < 25000 {
		return "🐬"
	}
	if scpBal < 50000 {
		return "🦈"
	}
	if scpBal < 100000 {
		return "🌊🦄"
	}
	if scpBal < 250000 {
		return "🌊🐫"
	}
	if scpBal < 500000 {
		return "🐋"
	}
	if scpBal < 1000000 {
		return "🐙"
	}
	return "🐳"
}

func balancesHelper(sessionID string) (string, string, string, string, string, string) {
	fmtScpBal := "?"
	fmtUncBal := "?"
	fmtSpfaBal := "?"
	fmtSpfbBal := "?"
	fmtClmBal := "?"
	fmtWhale := "?"
	wallet, _ := getWallet(sessionID)
	if wallet == nil {
		return fmtScpBal, fmtUncBal, fmtSpfaBal, fmtSpfbBal, fmtClmBal, fmtWhale
	}
	unlocked, err := wallet.Unlocked()
	if err != nil {
		fmt.Printf("Unable to determine if wallet is unlocked: %v", err)
	}
	if unlocked {
		allBals, err := wallet.ConfirmedBalance()
		if err != nil {
			fmt.Printf("Unable to obtain confirmed balance: %v", err)
		} else {
			scpBal := allBals.CoinBalance
			fundABal := allBals.FundBalance
			fundBBal := allBals.FundbBalance
			// fundbBBal := allBals.FundbBalance
			claimBal := allBals.ClaimBalance
			// claimbBal := allBals.ClaimbBalance
			scpBalFloat, _ := new(big.Rat).SetFrac(scpBal.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			scpClaimBalFloat, _ := new(big.Rat).SetFrac(claimBal.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			fmtScpBal = fmt.Sprintf("%15.4f", scpBalFloat)
			fmtSpfaBal = fmt.Sprintf("%s", fundABal)
			fmtSpfbBal = fmt.Sprintf("%s", fundBBal)
			fmtClmBal = fmt.Sprintf("%15.4f", scpClaimBalFloat)
			fmtWhale = whaleHelper(scpBalFloat)

		}
		scpOut, scpIn, err := wallet.UnconfirmedBalance()
		if err != nil {
			fmt.Printf("Unable to obtain unconfirmed balance: %v", err)
		} else {
			scpInFloat, _ := new(big.Rat).SetFrac(scpIn.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			scpOutFloat, _ := new(big.Rat).SetFrac(scpOut.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			fmtUncBal = fmt.Sprintf("%15.4f", (scpInFloat - scpOutFloat))
		}
	}
	return fmtScpBal, fmtUncBal, fmtSpfaBal, fmtSpfbBal, fmtClmBal, fmtWhale
}

func blockHeightHelper(sessionID string) (string, string, string) {
	fmtHeight := "?"
	wallet, _ := getWallet(sessionID)
	if wallet == nil {
		return fmtHeight, "Offline", "red"
	}
	height, err := wallet.Height()
	if err != nil {
		fmt.Printf("Unable to obtain block height: %v", err)
	} else {
		fmtHeight = fmt.Sprintf("%d", height)
	}
	if status != "" {
		return fmtHeight, status, "yellow"
	}
	rescanning, err := wallet.Rescanning()
	if err != nil {
		fmt.Printf("Unable to determine if wallet is being scanned: %v", err)
	}
	if rescanning {
		return fmtHeight, "Rescanning", "cyan"
	}
	synced := n.ConsensusSet.Synced()
	if synced {
		return fmtHeight, "Synchronized", "blue"
	}
	return fmtHeight, "Synchronizing", "yellow"
}

func initializeSeedHelper(newPassword string, sessionID string) {
	setStatus("Initializing")
	msgPrefix := "Unable to initialize new wallet seed: "
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		setAlert(msg, sessionID)
		if status == "Initializing" {
			status = ""
		}
		return
	}
	var encryptionKey crypto.CipherKey = crypto.NewWalletKey(crypto.HashObject(newPassword))
	_, err = wallet.Encrypt(encryptionKey)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		setAlert(msg, sessionID)
		if status == "Initializing" {
			status = ""
		}
		return
	}
	potentialKeys, _ := encryptionKeys(newPassword)
	for _, key := range potentialKeys {
		unlocked, err := wallet.Unlocked()
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			setAlert(msg, sessionID)
			if status == "Initializing" {
				status = ""
			}
			return
		}
		if !unlocked {
			wallet.Unlock(key)
		}
	}
	setStatus("")
}

func isPasswordValid(wallet modules.Wallet, password string) (bool, error) {
	keys, _ := encryptionKeys(password)
	var err error
	for _, key := range keys {
		valid, keyErr := wallet.IsMasterKey(key)
		if keyErr == nil {
			if valid {
				return true, nil
			}
			return false, nil
		}
		err = nebErrors.Compose(err, keyErr)
	}
	return false, err
}

func restoreSeedHelper(newPassword string, seed modules.Seed, sessionID string) {
	setStatus("Restoring")
	for !n.ConsensusSet.Synced() {
		time.Sleep(25 * time.Millisecond)
	}
	msgPrefix := "Unable to restore new wallet seed: "
	wallet, err := getWallet(sessionID)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		setAlert(msg, sessionID)
		if status == "Restoring" {
			status = ""
		}
		return
	}
	var encryptionKey crypto.CipherKey = crypto.NewWalletKey(crypto.HashObject(newPassword))
	err = wallet.InitFromSeed(encryptionKey, seed)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		setAlert(msg, sessionID)
		if status == "Restoring" {
			status = ""
		}
		return
	}
	potentialKeys, _ := encryptionKeys(newPassword)
	for _, key := range potentialKeys {
		unlocked, err := wallet.Unlocked()
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			setAlert(msg, sessionID)
			if status == "Restoring" {
				status = ""
			}
			return
		}
		if !unlocked {
			wallet.Unlock(key)
		}
	}
	setStatus("")
}

func transactionExplorerHelper(txn modules.ProcessedTransaction) (string, error) {
	unixTime, _ := strconv.ParseInt(fmt.Sprintf("%v", txn.ConfirmationTimestamp), 10, 64)
	fmtTime := strings.ToUpper(time.Unix(unixTime, 0).Format("2006-01-02 15:04"))
	fmtTxnID := strings.ToUpper(fmt.Sprintf("%v", txn.TransactionID))
	fmtTxnType := strings.ToUpper(strings.Replace(fmt.Sprintf("%v", txn.TxType), "_", " ", -1))
	fmtTxnBlock := strings.ToUpper(fmt.Sprintf("%v", txn.ConfirmationHeight))
	html := resources.TransactionInfoTemplate()
	html = strings.Replace(html, "&TXN_TYPE;", fmtTxnType, -1)
	html = strings.Replace(html, "&TXN_ID;", fmtTxnID, -1)
	html = strings.Replace(html, "&TXN_TIME;", fmtTime, -1)
	html = strings.Replace(html, "&TXN_BLOCK;", fmtTxnBlock, -1)
	inputs := ""
	for _, input := range txn.Inputs {
		fmtValue := strings.ToUpper(fmt.Sprintf("%v", input.Value))
		fmtAddress := strings.ToUpper(fmt.Sprintf("%v", input.RelatedAddress))
		fmtFundType := strings.ToUpper(strings.Replace(fmt.Sprintf("%v", input.FundType), "_", " ", -1))
		fmtFundType = strings.Replace(fmtFundType, "SIACOIN", "SCP", -1)
		fmtFundType = strings.Replace(fmtFundType, "SIAFUND", "SPF", -1)
		row := resources.TransactionInputTemplate()
		row = strings.Replace(row, "&VALUE;", fmtValue, -1)
		row = strings.Replace(row, "&ADDRESS;", fmtAddress, -1)
		row = strings.Replace(row, "&FUND_TYPE;", fmtFundType, -1)
		inputs = inputs + row
	}
	html = strings.Replace(html, "&TXN_INPUTS;", inputs, -1)
	outputs := ""
	for _, output := range txn.Outputs {
		fmtValue := strings.ToUpper(fmt.Sprintf("%v", output.Value))
		fmtAddress := strings.ToUpper(fmt.Sprintf("%v", output.RelatedAddress))
		fmtFundType := strings.ToUpper(strings.Replace(fmt.Sprintf("%v", output.FundType), "_", " ", -1))
		fmtFundType = strings.Replace(fmtFundType, "SIACOIN", "SCP", -1)
		fmtFundType = strings.Replace(fmtFundType, "SIAFUND", "SPF", -1)
		row := resources.TransactionOutputTemplate()
		row = strings.Replace(row, "&VALUE;", fmtValue, -1)
		row = strings.Replace(row, "&ADDRESS;", fmtAddress, -1)
		row = strings.Replace(row, "&FUND_TYPE;", fmtFundType, -1)
		outputs = outputs + row
	}
	html = strings.Replace(html, "&TXN_OUTPUTS;", outputs, -1)
	return html, nil
}

type TransactionHistoryPage struct {
	TransactionHistoryLines []TransactionHistoryLine `json:"lines"`
	Current                 int                      `json:"current"`
	Total                   int                      `json:"total"`
}

type TransactionHistoryLine struct {
	TransactionID      string `json:"transaction_id"`
	ShortTransactionID string `json:"short_transaction_id"`
	Type               string `json:"type"`
	Time               string `json:"time"`
	Amount             string `json:"amount"`
	Fee                string `json:"fee"`
	Confirmed          string `json:"confirmed"`
}

func transactionHistoryJson(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	msgPrefix := "Unable to generate transaction history: "
	sessionID := req.FormValue("session_id")
	if sessionID == "" || !sessionIDExists(sessionID) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msgPrefix + "no session ID"))
		return
	}
	if !sessionIDExists(sessionID) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msgPrefix + "invalid session ID"))
		return
	}
	wallet, err := getWallet(sessionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(msgPrefix+"%v", err)))
		return
	}
	page := getTxHistoryPage(sessionID)
	pageSize := 20
	pageMin := (page - 1) * pageSize
	pageMax := page * pageSize
	count := 0
	heightMin := 0
	confirmedTxns, err := wallet.Transactions(types.BlockHeight(heightMin), n.ConsensusSet.Height())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(msgPrefix+"%v", err)))
		return
	}
	unconfirmedTxns, err := wallet.UnconfirmedTransactions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(msgPrefix+"%v", err)))
		return
	}
	sts, err := ComputeSummarizedTransactions(append(confirmedTxns, unconfirmedTxns...), n.ConsensusSet.Height(), wallet)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(msgPrefix+"%v", err)))
		return
	}
	lines := []TransactionHistoryLine{}
	// iterate in reverse
	for i := len(sts) - 1; i >= 0; i-- {
		txn := sts[i]
		// Format transaction type
		isSetup := txn.Type == "SETUP" && txn.Scp == 0 && txn.SpfA == 0 && txn.SpfB == 0
		if !isSetup {
			count++
			if count >= pageMin && count < pageMax {
				var amountArr []string
				if txn.Scp != 0 {
					amountArr = append(amountArr, strings.TrimRight(strings.TrimRight(fmt.Sprintf("%15.4f", txn.Scp), "0"), ".")+" SCP")
				}
				if txn.SpfA != 0 {
					postfix := "SPF-A"
					if txn.Confirmed == _stUnconfirmedStr { // in case of unconfirmed we just show SPF
						postfix = "SPF"
					}
					amountArr = append(amountArr, fmt.Sprintf("%14v %s", txn.SpfA, postfix))
				}
				if txn.SpfB != 0 {
					amountArr = append(amountArr, fmt.Sprintf("%14v SPF-B", txn.SpfB))
				}
				fmtAmount := strings.Join(amountArr, "; ")
				if fmtAmount == "" {
					fmtAmount = "0 SCP/SPF"
				}
				var fmtFee string
				if txn.ScpFee != 0 {
					fmtFee = fmt.Sprintf("%f SCP fee", txn.ScpFee)
				}
				line := TransactionHistoryLine{}
				line.TransactionID = txn.TxnID
				line.ShortTransactionID = txn.TxnID[0:16] + "..." + txn.TxnID[len(txn.TxnID)-16:]
				line.Type = txn.Type
				line.Time = txn.Time
				line.Amount = fmtAmount
				line.Fee = fmtFee
				line.Confirmed = txn.Confirmed
				lines = append(lines, line)
			}
		}
	}
	txHistoryPage := TransactionHistoryPage{}
	txHistoryPage.TransactionHistoryLines = lines
	txHistoryPage.Current = page
	txHistoryPage.Total = (count / pageSize) + 1
	json, err := json.Marshal(txHistoryPage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(msgPrefix+"%v", err)))
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(json))) //len(dec)
	w.Write(json)
}

// scanAddress scans a types.UnlockHash from a string.
// copied from "gitlab.com/scpcorp/ScPrime/node/scan.go"
func scanAddress(addrStr string) (addr types.UnlockHash, err error) {
	err = addr.LoadString(addrStr)
	if err != nil {
		return types.UnlockHash{}, err
	}
	return addr, nil
}

// encryptionKeys enumerates the possible encryption keys that can be derived
// from an input string.
// copied from "gitlab.com/scpcorp/ScPrime/node/wallet.go"
func encryptionKeys(seedStr string) (validKeys []crypto.CipherKey, seeds []modules.Seed) {
	dicts := []mnemonics.DictionaryID{"english", "german", "japanese"}
	for _, dict := range dicts {
		seed, err := modules.StringToSeed(seedStr, dict)
		if err != nil {
			continue
		}
		validKeys = append(validKeys, crypto.NewWalletKey(crypto.HashObject(seed)))
		seeds = append(seeds, seed)
	}
	validKeys = append(validKeys, crypto.NewWalletKey(crypto.HashObject(seedStr)))
	return
}
