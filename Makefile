# Copyright QuakerChain All Rights Reserved.

BASE_VERSION = 1.1.1

all: ecoball ecoclient ecowallet proto plugins

.PHONY: proto plugins ecoball ecoclient ecowallet
ecoball: proto plugins
	@echo "\033[;32mbuild ecoball \033[0m"
	mkdir -p build/
	go build -v -o ecoball node/*.go
	mv ecoball build/

ecoclient: 
	@echo "\033[;32mbuild ecoclient \033[0m"
	mkdir -p build/
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

plugins:
	@echo "\033[;32mbuild ipld plugin file \033[0m"
	mkdir -p build/plugins
	make -C net/ipfs/ipld/plugin
	chmod +x net/ipfs/ipld/plugin/ecoball.so
	mv net/ipfs/ipld/plugin/ecoball.so build/plugins

.PHONY: clean
clean:
	@echo "\033[;31mclean project \033[0m"
	-rm -rf build/
	-rm ~/ecoball.toml
	make -C core/pb/ clean
	make -C client/protos clean

.PHONY: test
test:
	@echo "\033[;31mhello world \033[0m"
