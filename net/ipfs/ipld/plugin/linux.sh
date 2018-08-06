#!/bin/bash
set -e

# Change
echo "ipldecoball github.com/ecoball/go-ecoball/net/ipfs/ipld/plugin 0" >> $GOPATH/src/github.com/ipfs/go-ipfs/plugin/loader/preload_list
sed -i  's/package main/package plugin/' $GOPATH/src/github.com/ecoball/go-ecoball/net/ipfs/ipld/plugin/ecoball.go
gx-go rw

# Build and Execute
cd $GOPATH/src/github.com/ipfs/go-ipfs/
make build
./cmd/ipfs/ipfs daemon

# Revert
git checkout -- $GOPATH/src/github.com/ipfs/go-ipfs/plugin/loader/preload_list
git checkout -- $GOPATH/src/github.com/ipfs/go-ipfs/plugin/loader/preload.go
cd $GOPATH/src/github.com/ecoball/go-ecoball/net/ipfs/ipld
gx-go uw
sed -i  's/package plugin/package main/' $GOPATH/src/github.com/ecoball/go-ecoball/net/ipfs/ipld/plugin/ecoball.go
