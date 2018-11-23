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
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/viper"

	"flag"

	"strings"

	"github.com/ecoball/go-ecoball/account"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/utils"
	"path/filepath"
)

const VirtualBlockCpuLimit float64 = 200000000.0
const VirtualBlockNetLimit float64 = 1048576000.0
const BlockCpuLimit float64 = 200000.0
const BlockNetLimit float64 = 1048576.0

// peer list
var PeerList []string
var PeerIndex []string

const (
	StringBlock  = "/Block"
	StringHeader = "/Header"
	StringTxs    = "/Txs"
	StringState  = "/State"
	ListPeers    = "peer_list"
	IndexPeers   = "peer_index"
)

var configDefault = `#toml configuration for EcoBall system
http_port = "20678"          # client http port
wallet_http_port = "20679"   # client wallet http port
version = "1.0"              # system version
onlooker_port = "9001"		 #port for browser
output_to_terminal = "true"  # debug output type	 	
root_dir = "/tmp/"        		 # level file location
log_level = 1                # debug level	
consensus_algorithm = "SHARD" # can set as SOLO, DPOS, ABABFT, SHARD
time_slot = 500              # block interval time, uint ms
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


peer_list = [ "120202c924ed1a67fd1719020ce599d723d09d48362376836e04b0be72dfe825e24d810000", 
              "120202935fb8d28b70706de6014a937402a30ae74a56987ed951abbe1ac9eeda56f0160000" ]
peer_index = [ "1", "2" ]

#p2p swarm config info
p2p_peer_privatekey  = "CAAS4QQwggJdAgEAAoGBALna9LG/OdOImFPZ19WXzpCnCegonngYny888RvEUl/YcMpNQ1Rclpo/rtNiBlcxuXW7TepW/afQ0Y1yq8aRuRe7526RUQ8sLWc2mfCvV/HL6b1614qH8Q9HODnHTNIKzya+0PZuLNsS4Rug5dwMJHMKW8sAQK7TVvz5sdU+qa4vAgMBAAECgYB+gMqNMdvqX89PQ7flaq7vRsM3gm5a0GeJf7GddMOc+XXMPUrW4S6hTzdwKgim0PGrcRJXr154G2qHHMZPImEY3ZBgI1k7wawJFiTpFq6KEK7kN1yh0Baj3XmtDVysa0x3gzkuKmDEgyoaXilOMYkDU1egJHQpm7Q1gL7lY4/iAQJBAN4OcEl83zFG2J4Yb/QOP1eshKMdEPVYN45jZLgkG0EKcM4QCTBLDNbnCnDKcxbYwBJGiwCtf+XSAHGtG5KYDuUCQQDWQ+Mr8/aHV/zFDROsF+zbfNOebTMp9pIBYouPp3bVj/0atlv1cMdquOM6vMMoNzHjXDVelgp5pwunTfbPweODAkEAzwvhcPQI29Z2FfstL/+02hfW2Iw6irkFnDNa70NjUiLdCZX0K15fC2YD2yU5aH0Toja6VxhvH6fOmC/TfL1hbQJBAJXG1uI+o7Jwey1zurCt+NBlLbitNPq8dcuqC0zcD2GySYeGujmUIJIltBG3KeTO0HzSVCxOTfxEHQ1SnpkUO+kCQGrAkPrA0qIGsYHe3Kk+FbvY6orzyiPBhRaAQphAx96gg2lUxi4NeM3qxlakHq+Vh8Y+xr1b7VZ2mw9bfJViLkY="
p2p_peer_publickey   = "CAASogEwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBALna9LG/OdOImFPZ19WXzpCnCegonngYny888RvEUl/YcMpNQ1Rclpo/rtNiBlcxuXW7TepW/afQ0Y1yq8aRuRe7526RUQ8sLWc2mfCvV/HL6b1614qH8Q9HODnHTNIKzya+0PZuLNsS4Rug5dwMJHMKW8sAQK7TVvz5sdU+qa4vAgMBAAE="
p2p_listen_address   = ["/ip4/0.0.0.0/tcp/4013","/ip6/::/tcp/4013"]
announce_address     = []
no_announce_address  = []
bootstrap_address    = []
disable_nat_port_map = false
disable_relay        = false
enable_relay_hop     = false
conn_mgr_lowwater    = 600
conn_mgr_highwater   = 900
conn_mgr_graceperiod = 20

#p2p local discovery config info
enable_local_discovery = false
disable_localdis_log   = true

#dsn config
dsn_storage = false
dsn_path = "/tmp/storage"

#sharding config info
disable_sharding  = false

log_dir = "/tmp/Log/"        	 		# log file location
[logbunny]
debug_level=0                           # 0: debug 1: info 2: warn 3: error 4: panic 5: fatal
logger_type=0                           # 0: zap 1: logrus
with_caller=false
logger_encoder=1                        # 0: json 1: console
skip=4                                  # call depth, zap log is 3, logger is 4
time_pattern="2006-01-02 15:04:05.000"
#file name, file location is log_dir + name
debug_log_filename="ecoball.log"   		# or 'stdout' / 'stderr'
info_log_filename="ecoball.log"     	# or 'stdout' / 'stderr'
warn_log_filename="ecoball.log"     	# or 'stdout' / 'stderr'
error_log_filename="error.log"   		# or 'stdout' / 'stderr'
fatal_log_filename="fatal.log"  	 	# or 'stdout' / 'stderr'
#debug_log_filename="stdout"            # or 'stdout' / 'stderr'
#info_log_filename="stdout"             # or 'stdout' / 'stderr'
#error_log_filename="stdout"            # or 'stdout' / 'stderr'
http_port=":50015"                      # RESTFul API to change logout level dynamically
rolling_time_pattern="0 0 0 * * *"      # rolling the log everyday at 00:00:00
`

