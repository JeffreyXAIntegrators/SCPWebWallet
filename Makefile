# These variables get inserted into ./build/commit.go
BUILD_TIME=$(shell date)
GIT_REVISION=$(shell git rev-parse --short HEAD)
GIT_DIRTY=$(shell git diff-index --quiet HEAD -- || echo "modified-")

ldflags= -X gitlab.com/scpcorp/webwallet/build.GitRevision=${GIT_DIRTY}${GIT_REVISION} \
-X "gitlab.com/scpcorp/webwallet/build.BuildTime=${BUILD_TIME}"

racevars= history_size=3 halt_on_error=1 atexit_sleep_ms=2000

# all will build and install release binaries
all: release

# count says how many times to run the tests.
count = 1

# pkgs changes which packages the makefile calls operate on. run changes which
# tests are run during testing.
pkgs = \
	./cmd/scp-webwallet \
	./daemon \
	./resources \
	./server

# release-pkgs determine which packages are built for release and distrubtion
# when running a 'make release' command.
release-pkgs = \
	./cmd/scp-webwallet \
	./cmd/scp-webwallet-server

# run determines which tests run when running any variation of 'make test'.
run = .

# fmt calls go fmt on all packages.
fmt:
	gofmt -s -l -w $(pkgs)

# vet calls go vet on all packages.
# NOTE: go vet requires packages to be built in order to obtain type info.
vet:
	go vet $(pkgs)

# lint runs golangci-lint.
lint:
	golangci-lint run -c .golangci.yml ./...

# debug builds and installs debug binaries.
debug:
	go install -tags='debug profile netgo' -ldflags='$(ldflags)' $(release-pkgs)

# dev builds and installs developer binaries.
dev:
	go install -tags='dev debug profile netgo' -ldflags='$(ldflags)' $(release-pkgs)

# release builds and installs release binaries.
release:
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./resources/resources/wasm_exec.js \
	&& GOOS=js GOARCH=wasm go build -o ./resources/resources/wallet.wasm ./cmd/wasm/main.go \
  && go install -tags='netgo' -ldflags='-s -w $(ldflags)' $(release-pkgs)

# clean removes all directories that get automatically created during
# development.
clean:
ifneq ("$(OS)","Windows_NT")
# Linux
	rm -rf cover release
else
# Windows
	- DEL /F /Q cover release
endif

test:
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./resources/resources/wasm_exec.js \
	&& GOOS=js GOARCH=wasm go build -o ./resources/resources/wallet.wasm ./cmd/wasm/main.go \
	&& go test -short -tags='debug testing netgo' -timeout=60s $(pkgs) -run=$(run) -count=$(count)
cover: clean
	@mkdir -p cover
	@for package in $(pkgs); do                                                                                                                                 \
		mkdir -p `dirname cover/$$package`                                                                                                                      \
		&& go test -tags='testing debug netgo' -timeout=500s -covermode=atomic -coverprofile=cover/$$package.out ./$$package -run=$(run) || true \
		&& go tool cover -html=cover/$$package.out -o=cover/$$package.html ;                                                                                    \
	done
.PHONY: all fmt vet lint release clean test cover
