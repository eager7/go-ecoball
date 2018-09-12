package main

import "github.com/ecoball/go-ecoball/sharding"

func main() {
	instance := sharding.MakeSharding()
	instance.Start()

	select {}
}
