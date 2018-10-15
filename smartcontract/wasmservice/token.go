package wasmservice

import (
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"math/big"
)

func (ws *WasmService) createToken(proc *exec.Process, name, nameLen, maxSupply, supply, issuer, issuerLen int32) int32 {
	name_msg := make([]byte, nameLen)
	err := proc.ReadAt(name_msg, int(name), int(nameLen))
	if err != nil{
		return -1
	}

	issuer_msg := make([]byte, issuerLen)
	err = proc.ReadAt(issuer_msg, int(issuer), int(issuerLen))
	if err != nil{
		return -1
	}

	_, err = ws.state.CreateToken(string(name_msg), maxSupply, supply, common.NameToIndex(string(issuer_msg)))
	if err != nil{
		return -2
	}

	return 0
}

func (ws *WasmService) issueToken(proc *exec.Process, to, toLen, amount, name, nameLen int32) int32{
	to_msg := make([]byte, toLen)
	err := proc.ReadAt(to_msg, int(to), int(toLen))
	if err != nil{
		return -1
	}

	name_msg := make([]byte, nameLen)
	err = proc.ReadAt(name_msg, int(name), int(nameLen))
	if err != nil{
		return -1
	}

	err = ws.state.IssueToken(common.NameToIndex(string(to_msg)), amount, string(name_msg))
	if err != nil{
		return -2
	}

	return 0
}

// C API: inline_action(char *from, char *to, int amount, char *perm)
func (ws *WasmService)transfer(proc *exec.Process, from, fromLen, to, toLen, amount, perm, permLen int32) int32{
	from_msg := make([]byte, fromLen)
	err := proc.ReadAt(from_msg, int(from), int(fromLen))
	if err != nil{
		return -1
	}

	perm_msg := make([]byte, permLen)
	err = proc.ReadAt(perm_msg, int(perm), int(permLen))
	if err != nil{
		return -1
	}

	if err := ws.state.CheckAccountPermission(common.NameToIndex(string(from_msg)), ws.action.ContractAccount, string(perm_msg)); err != nil {
		return -1
	}

	to_msg := make([]byte, toLen)
	err = proc.ReadAt(to_msg, int(to), int(toLen))
	if err != nil{
		return -1
	}


	if err := ws.state.AccountSubBalance(common.NameToIndex(string(from_msg)), state.AbaToken, big.NewInt(int64(amount))); err != nil {
		return -2
	}
	if err := ws.state.AccountAddBalance(common.NameToIndex(string(to_msg)), state.AbaToken, big.NewInt(int64(amount))); err != nil {
		return -3
	}

	return 0
}
