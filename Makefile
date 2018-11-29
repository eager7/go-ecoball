# Copyright QuakerChain All Rights Reserved.

include common/config/config

BASE_VERSION = 1.1.1

all: ecoball ecoclient ecowallet plugins tools

.PHONY: proto plugins ecoball ecoclient ecowallet tools
ecoball: plugins
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
	make -C client/protos
	make -C net/message/pb
	make -C net/gossip/protos

plugins:
	@echo "\033[;32mbuild ipld plugin file \033[0m"
	mkdir -p build/storage/plugins
	make -C dsn/host/ipfs/plugin
	chmod +x dsn/host/ipfs/plugin/ecoball.so
	mv dsn/host/ipfs/plugin/ecoball.so build/storage/plugins
tools:
	@echo "\033[;32mbuild tools \033[0m"
	mkdir -p tools/c2wasm/build
	go build -v -o c2wasm tools/c2wasm/src/main.go
	mv c2wasm tools/c2wasm/build

.PHONY: clean
clean:
	@echo "\033[;31mclean project \033[0m"
	-rm -rf build/
	-rm ./build/ecoball.toml
	#make -C core/pb/ clean
	make -C client/protos clean
	make -C  net/message/pb/ clean
    make -C  net/gossip/protos/ clean


.PHONY: test

test:
	@echo "\033[;31mhello world \033[0m"
	@echo $(PATH)
	@echo $(GOPATH)
	@echo $(GOBIN)
