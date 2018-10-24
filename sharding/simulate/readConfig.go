package simulate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type NodeConfig struct {
	Pubkey  string
	Address string
	Port    string
}

type config struct {
	Pubkey    string
	Address   string
	Port      string
	Committee []NodeConfig
	Shard     []NodeConfig
}

func readConfigFile() *config {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	file := "config.json"

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Info("read config file error")
		return nil
	}

	str := string(bytes)

	var c config
	if err := json.Unmarshal([]byte(str), &c); err != nil {
		log.Info("json unmarshal error")
		return nil
	}

	return &c
}

var cfg *config

func LoadConfig() {
	cfg = readConfigFile()
	if cfg == nil {
		panic("read config error")
		return
	}
}

func GetNodeInfo() (self NodeConfig) {
	self.Pubkey = cfg.Pubkey
	self.Address = cfg.Address
	self.Port = cfg.Port

	return
}

func GetCommittee() []NodeConfig {
	return cfg.Committee
}

func GetShards() []NodeConfig {
	return cfg.Shard
}

func GetNodePubKey() []byte {
	return []byte(cfg.Pubkey)
}
