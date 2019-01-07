package simulate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	sc "github.com/ecoball/go-ecoball/sharding/common"
)

type config struct {
	Pubkey    string
	Address   string
	Port      string
	Size      string
	Committee []sc.Worker
	Shard     []sc.Worker
	Candidate []sc.Worker
}

func readConfigFile(path string) *config {
	//file := "sharding.json"
	fmt.Println("sharding config file:", path)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Info("read config file error")
		panic("sharding configure not exist")
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

func LoadConfig(path string) {
	cfg = readConfigFile(path)
	if cfg == nil {
		panic("read config error")
		return
	}
}

func GetNodeInfo() (self sc.Worker) {
	self.Pubkey = cfg.Pubkey
	self.Address = cfg.Address
	self.Port = cfg.Port

	return
}

func GetCommittee() []sc.Worker {
	return cfg.Committee
}

func GetShards() []sc.Worker {
	return cfg.Shard
}

func GetCandidate() []sc.Worker {
	return cfg.Candidate
}

func GetShardSize() int {
	if cfg.Size == "" {
		return 5
	} else {
		i, err := strconv.Atoi(cfg.Size)
		if err != nil {
			panic("error")
		}

		if i < 1 || i > 200 {
			panic("error")
		}
		return i
	}
}

func GetNodePubKey() []byte {
	return []byte(cfg.Pubkey)
}
