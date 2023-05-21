package main

import (
	"fmt"

	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
	"gitlab.com/scpcorp/webwallet/utils/wallet"
)

//nolint:typecheck
// the syscall/js import is required for the Golang code to interact with the
// DOM and compiles correctly as long as GOARCH=wasm GOOS=js is used.
import (
	"syscall/js"
)

// WasmNewWalletSeed returns a new random wallet seed
// requires exactly zero parameters
func WasmNewWalletSeed() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 0 {
			fmt.Println("invalid number of arguments passed")
			return nil
		}
		seed, err := wallet.NewWalletSeed()
		if err != nil {
			fmt.Printf("unable to create new wallet seed: %s\n", err)
			return err.Error()
		}
		return seed
	})
	return jsonFunc
}

// WasmAddressFromSeed returns an address from the seed.
// requires exactly two parameters.
// first parameter must be the wallet seed supplied as a string
// second parameter must be the offset
func WasmAddressFromSeed() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 2 {
			fmt.Println("invalid number of arguments passed")
			return nil
		}
		seed, _ := wallet.StringToSeed(args[0].String(), mnemonics.DictionaryID("english"))
                offset := uint64(args[1].Int())
		key := wallet.GetAddress(seed, offset)
		return key.UnlockConditions.UnlockHash().String()
	})
	return jsonFunc
}

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("wasmAddressFromSeed", WasmAddressFromSeed())
	js.Global().Set("wasmNewWalletSeed", WasmNewWalletSeed())
	<-make(chan bool)
}
