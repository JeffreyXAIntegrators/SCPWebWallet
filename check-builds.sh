#!/usr/bin/env bash

echo "check-builds is used to check that the Web Wallet can be compiled on all supported systems"

# Create fresh artifacts folder.
rm -rf artifacts
mkdir artifacts

# Return first error encountered by any command.
set -e

# Build binaries and sign them.
function build {
  os=${1}
  arch=${2}
  pkg=${3}
  echo Building ${pkg}-${os}-${arch}...
  # set binary name
  bin=${pkg}
  # Different naming convention for windows.
  if [ "$os" == "windows" ]; then
    bin=${pkg}.exe
  fi
  # Build binary.
  GOOS=${os} GOARCH=${arch} go build -tags='netgo' -o artifacts/$arch/$os/$bin ./cmd/$pkg
}

# Build the wallet.wasm resource
echo Building wallet.wasm
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./resources/resources/wasm_exec.js
GOOS=js GOARCH=wasm go build -o ./resources/resources/wallet.wasm ./cmd/wasm/main.go

# Build amd64 binaries.
for os in darwin linux windows; do
  build ${os} "amd64" "scp-webwallet"
done

# Build arm64 binaries.
for os in darwin linux; do
  build ${os} "arm64" "scp-webwallet"
done

