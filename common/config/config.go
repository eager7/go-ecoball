// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/viper"

	"flag"

	"strings"

	"fmt"
	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/utils"
	"path/filepath"
)

const VirtualBlockCpuLimit float64 = 200000000.0
const VirtualBlockNetLimit float64 = 1048576000.0
const BlockCpuLimit float64 = 200000.0
const BlockNetLimit float64 = 1048576.0 //1M

const (
	StringBlock      = "/Block"
	StringBlockCache = "/BlockCache"
	StringHeader     = "/Header"
	StringTxs        = "/Txs"
	StringState      = "/State"
	ListPeers        = "peer_list"
	IndexPeers       = "peer_index"
)

var configDefault = `#toml configuration for EcoBall system
http_port = "20678"          # client http port
wallet_http_port = "20679"   # client wallet http port
version = "1.0"              # system version
onlooker_port = "9001"		 #port for browser
root_dir = "./"        		 # level file location
log_dir = "/tmp/Log/"        # log file location
output_to_terminal = "true"  # debug output type	 
log_level = 1                # debug level	
consensus_algorithm = "SOLO" # can set as SOLO, DPOS
time_slot = 5000             # block interval time, uint ms
start_node = "true"
root_privkey = "34a44d65ec3f517d6e7550ccb17839d391b69805ddd955e8442c32d38013c54e"
root_pubkey = "04de18b1a406bfe6fb95ef37f21c875ffc9f6f59e71fea8efad482b82746da148e0f154d708001810b52fb1762d737fec40508b492628f86c605391a891a61ad0b" # used to chain ID
user_privkey = "34a44d65ec3f517d6e7550ccb17839d391b69805ddd955e8442c32d38013c54e"
user_pubkey = "04de18b1a406bfe6fb95ef37f21c875ffc9f6f59e71fea8efad482b82746da148e0f154d708001810b52fb1762d737fec40508b492628f86c605391a891a61ad0b"


#debug config info
worker1_privkey = "cb0324ee8f7bd11dec57e39c4f560b9343c6c81c71012b96be29f26b92fef6f9"
worker1_pubkey = "0425adbea1ddc21124279059057b4c9b5df4d40e49f2625504b45e0d43aea22c25621e42307eb8224f7ea0e65d40c0495d3cbd3f020f801f38b73cec5740bf1ec9"

worker2_privkey = "05cac9544f828b570724eb52b5903a68fbe0c8f23a15851cb717a5f7eda801cd"
worker2_pubkey = "040cf9d46f4f5945ed7986cb8920feb5ac4eb06bb26cb048ed9dc2de4c54c19914bf4adf5ca0571a6f106bf4542fc7bfcfd164d8065598fc76042c074b24048960"

worker3_privkey = "79b99bbd11bd14e8c0da65c20bae059d1eee06f92380fb88ff31a88c84d3fc6e"
worker3_pubkey = "04717944fa32da2261eeda1e810c3b3c62216ed486785a9aa78e2cde0f18805882631033aed956d02721e9fae079e600bd512d4feb0375a14d882a63e48971d413"

delegate_privkey = "56bd8432606e6e2eb354794d89059f7f9e9a0de62166145b898136b496be6aed"
delegate_pubkey  = "04070a106e034b11e03bab17aab0d2e75d7795bae8346f6f527f436cd714f7798efdeced276f326ed3406e3baab257487330e61896c838920328a4d745a87f06d1"

worker_privkey = "8bbd547fe9d9e867721c6fa643fbe637fc3d955e588358a45c11d63dd5a25016"
worker_pubkey = "041a0a2b0bfce1d624c125d2a9fcca16c5b2b96bc78ab827e1c23818df4a70a4441c9665850268d48ab23e102cf1dc6864596a19e748c0867dce400a3f219e3f62"


[p2p] #p2p swarm config info
p2p_peer_privatekey  = "CAAS4QQwggJdAgEAAoGBALna9LG/OdOImFPZ19WXzpCnCegonngYny888RvEUl/YcMpNQ1Rclpo/rtNiBlcxuXW7TepW/afQ0Y1yq8aRuRe7526RUQ8sLWc2mfCvV/HL6b1614qH8Q9HODnHTNIKzya+0PZuLNsS4Rug5dwMJHMKW8sAQK7TVvz5sdU+qa4vAgMBAAECgYB+gMqNMdvqX89PQ7flaq7vRsM3gm5a0GeJf7GddMOc+XXMPUrW4S6hTzdwKgim0PGrcRJXr154G2qHHMZPImEY3ZBgI1k7wawJFiTpFq6KEK7kN1yh0Baj3XmtDVysa0x3gzkuKmDEgyoaXilOMYkDU1egJHQpm7Q1gL7lY4/iAQJBAN4OcEl83zFG2J4Yb/QOP1eshKMdEPVYN45jZLgkG0EKcM4QCTBLDNbnCnDKcxbYwBJGiwCtf+XSAHGtG5KYDuUCQQDWQ+Mr8/aHV/zFDROsF+zbfNOebTMp9pIBYouPp3bVj/0atlv1cMdquOM6vMMoNzHjXDVelgp5pwunTfbPweODAkEAzwvhcPQI29Z2FfstL/+02hfW2Iw6irkFnDNa70NjUiLdCZX0K15fC2YD2yU5aH0Toja6VxhvH6fOmC/TfL1hbQJBAJXG1uI+o7Jwey1zurCt+NBlLbitNPq8dcuqC0zcD2GySYeGujmUIJIltBG3KeTO0HzSVCxOTfxEHQ1SnpkUO+kCQGrAkPrA0qIGsYHe3Kk+FbvY6orzyiPBhRaAQphAx96gg2lUxi4NeM3qxlakHq+Vh8Y+xr1b7VZ2mw9bfJViLkY="
p2p_peer_publickey   = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALna9LG/OdOImFPZ19WXzpCnCegonngYny888RvEUl/YcMpNQ1Rclpo/rtNiBlcxuXW7TepW/afQ0Y1yq8aRuRe7526RUQ8sLWc2mfCvV/HL6b1614qH8Q9HODnHTNIKzya+0PZuLNsS4Rug5dwMJHMKW8sAQK7TVvz5sdU+qa4vAgMBAAE="
p2p_listen_address   = ["/ip4/0.0.0.0/tcp/4013","/ip6/::/tcp/4013"]
announce_address     = []
no_announce_address  = []
bootstrap_address    = ["/ip4/192.168.8.35/tcp/4013/ipfs/QmUTyDE2SGS1kmZYZqKtHn87CYwWcyjwL72WRVU4xJCScw","/ip4/192.168.8.140/tcp/4013/ipfs/QmXbAYKPLHDdRb9GyFf8QtapherBWCsM787CuZr8g3CcWd"]
disable_nat_port_map = false
disable_relay        = false
enable_relay_hop     = false
conn_mgr_lowwater    = 600
conn_mgr_highwater   = 900
conn_mgr_graceperiod = 20
#p2p local discovery config info
enable_local_discovery = false
disable_localdis_log   = true

	
[logbunny]
debug_level=0                           # 0: debug 1: info 2: warn 3: error 4: panic 5: fatal
logger_type=0                           # 0: zap 1: logrus
with_caller=false
logger_encoder=1                        # 0: json 1: console
skip=4                                  # call depth, zap log is 3, logger is 4
time_pattern="2006-01-02 15:04:05.00000"
#file name, file location is log_dir + name
debug_log_filename="ecoball.log"   		# or 'stdout' / 'stderr'
info_log_filename="ecoball.log"     	# or 'stdout' / 'stderr'
warn_log_filename="ecoball.log"     	# or 'stdout' / 'stderr'
error_log_filename="ecoball.log"   		# or 'stdout' / 'stderr'
fatal_log_filename="ecoball.log"  	 	# or 'stdout' / 'stderr'
#debug_log_filename="stdout"            # or 'stdout' / 'stderr'
#info_log_filename="stdout"             # or 'stdout' / 'stderr'
#error_log_filename="stdout"            # or 'stdout' / 'stderr'
http_port=":50015"                      # RESTFul API to change logout level dynamically
rolling_time_pattern="0 0 0 * * *"      # rolling the log everyday at 00:00:00

[consensus]
Pubkey1  = "CAAS4AQwggJcAgEAAoGBALjW6cjkZl/mqR8L1zY/8swFAvnwFCUJdVAz52a61UGhgeLegT01Rh4YexoXizQNnDLRLN991Sza2GTSpPJKj0SQ1ofj/MOhPULcxUeK+1Py2Dgq8VIoKFl/D9Lv7Dx59fMl2NpQiax+RansJJtSeiSlqj0oJPqnCFvnq8vLsG+zAgMBAAECgYEAh0kNPXMmFuT9PXLuJo+xhm/YmNSF+gGtMnF62W6/rVSne0Q9tW3rjxV97D/1K7kWbP86Z61yvGzE2y5tecTmjBtBDXeTS0wWJvTy1iWNhvRogjrbhiMjE9Y6aLEO7NOIUpEUboTkBdCvvCe9PpsuarCcsAbGlOih7GFLdkMwWvECQQDbr00Mu86XgKojnFBZgokMlZ8+Zv8HAPHLIaCLds8Oy/EGZeBQCWCc8qYtnjnGO5pB3+cc2cXM1tzsUeOCSfB7AkEA12UEj2cnoB1yaerU6rypGGChbz6tlpJ55oCpH0mw+NFp3IHSJ3gUON07sINef4VQtROE3lZC1Yj4LF9WZBUEKQJAdjGuxrcUw7ZZ06b6I+5zRe4KK0zG0UHU1XFWKzLU3CUlnEebk/Q3orl6ZvjGJL1UlTSd54vTPA4t9odoXGTjmQJAentsG2uqQcdc71PlHVKIyV7xjcPTjCLhBK02/p617teOXiDIcz86KJfNQHODgfo6Sa2+yXu955VKoljYVHMK+QJAcTe4VTuUXsodOL+mu3zfBgle2B8BcFnaaQINn4+S5t/Rgn+qqMpykXTlidL28ulKWmJFMFBvwm4Z50vqS8REXA=="
Private1 = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALjW6cjkZl/mqR8L1zY/8swFAvnwFCUJdVAz52a61UGhgeLegT01Rh4YexoXizQNnDLRLN991Sza2GTSpPJKj0SQ1ofj/MOhPULcxUeK+1Py2Dgq8VIoKFl/D9Lv7Dx59fMl2NpQiax+RansJJtSeiSlqj0oJPqnCFvnq8vLsG+zAgMBAAE="
Address1 = "127.0.0.1"
Port1    = "8001"

Pubkey2  = "CAAS4QQwggJdAgEAAoGBAMzzmP5b/eHZQJZfp9657Tyh7T8shzxCVNavPTveqpgOkVGIT432rBm7eqwizikVq3U/pRzRYr9A1nD5L0WjMd3PUFhIb0sy26d9zpI02ROUhAi1ZPkzh8eKrfrqw8N23XB8In6gjC498bbH1iPyXjuRlF8MaSqsR3Q2ukcF/dp/AgMBAAECgYAkxtF9UySLkmB0m1WUMejQKH5aB7N8rKpsm6VxSNNz1ald6AfegZUASRQKL3SvCqRptbH7Kdd+WjQgsZY5+L7JrEJUkQFtQTsqhdrMIZEkjbsqnqk0QKn9Y8rGcJofGh4d5VzTobs9R8Phr+dwDGlBFpO6R7EmfquvQU7n/xUBgQJBAPKbw4vmsbcfaVPICybQPRqK+DKugCCua09vqQegJRVfgWQRHyP1vWX8S9Coqyj+RP6rxn5E6joQ0q+DaU3ZMOMCQQDYQ7bME5vdE2j7lfvQceQVqzZC5AlnLP6WUfNAH3IN9ew7XGjsEWl2wgMXBcRIpl/DkM5LBPG8AWLmVw4tPK61AkEArb8d+Vh7F9GYJhdS3TYvPI4gGHPecQlY8ufd3wcy566hRN/6NE+ul5ZrWYEiK1aGZPjyS8XhFTqtGGN9i/IqzwJBAKCgCpEgr09QL1VNXK7BKIr/k1mzTViYjq7PR0CFGo1L7p1YUYWkmRRfnTPoUJU1HUN/tfj6PyFIVlCGsDzhKVECQECUWqEjHSVTM0mz8p3vgWZSN5F5x1qpQTxRArEfCfmjzlseW/ZUBgE8ABRHraJ75AAM0Y2TCodtwt5y4jVyHuc="
Private2 = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMzzmP5b/eHZQJZfp9657Tyh7T8shzxCVNavPTveqpgOkVGIT432rBm7eqwizikVq3U/pRzRYr9A1nD5L0WjMd3PUFhIb0sy26d9zpI02ROUhAi1ZPkzh8eKrfrqw8N23XB8In6gjC498bbH1iPyXjuRlF8MaSqsR3Q2ukcF/dp/AgMBAAE="
Address2 = "127.0.0.1"
Port2    = "8002"

Pubkey3  = "CAAS4AQwggJcAgEAAoGBALXF5CCvkjnK4cSEdF7wTtZui85X6X3+tox0omWr+rEU97DFkL+Z5n2GoRZ0xnSJ22v5uMwHf9Vo8DL9ncoccZH3OI842Do+7RON1+jxLuNdVIc1neuaQULniLVLpnl1Hd3MljZNelPtBMZWW7vzekbWGx3arvodeTlMExxFUszdAgMBAAECgYBC4eeIp1FUdnQPzPTMoftAJzjF2c8ODxS4JYpDgr4hPifNIUSbW1NVyJ2pF5qV8suLtTzrxa6hpZUMDglq/oBCqAwGg/IwD18oSqT92H7bZ4ZWhMwotG+xtwe6EJTQmL485QrqeV2CWOD7Z/p7UDeJZmKBvpihCEie8cZPM6sNeQJBAOOwD7LzLIF3nfRWQXcQAuQg3avX6nw5VXBzpoXv4mvMM21hJncJMqZ6BTw4NPbjkhfpA1NT/d1uJVEp199LxMsCQQDMYDxCQwdPsfm7QKKZWJvo4ytiS4pr4uAUvUVhBP0wyATqMnSVMdbJs8BOahGw6Di5kWoz2BdP2AsZyoiOFyf3AkEAjR8W2+d08lndgQ/lS5KU+CiWvGf7Yjt3BVfpIqLoR8AtL+JDIQyGZEDE9eowicXLSx6VfRRWCOS4JHI25qPjuQJAAgVRkzYmdFtGJNvWv71ojTzxyN8GV1q+7HWSogrylfDkW4x0KqV7gjMMy7mwwxcIuIz/h9OzJ07zjSW7g+wmsQJARWsoCMt+TOYd7jwFTZP/27MIgd41OMbVgaLgZTkl71mm08wwQSm7g3O8xqRQrjpK+aHwmo8Z+yQSFEGF+DqYtA=="
Private3 = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALXF5CCvkjnK4cSEdF7wTtZui85X6X3+tox0omWr+rEU97DFkL+Z5n2GoRZ0xnSJ22v5uMwHf9Vo8DL9ncoccZH3OI842Do+7RON1+jxLuNdVIc1neuaQULniLVLpnl1Hd3MljZNelPtBMZWW7vzekbWGx3arvodeTlMExxFUszdAgMBAAE="
Address3 = "127.0.0.1"
Port3    = "8003"
`

