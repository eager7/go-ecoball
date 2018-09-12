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

func readConfigFile(filename string) *config {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	file := dir + "\\config.json"

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

var configLoad = false
var candidate []NodeConfig

func LoadConfig() {
	c := readConfigFile("config.json")
	if c == nil {
		return
	}

	for _, member := range c.Shard {
		candidate = append(candidate, member)
	}

}