type SwarmConfigInfo struct {
	PrivateKey         string
	PublicKey          string
	ListenAddress      []string
	AnnounceAddr       []string
	NoAnnounceAddr     []string
	BootStrapAddr      []string
	DisableNatPortMap  bool
	DisableRelay       bool
	EnableRelayHop     bool
	ConnLowWater       int
	ConnHighWater      int
	ConnGracePeriod    int
}

var (
	ChainHash            common.Hash
	TimeSlot             int
	HttpLocalPort        string
	WalletHttpPort       string
	OnlookerPort         string
	EcoVersion           string
	RootDir              string
	LogDir               string
	OutputToTerminal     bool
	LogLevel             int
	IpfsDir              string
	ConsensusAlgorithm   string
	StartNode            bool
	Root                 account.Account
	User                 account.Account
	Delegate             account.Account
	Worker               account.Account
	Worker1              account.Account
	Worker2              account.Account
	Worker3              account.Account
	SwarmConfig          SwarmConfigInfo
	EnableLocalDiscovery bool
	DisableLocalDisLog   bool
	DsnStorage           bool
	DisableSharding      bool
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
	fmt.Println("init config")
	//set ecoball.toml dir
	var configDir string
	if flag.Lookup("test.v") == nil {
		fmt.Println("normal run")
		IpfsDir = "/tmp/storage"
		configDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		configDir = strings.Replace(configDir, "\\", "/", -1)
	} else {
		fmt.Println("run under go test")
		IpfsDir = "/tmp/storage"
		configDir = "/tmp/"
	}
	fmt.Println("program run dir:", configDir)
	if err := SetConfig(configDir); err != nil {
		fmt.Println("init config failed: ", err)
		os.Exit(-1)
	}
	initVariable()
	fmt.Println("log dir:", LogDir)
	fmt.Println("root dir:", RootDir)
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
	PeerList = viper.GetStringSlice(ListPeers)
	PeerIndex = viper.GetStringSlice(IndexPeers)
	ChainHash = common.SingleHash(common.FromHex(viper.GetString("root_pubkey")))

	//init p2p swarm configuration
	SwarmConfig = SwarmConfigInfo{
		PrivateKey: viper.GetString("p2p_peer_privatekey"),
		PublicKey: viper.GetString("p2p_peer_publickey"),
		ListenAddress: viper.GetStringSlice("p2p_listen_address"),
		AnnounceAddr: viper.GetStringSlice("announce_address"),
		NoAnnounceAddr: viper.GetStringSlice("no_announce_address"),
		BootStrapAddr: viper.GetStringSlice("bootstrap_address"),
		DisableNatPortMap: viper.GetBool("disable_nat_port_map"),
		DisableRelay:      viper.GetBool("disable_relay"),
		EnableRelayHop:    viper.GetBool("enable_relay_hop"),
		ConnLowWater:      viper.GetInt("conn_mgr_lowwater"),
		ConnHighWater:     viper.GetInt("conn_mgr_highwater"),
		ConnGracePeriod:   viper.GetInt("conn_mgr_graceperiod"),
	}

	EnableLocalDiscovery = viper.GetBool("enable_local_discovery")
	DisableLocalDisLog = viper.GetBool("disable_localdis_log")
	DsnStorage = viper.GetBool("dsn_storage")
	IpfsDir = viper.GetString("dsn_path")
	DisableSharding = viper.GetBool("disable_sharding")
}
