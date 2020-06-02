# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gtrcn android ios gtrcn-cross evm all test clean
.PHONY: gtrcn-linux gtrcn-linux-386 gtrcn-linux-amd64 gtrcn-linux-mips64 gtrcn-linux-mips64le
.PHONY: gtrcn-linux-arm gtrcn-linux-arm-5 gtrcn-linux-arm-6 gtrcn-linux-arm-7 gtrcn-linux-arm64
.PHONY: gtrcn-darwin gtrcn-darwin-386 gtrcn-darwin-amd64
.PHONY: gtrcn-windows gtrcn-windows-386 gtrcn-windows-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

gtrcn:
	$(GORUN) build/ci.go install ./cmd/gtrcn
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gtrcn\" to launch gtrcn."

all:
	$(GORUN) build/ci.go install

android:
	$(GORUN) build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gtrcn.aar\" to use the library."

ios:
	$(GORUN) build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Gtrcn.framework\" to use the library."

test: all
	$(GORUN) build/ci.go test

lint: ## Run linters.
	$(GORUN) build/ci.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gtrcn-cross: gtrcn-linux gtrcn-darwin gtrcn-windows gtrcn-android gtrcn-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-*

gtrcn-linux: gtrcn-linux-386 gtrcn-linux-amd64 gtrcn-linux-arm gtrcn-linux-mips64 gtrcn-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-*

gtrcn-linux-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gtrcn
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep 386

gtrcn-linux-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gtrcn
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep amd64

gtrcn-linux-arm: gtrcn-linux-arm-5 gtrcn-linux-arm-6 gtrcn-linux-arm-7 gtrcn-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep arm

gtrcn-linux-arm-5:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gtrcn
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep arm-5

gtrcn-linux-arm-6:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gtrcn
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep arm-6

gtrcn-linux-arm-7:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gtrcn
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep arm-7

gtrcn-linux-arm64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gtrcn
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep arm64

gtrcn-linux-mips:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gtrcn
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep mips

gtrcn-linux-mipsle:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gtrcn
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep mipsle

gtrcn-linux-mips64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gtrcn
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep mips64

gtrcn-linux-mips64le:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gtrcn
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-linux-* | grep mips64le

gtrcn-darwin: gtrcn-darwin-386 gtrcn-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-darwin-*

gtrcn-darwin-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gtrcn
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-darwin-* | grep 386

gtrcn-darwin-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gtrcn
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-darwin-* | grep amd64

gtrcn-windows: gtrcn-windows-386 gtrcn-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-windows-*

gtrcn-windows-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gtrcn
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-windows-* | grep 386

gtrcn-windows-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gtrcn
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gtrcn-windows-* | grep amd64
