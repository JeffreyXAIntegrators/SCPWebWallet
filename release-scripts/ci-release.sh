#!/usr/bin/env bash
set -e

# version and keys are supplied as arguments
version=${1}
rc=${version}
passphrase=$keypw
keyfile=$keyfile

rc=`echo ${version} | awk -F - '{print $2}'`
echo $keyfile | base64 -d | gpg --batch --passphrase $passphrase --import

# setup build-time vars
sharedldflags="-s -w -X 'gitlab.com/scpcorp/webwallet/build.GitRevision=`git rev-parse --short HEAD`' -X 'gitlab.com/scpcorp/webwallet/build.BuildTime=`git show -s --format=%ci HEAD`' -X 'gitlab.com/scpcorp/webwallet/build.ReleaseTag=${rc}'"

function build {
  os=${1}
  arch=${2}
  pkg=${3}
  releasePkg=${pkg}-${version}-${os}-${arch}
  echo Building ${releasePkg}...
  # set binary name
  bin=${pkg}
  if [ ${os} == "windows" ]; then
    bin=${bin}.exe
  elif [ ${os} == "darwin" ] && [ ${pkg} == "scp-webwallet" ]; then
    bin=${bin}.app
  fi
  # set folder
  folder=release/${releasePkg}
  rm -rf ${folder}
  mkdir -p ${folder}
  # set path to bin
  binpath=${folder}/${bin}
  if [ ${os} == "darwin" ] && [ ${pkg} == "scp-webwallet" ]; then
    # Appify the scp-webwallet darwin release. More documentation at `./release-scripts/app_resources/darwin/scp-webwallet/README.md`.
    cp -a ./release-scripts/app_resources/darwin/${pkg}/${bin} ${binpath}
    # touch the scp-webwallet.app container to reset the time created timestamp
    touch ${binpath}
    binpath=${binpath}/Contents/MacOS/${pkg}
  fi
  # set ldflags
  ldflags=$sharedldflags
  if [ ${os} == "windows" ] && [ ${pkg} == "scp-webwallet" ]; then
    # Appify the scp-webwallet windows release. More documentation at `./release-scripts/app_resources/windows/scp-webwallet/README.md`
    cp ./release-scripts/app_resources/windows/${pkg}/rsrc_windows_386.syso ./cmd/${pkg}/rsrc_windows_386.syso
    cp ./release-scripts/app_resources/windows/${pkg}/rsrc_windows_amd64.syso ./cmd/${pkg}/rsrc_windows_amd64.syso
    # on windows build an application binary instead of a command line binary.
    ldflags="${sharedldflags} -H windowsgui"
  fi
  GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 go build -a -tags 'netgo' -trimpath -ldflags="${ldflags}" -o ${binpath} ./cmd/${pkg}
  # Set the user permissions on the binary
  chmod 755 ${binpath}
  # Cleanup scp-webwallet windows release.
  if [ ${os} == "windows" ] && [ ${pkg} == "scp-webwallet" ]; then
    rm ./cmd/${pkg}/rsrc_windows_386.syso
    rm ./cmd/${pkg}/rsrc_windows_amd64.syso
  fi
  if ! [[ -z ${keyfile} ]]; then
    echo $passphrase | PASSPHRASE=$passphrase gpg --batch --pinentry-mode loopback --command-fd 0 --armour --output  ${folder}/${bin}.asc --detach-sig ${binpath}
  fi
  sha1sum ${binpath} >> release/${pkg}-${version}-SHA1SUMS.txt
  cp -r LICENSE README.md ${folder}  
  (
    cd release/
    zip -rq ${releasePkg}.zip ${releasePkg}
  )
  if ! [[ -z ${keyfile} ]]; then
    echo $passphrase | PASSPHRASE=$passphrase gpg --batch --pinentry-mode loopback --command-fd 0 --armour --output  ${folder}.zip.asc --detach-sig ${folder}.zip
  fi
}

# Build the wallet.wasm resource
echo Building wallet.wasm
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./resources/resources/wasm_exec.js
GOOS=js GOARCH=wasm go build -o ./resources/resources/wallet.wasm ./cmd/wasm/main.go

# Build the cold wallet
mkdir -p release
go run cmd/scp-cold-wallet/main.go > release/scp-cold-wallet.html

# Build the packages
for pkg in scp-webwallet scp-webwallet-server; do
  # Build amd64 binaries.
  for os in darwin linux windows; do
    build ${os} "amd64" ${pkg}
  done
  # Build arm64 binaries.
  for os in darwin linux; do
    build ${os} "arm64" ${pkg}
  done
done
