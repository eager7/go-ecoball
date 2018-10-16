package wasmservice

import (
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/common"
	"math/big"
)

// C API: issueToken(char *name, int32 nameLen, int32 maxSupply, char *issuer, int32 issuerLen)
func (ws *WasmService) createToken(proc *exec.Process, name, nameLen, maxSupply, issuer, issuerLen int32) int32 {
	if maxSupply <= 0 {
		return -1
	}

	name_msg := make([]byte, nameLen)
	err := proc.ReadAt(name_msg, int(name), int(nameLen))
	if err != nil{
		return -1
	}

	// C string end with '\0', but Go not. So delete '\0'
	Length := len(name_msg)
	var nameSlice []byte = name_msg[:Length - 1]
	if name_msg[Length - 1] != 0 {
		nameSlice = append(nameSlice, name_msg[Length - 1])
	}


	issuer_msg := make([]byte, issuerLen)
	err = proc.ReadAt(issuer_msg, int(issuer), int(issuerLen))
	if err != nil{
		return -1
	}

	// C string end with '\0', but Go not. So delete '\0'
	Length = len(issuer_msg)
	var issuerSlice []byte = issuer_msg[:Length - 1]
	if issuer_msg[Length - 1] != 0 {
		issuerSlice = append(issuerSlice, issuer_msg[Length - 1])
	}

	_, err = ws.state.CreateToken(string(nameSlice), maxSupply, common.NameToIndex(string(issuerSlice)))
	if err != nil{
		return -2
	}

	return 0
}

// C API: issueToken(char *to, int32 toLen, int32 amount, char *name, int32 nameLen, char *perm, int32 permLen)
func (ws *WasmService) issueToken(proc *exec.Process, to, toLen, amount, name, nameLen int32) int32{
	if amount <= 0 {
		return -1
	}

	to_msg := make([]byte, toLen)
	err := proc.ReadAt(to_msg, int(to), int(toLen))
	if err != nil{
		return -1
	}

	// C string end with '\0', but Go not. So delete '\0'
	Length := len(to_msg)
	var toSlice []byte = to_msg[:Length - 1]
	if to_msg[Length - 1] != 0 {
		toSlice = append(toSlice, to_msg[Length - 1])
	}


	name_msg := make([]byte, nameLen)
	err = proc.ReadAt(name_msg, int(name), int(nameLen))
	if err != nil{
		return -1
	}

	// C string end with '\0', but Go not. So delete '\0'
	Length = len(name_msg)
	var nameSlice []byte = name_msg[:Length - 1]
	if name_msg[Length - 1] != 0 {
		nameSlice = append(nameSlice, name_msg[Length - 1])
	}

	err = ws.state.IssueToken(common.NameToIndex(string(toSlice)), amount, string(nameSlice))
	if err != nil{
		return -2
	}

	return 0
}

// C API: transfer(char *from, int32 fromLen, char *to, int32 toLen, int32 amount, char *name, int32 nameLen, char *perm, int32 permLen)
func (ws *WasmService)transfer(proc *exec.Process, from, fromLen, to, toLen, amount, name, nameLen, perm, permLen int32) int32{
	if amount <= 0 {
		return -1
	}

	from_msg := make([]byte, fromLen)
	err := proc.ReadAt(from_msg, int(from), int(fromLen))
	if err != nil{
		return -1
	}

	// C string end with '\0', but Go not. So delete '\0'
	Length := len(from_msg)
	var fromSlice []byte = from_msg[:Length - 1]
	if from_msg[Length - 1] != 0 {
		fromSlice = append(fromSlice, from_msg[Length - 1])
	}

	perm_msg := make([]byte, permLen)
	err = proc.ReadAt(perm_msg, int(perm), int(permLen))
	if err != nil{
		return -1
	}

	// C string end with '\0', but Go not. So delete '\0'
	Length = len(perm_msg)
	var permSlice []byte = perm_msg[:Length - 1]
	if perm_msg[Length - 1] != 0 {
		permSlice = append(permSlice, perm_msg[Length - 1])
	}

	if err := ws.state.CheckAccountPermission(common.NameToIndex(string(fromSlice)), ws.action.ContractAccount, string(permSlice)); err != nil {
		return -1
	}

	to_msg := make([]byte, toLen)
	err = proc.ReadAt(to_msg, int(to), int(toLen))
	if err != nil{
		return -1
	}

	// C string end with '\0', but Go not. So delete '\0'
	Length = len(to_msg)
	var toSlice []byte = to_msg[:Length - 1]
	if to_msg[Length - 1] != 0 {
		toSlice = append(toSlice, to_msg[Length - 1])
	}

	name_msg := make([]byte, nameLen)
	err = proc.ReadAt(name_msg, int(name), int(nameLen))
	if err != nil{
		return -1
	}

	// C string end with '\0', but Go not. So delete '\0'
	Length = len(name_msg)
	var nameSlice []byte = name_msg[:Length - 1]
	if name_msg[Length - 1] != 0 {
		nameSlice = append(nameSlice, name_msg[Length - 1])
	}

	if err := ws.state.AccountSubBalance(common.NameToIndex(string(fromSlice)), string(nameSlice), big.NewInt(int64(amount))); err != nil {
		return -2
	}
	if err := ws.state.AccountAddBalance(common.NameToIndex(string(toSlice)), string(nameSlice), big.NewInt(int64(amount))); err != nil {
		return -3
	}

	return 0
}