type P2pConfig struct {
	PrivateKey           string
	PublicKey            string
	ListenAddress        []string
	AnnounceAddr         []string
	NoAnnounceAddr       []string
	BootStrapAddr        []string
	DisableNatPortMap    bool
	DisableRelay         bool
	EnableRelayHop       bool
	ConnLowWater         int
	ConnHighWater        int
	ConnGracePeriod      int
	EnableLocalDiscovery bool
	DisableLocalDisLog   bool
}
type Producer struct {
	Pubkey  string
	Address string
	Port    string
}
type ConsConfig struct {
	Pubkey    string
	Address   string
	Port      string
	Size      string
	Committee []Producer
	Shard     []Producer
	Candidate []Producer
}

var (
	ChainHash          common.Hash
	TimeSlot           int
	HttpLocalPort      string
	WalletHttpPort     string
	OnlookerPort       string
	EcoVersion         string
	RootDir            string
	LogDir             string
	OutputToTerminal   bool
	LogLevel           int
	ConsensusAlgorithm string
	StartNode          bool
	Root               account.Account
	User               account.Account
	Delegate           account.Account
	Worker             account.Account
	Worker1            account.Account
	Worker2            account.Account
	Worker3            account.Account
	PConfig            P2pConfig
	CConfig            ConsConfig
)

