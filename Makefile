# Copyright QuakerChain All Rights Reserved.

include common/config/config

BASE_VERSION = 1.1.1

all: ecoball ecoclient ecowallet tools

.PHONY: proto ecoball ecoclient ecowallet tools
ecoball:
	@echo "\033[;32mbuild ecoball \033[0m"
	mkdir -p build/
	go install -v node/*.go
	go build -v -o ecoball node/*.go
	mv ecoball build/

ecoclient: 
	@echo "\033[;32mbuild ecoclient \033[0m"
	mkdir -p build/
	go install -v client/client.go
	go build -v -o ecoclient client/client.go
	mv ecoclient build/

ecowallet: 
	@echo "\033[;32mbuild ecowallet \033[0m"
	mkdir -p build/
	go build -v -o ecowallet walletserver/main.go 
	mv ecowallet build/

proto:
	@echo "\033[;32mbuild protobuf file \033[0m"
	make -C core/pb
	make -C common/message/mpb
	make -C client/protos
	make -C net/message/pb
	make -C net/gossip/protos

tools:
	@echo "\033[;32mbuild tools \033[0m"
	mkdir -p tools/c2wasm/build
	go build -v -o c2wasm tools/c2wasm/src/main.go
	mv c2wasm tools/c2wasm/build

.PHONY: clean
clean:
	@echo "\033[;31mclean project \033[0m"
	-rm -rf build/

.PHONY: test

test:
	@echo "\033[;31mhello world \033[0m"
	@echo $(PATH)
	@echo $(GOPATH)
	@echo $(GOBIN)
