# [![ScPrime WebWallet][ScPrime Logo]][ScPrime]


[![Latest Release][Latest Release Badge]][Latest Releases]
[![Build Status][Build Status Badge]][Commit History]
[![GoDoc][GoDoc Badge]][GoDoc SCP Corp Web Wallet]
[![Go Report Card][Go Report Card Badge]][Go Report Card SCP Corp Web Wallet]
[![License MIT][License Badge]][License Details]

ScPrime has released a browser based GUI (Graphical User Interface) wallet called the WebWallet, that is purely for storing your SCP and doesn't have the 'hosting' aspect that the CLI or UI software has. If you're just looking for software to just send, receive, and hold your SCP, this is a good option for you.

Usage
-----

An online walk through of the web wallet is available in the [ScPrime Documents repository][].

Environment Variables
---------------------

You can configure the web wallet to persist and retrieve application data to a specific directory by setting the `SCPRIME_WEB_WALLET_DATA_DIR` environment variable to the desired directory path. If this environment variable is not set then a default directory will be determined according to your operating system as follows:
  * Linux:   `$HOME/.scprime-webwallet`
  * MacOS:   `$HOME/Library/Application Support/ScPrime-WebWallet`
  * Windows: `%LOCALAPPDATA%\ScPrime-WebWallet`

Building From Source
--------------------

To build from source, [Go 1.17 or above][] must be installed on the system. Then clone the repo and run make. Example:

```sh
git clone https://gitlab.com/scpcorp/webwallet
cd webwallet && make
```

This will install the `scp-webwallet` and `scp-webwallet-server` binaries into your $GOPATH/bin folder (By default, this is $HOME/go/bin).

Other Makefile commands are:
* `make all`, another way to build and install the release binaries
* `make fmt`, uses go fmt to format all golang files
* `make vet`, uses go vet to analyze all golang files for suspicious, abnormal, or useless code
* `make lint`, lints all golang files with the linters defined in `.golangci.yml`
* `make debug`, builds and installs the debug binary
* `make dev`, builds and installs the developer binary
* `make release`, builds and installs the release binary
* `make clean`, deletes and cruft from this code repository
* `make test`, runs the test suite
* `make code`, generates code coverage reports and saves them to this project's cover folder

Building Release Binaries
-------------------------

To build the release binaries from source; zip, sha1sum, and [Go 1.17 or above][] must be installed on the system. Then clone the repo and run the release script. Example:

```sh
git clone https://gitlab.com/scpcorp/webwallet
cd webwallet && ./release-scripts/release.sh v0.0.0
cd release
```

This will save the `scp-webwallet` release binaries to the webwallet's `./release` directory.

Building Signed Release Binaries
--------------------------------

To build signed release binaries; zip, gpg, sha1sum, and [Go 1.17 or above][] must be installed on the system. Then clone the repo and run the release script. Example:
```sh
git clone https://gitlab.com/scpcorp/webwallet
cd webwallet && ./release-scripts/release.sh v0.0.0 keyfile
cd release
```

This will save the signed `scp-webwallet` release binaries to the webwallet's `./release` directory.

[ScPrime]: https://scpri.me
[ScPrime Logo]: https://scpri.me/imagestore/SPRho_256x256.png
[Latest Release Badge]: https://gitlab.com/scpcorp/webwallet/-/badges/release.svg
[Latest Releases]: https://gitlab.com/scpcorp/webwallet/-/releases
[Build Status Badge]: https://gitlab.com/scpcorp/webwallet/badges/main/pipeline.svg
[Commit History]: https://gitlab.com/scpcorp/webwallet/commits/main
[GoDoc Badge]: https://godoc.org/gitlab.com/scpcorp/webwallet?status.svg
[GoDoc SCP Corp Web Wallet]: https://godoc.org/gitlab.com/scpcorp/webwallet
[Go Report Card Badge]: https://goreportcard.com/badge/gitlab.com/scpcorp/webwallet
[Go Report Card SCP Corp Web Wallet]: https://goreportcard.com/report/gitlab.com/scpcorp/webwallet
[License Badge]: https://img.shields.io/badge/License-MIT-brightgreen.svg
[License Details]: https://gitlab.com/scpcorp/webwallet/-/blob/main/LICENSE
[ScPrime Documents repository]: https://docs.scpri.me/software/webwallet
[Go 1.17 or above]: https://golang.org/doc/install