func SetConfig(filePath string) error {
	if err := CreateConfigFile(filePath, "ecoball.toml", configDefault); err != nil {
		return err
	}
	return InitConfig(filePath, "ecoball")
}

func CreateConfigFile(filePath, fileName, config string) error {
	var dir string
	var file string
	if "" == path.Ext(filePath) {
		dir = filePath
		file = path.Join(filePath, fileName)
	} else {
		dir = filePath
		file = path.Join(filePath, fileName)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Println("could not create directory:", dir, err)
			return err
		}
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if err := ioutil.WriteFile(file, []byte(config), 0644); err != nil {
			fmt.Println("write file err:", err)
			return err
		}
	}
	return nil
}

func defaultPath() (string, error) {
	return utils.DirHome()
}

func InitConfig(filePath, config string) error {
	viper.SetConfigName(config)
	viper.AddConfigPath(filePath)
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("can't load config file:", err)
		return err
	}
	return nil
}

func init() {
	//set ecoball.toml dir
	var configDir string
	if flag.Lookup("test.v") == nil {
		configDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		configDir = strings.Replace(configDir, "\\", "/", -1)
	} else {
		configDir = "/tmp/"
	}
	if err := SetConfig(configDir); err != nil {
		fmt.Println("init config failed: ", err)
		os.Exit(-1)
	}
	initVariable()
}

