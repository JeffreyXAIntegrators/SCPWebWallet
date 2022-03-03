# [![ScPrime WebWallet](https://scpri.me/imagestore/SPRho_256x256.png)](http://scpri.me)

[![Latest Release](https://gitlab.com/scpcorp/webwallet/-/badges/release.svg)](https://gitlab.com/scpcorp/webwallet/-/tags)
[![Build Status](https://gitlab.com/scpcorp/webwallet/badges/main/pipeline.svg)](https://gitlab.com/scpcorp/webwallet/commits/main)
[![GoDoc](https://godoc.org/gitlab.com/scpcorp/webwallet?status.svg)](https://godoc.org/gitlab.com/scpcorp/webwallet)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/scpcorp/webwallet)](https://goreportcard.com/report/gitlab.com/scpcorp/webwallet)
[![License MIT](https://img.shields.io/badge/License-MIT-brightgreen.svg)](https://gitlab.com/scpcorp/webwallet/-/blob/main/LICENSE)

ScPrime has released a browser based GUI (Graphical User Interface) wallet called the WebWallet, that is purely for storing your SCP and doesn't have the 'hosting' aspect that the CLI or UI software has. If you're just looking for software to just send, receive, and hold your SCP, this is a good option for you.

Usage
-----

Online documentation available at https://docs.scpri.me/software/webwallet

Building From Source
--------------------

To build from source, [Go 1.17 or above must be installed](https://golang.org/doc/install) on the system. Clone the repo and run make:

```sh
git clone https://gitlab.com/scpcorp/webwallet
cd webwallet && make
```

This will install the `scp-webwallet` binary in your $GOPATH/bin folder (By default, this is $HOME/go/bin).

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

