package net_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/lib-p2p/net"
	"github.com/hashicorp/golang-lru"
	"testing"
)

func TestNewMessage(t *testing.T) {
	csc, _ := lru.New(1000000)
	var num int
	{
		m := net.RandomUint64()
		if csc.Contains(m) {
			t.Fatal(num)
		}
		csc.Add(m, m)
		num++
		fmt.Println(num)
	}
}