func initVariable() {
	TimeSlot = viper.GetInt("time_slot")
	HttpLocalPort = viper.GetString("http_port")
	WalletHttpPort = viper.GetString("wallet_http_port")
	OnlookerPort = viper.GetString("onlooker_port")
	EcoVersion = viper.GetString("version")
	LogDir = viper.GetString("log_dir")
	RootDir = viper.GetString("root_dir")
	OutputToTerminal = viper.GetBool("output_to_terminal")
	StartNode = viper.GetBool("start_node")
	LogLevel = viper.GetInt("log_level")
	ConsensusAlgorithm = viper.GetString("consensus_algorithm")
	Root = account.Account{PrivateKey: common.FromHex(viper.GetString("root_privkey")), PublicKey: common.FromHex(viper.GetString("root_pubkey")), Alg: 0}
	User = account.Account{PrivateKey: common.FromHex(viper.GetString("user_privkey")), PublicKey: common.FromHex(viper.GetString("user_pubkey")), Alg: 0}
	Worker1 = account.Account{PrivateKey: common.FromHex(viper.GetString("worker1_privkey")), PublicKey: common.FromHex(viper.GetString("worker1_pubkey")), Alg: 0}
	Worker2 = account.Account{PrivateKey: common.FromHex(viper.GetString("worker2_privkey")), PublicKey: common.FromHex(viper.GetString("worker2_pubkey")), Alg: 0}
	Worker3 = account.Account{PrivateKey: common.FromHex(viper.GetString("worker3_privkey")), PublicKey: common.FromHex(viper.GetString("worker3_pubkey")), Alg: 0}
	Delegate = account.Account{PrivateKey: common.FromHex(viper.GetString("delegate_privkey")), PublicKey: common.FromHex(viper.GetString("delegate_pubkey")), Alg: 0}
	Worker = account.Account{PrivateKey: common.FromHex(viper.GetString("worker_privkey")), PublicKey: common.FromHex(viper.GetString("worker_pubkey")), Alg: 0}
	ChainHash = common.SingleHash(common.FromHex(viper.GetString("root_pubkey")))

	//init p2p swarm configuration
	PConfig = P2pConfig{
		PrivateKey:           viper.GetString("p2p.p2p_peer_privatekey"),
		PublicKey:            viper.GetString("p2p.p2p_peer_publickey"),
		ListenAddress:        viper.GetStringSlice("p2p.p2p_listen_address"),
		AnnounceAddr:         viper.GetStringSlice("p2p.announce_address"),
		NoAnnounceAddr:       viper.GetStringSlice("p2p.no_announce_address"),
		BootStrapAddr:        viper.GetStringSlice("p2p.bootstrap_address"),
		DisableNatPortMap:    viper.GetBool("p2p.disable_nat_port_map"),
		DisableRelay:         viper.GetBool("p2p.disable_relay"),
		EnableRelayHop:       viper.GetBool("p2p.enable_relay_hop"),
		ConnLowWater:         viper.GetInt("p2p.conn_mgr_lowwater"),
		ConnHighWater:        viper.GetInt("p2p.conn_mgr_highwater"),
		ConnGracePeriod:      viper.GetInt("p2p.conn_mgr_graceperiod"),
		EnableLocalDiscovery: viper.GetBool("p2p.enable_local_discovery"),
		DisableLocalDisLog:   viper.GetBool("p2p.disable_localdis_log"),
	}
	CConfig = ConsConfig{
		Pubkey:  viper.GetString("consensus.Pubkey1"),
		Address: viper.GetString("consensus.Address1"),
		Port:    viper.GetString("consensus.Port1"),
		Size:    "",
		Committee: []Producer{{
			Pubkey:  viper.GetString("consensus.Pubkey1"),
			Address: viper.GetString("consensus.Address1"),
			Port:    viper.GetString("consensus.Port1"),
		}},
		Shard:     nil,
		Candidate: nil,
	}
}
