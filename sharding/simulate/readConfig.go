package simulate

import (
	"encoding/json"
	"fmt"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type config struct {
	Pubkey    string
	Address   string
	Port      string
	Size      string
	Committee []sc.Worker
	Shard     []sc.Worker
}

func readConfigFile() *config {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	file := "sharding.json"

	bytes, err := ioutil.ReadFile(file)
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

func LoadConfig() {
	cfg = readConfigFile()
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
