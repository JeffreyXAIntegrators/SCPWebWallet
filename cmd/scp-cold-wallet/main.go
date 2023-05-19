package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"gitlab.com/scpcorp/webwallet/resources"
)

// main starts the daemon.
func main() {
	style := resources.ColdWalletCSSStyleSheet()
	script := resources.Javascript()
	logo := resources.Logo()
	wasmExec := resources.WasmExec()
	walletWasm := resources.WalletWasm()
	html := resources.ColdWalletHTML()
	html = strings.Replace(html, "&STYLE;", string(style), -1)
	html = strings.Replace(html, "&SCRIPT;", string(script), -1)
	html = strings.Replace(html, "&LOGO;", base64.StdEncoding.EncodeToString(logo), -1)
	html = strings.Replace(html, "&REGENERATE;", "<button onClick='window.location.reload();'>Regenerate</button>", -1)
	html = strings.Replace(html, "&CLOSE;", "", -1)
	html = strings.Replace(html, "&WASM_EXEC;", string(wasmExec), -1)
	html = strings.Replace(html, "&WALLET_WASM;", base64.StdEncoding.EncodeToString(walletWasm), -1)
	fmt.Println(html)
}
