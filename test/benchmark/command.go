package benchmark

import (
	"encoding/hex"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/common/errors"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var log = elog.Log

var num = 0

const (
	Committee       = "20681 "
	Shard1          = "20682 "
	Shard2          = "20683 "
	Shard3          = "20684 "
	ClientExecute   = "./ecoclient  --port="
	WalletCreate    = ClientExecute + "wallet create -n wallet.dat -p password"
	WalletOpen      = ClientExecute + "wallet open -n wallet.dat -p password"
	WalletImport    = ClientExecute + "wallet import -n wallet.dat -k "
	WalletCreateKey = ClientExecute + "wallet createkey -n wallet.dat"
	TxTransfer      = ClientExecute + "transfer -f root -t root -v 1"
	TxPledge        = ClientExecute
)

func runCmd(shell string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", shell)
	out, err := cmd.Output()
	if err != nil {
		log.Warn("exec ", cmd.Args, "failed, ", err.Error())
		log.Warn(string(out))
		return "", err
	}
	log.Notice("exec [", num, "]", cmd.Args, "success")
	num++
	return string(out), err
}

func RunCmd(shell string) error {
	_, err := runCmd(shell)
	return err
}

func GetDockerPort() {
	ret, err := runCmd("docker ps -a")
	errors.CheckErrorPanic(err)
	log.Info(ret)
	reg := regexp.MustCompile(`(.{5})->20678/tcp   ecoball_\d`)
	dockers := reg.FindAllString(ret, -1)
	log.Debug(dockers)
}

func CreateKey() (string, string) {
	ret, err := runCmd(WalletCreateKey)
	errors.CheckErrorPanic(err)
	reg := regexp.MustCompile(`(?i:PrivateKey:).*`)
	privateKey := reg.FindAllString(ret, -1)
	reg = regexp.MustCompile(`(?i:PublicKey:).*`)
	publicKey := reg.FindAllString(ret, -1)
	private := strings.Replace(privateKey[0], "PrivateKey:", "", -1)
	public := strings.Replace(publicKey[0], "PublicKey:", "", -1)

	return private, public
}

func ImportKey(private string) {
	runCmd(WalletImport + private)
}

func init() {
	RunCmd(WalletImport + hex.EncodeToString(config.Root.PrivateKey))
}

func SendTransaction(from, to, shard string) {
	ret, _ := runCmd(ClientExecute + shard + "transfer -f " + from + " -t " + to + " -v 1")
	log.Debug(ret)
}

func SendPledgeTx(from, to, shard string) {
	runCmd(ClientExecute + shard + "contract invoke -n root -i root -m pledge -p " + from + "," + to + ",100,100")
}

func BenchMarkTransaction() {
	var wg sync.WaitGroup
	for i := 0; i < 250; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 4; i++ {
				SendTransaction("root", "root", Shard3)
			}
		}()
	}
	wg.Wait()
}
