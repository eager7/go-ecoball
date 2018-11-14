package wasmservice

import (
	"github.com/ecoball/go-ecoball/vm/wasmvm/exec"
	"github.com/ecoball/go-ecoball/common"
	"math/big"
	"github.com/ecoball/go-ecoball/core/state"
	"unsafe"
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

	token, err := ws.state.CreateToken(string(nameSlice), big.NewInt(int64(maxSupply)), ws.action.ContractAccount, common.NameToIndex(string(issuerSlice)))
	if err != nil{
		return -2
	}

	// generate trx receipt
	data, err := token.Serialize()
	if err != nil {
		return -3
	}
	ws.context.Tc.Trx.Receipt.NewToken = data

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

	// Get token issuer
	tokenInfo, err := ws.state.GetTokenInfo(string(nameSlice))
	if err != nil {
		return -4
	}

	// if declared permission actor had permission of token issuer
	if tokenInfo.Issuer != ws.action.Permission.Actor {
		if err := ws.state.CheckAccountPermission(tokenInfo.Issuer, ws.action.Permission.Actor, "active"); err != nil {
			return -5
		}
	}

	err = ws.state.IssueToken(common.NameToIndex(string(toSlice)), big.NewInt(int64(amount)), string(nameSlice))
	if err != nil{
		return -2
	}

	// generate trx receipt
	token := state.Token{
		Name:		string(nameSlice),
		Balance: 	big.NewInt(int64(amount)),
	}
	acc := state.Account{
		Index:			common.NameToIndex(string(toSlice)),
		Tokens:			make(map[string]state.Token),
	}
	acc.Tokens[string(nameSlice)] = token

	data, err := acc.Serialize()
	if err != nil {
		return -3
	}
	ws.context.Tc.Trx.Receipt.Accounts[0] = data

	return 0
}

// C API: ABA_transfer(char *from, int32 fromLen, char *to, int32 toLen, int32 amount, char *name, int32 nameLen, char *perm, int32 permLen)
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

	// generate trx receipt
	ws.context.Tc.Trx.Receipt.From = common.NameToIndex(string(fromSlice))
	ws.context.Tc.Trx.Receipt.To = common.NameToIndex(string(toSlice))
	ws.context.Tc.Trx.Receipt.TokenName = string(nameSlice)
	ws.context.Tc.Trx.Receipt.Amount = big.NewInt(int64(amount))

	return 0
}
// C API: ABA_tokenExisted(char *name, int nameLen)
func (ws *WasmService)tokenExisted(proc *exec.Process, name, nameLen int32) int32 {
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

	if ws.state.TokenExisted(string(nameSlice)) {
		return 1
	} else {
		return 0
	}
}

type TokenStatus struct {
	Symbol 		 [12]byte				`json:"symbol"`
	MaxSupply 	 int64					`json:"max_supply"`
	Supply		 int64 					`json:"supply"`
	Issuer       [12]byte				`json:"issuer"`
}

type SliceMock struct {
	addr uintptr
	len  int
	cap  int
}
// C API: ABA_getTokenStatus(char *name, int nameLen, char *stat, int statLen)
func (ws *WasmService)getTokenStatus(proc *exec.Process, name, nameLen, stat, statLen int32) int32{
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

	// Get token info
	tokenInfo, err := ws.state.GetTokenInfo(string(nameSlice))
	if err != nil {
		return -2
	}

	// construct TokenStatus var
	var symbol [12]byte
	var issuer [12]byte
	issuerByte := common.IndexToName(tokenInfo.Issuer)
	//for i := 0; i < len(issuerByte); i++ {
	//	issuer[i] = issuerByte[i]
	//}

	//for i := 0; i < len(tokenInfo.Symbol); i++ {
	//	symbol[i] = tokenInfo.Symbol[i]
	//}

	copy(issuer[:], issuerByte)
	copy(symbol[:], tokenInfo.Symbol)

	status := &TokenStatus{
		Symbol:		symbol,
		MaxSupply:	tokenInfo.MaxSupply.Int64(),
		Supply:		tokenInfo.Supply.Int64(),
		Issuer:		issuer,
	}

	if int(statLen) > (int)(unsafe.Sizeof(*status)) {
		statLen = int32(unsafe.Sizeof(*status))
	}

	Len := unsafe.Sizeof(*status)
	bytes := &SliceMock{
		addr: uintptr(unsafe.Pointer(status)),
		cap:  int(Len),
		len:  int(Len),
	}

	// convert TokenStatus to []byte
	data := *(*[]byte)(unsafe.Pointer(bytes))

	err = proc.WriteAt(data[:], int(stat), int(statLen))
	if err != nil{
		return -1
	}

	return 0
}
// C API: ABA_putTokenStatus(char *name, int nameLen, char *stat, int statLen)
func (ws *WasmService)putTokenStatus(proc *exec.Process, name, nameLen, stat, statLen int32) int32{
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

	if ws.state.TokenExisted(string(nameSlice)) {
		// Get token creator
		tokenInfo, err := ws.state.GetTokenInfo(string(nameSlice))
		if err == nil {
			// only token creator can modify token info
			if tokenInfo.Creator != ws.action.ContractAccount {
				return -5
			}
		}

		// if declared permission actor had permission of token issuer
		//if tokenInfo.Issuer != ws.action.Permission.Actor {
		//	if err := ws.state.CheckAccountPermission(tokenInfo.Issuer, ws.action.Permission.Actor, "active"); err != nil {
		//		return -5
		//	}
		//}
	}

	data := make([]byte, statLen)
	err = proc.ReadAt(data, int(stat), int(statLen))
	if err != nil{
		return -1
	}

	status := *(**TokenStatus)(unsafe.Pointer(&data))
	//issuerByte := make([]byte, len(status.Issuer))
	//for i := 0; i < len(status.Issuer); i++ {
	//	issuerByte[i] =  status.Issuer[i]
	//}

	tokenInfo, err := ws.state.SetTokenInfo(string(nameSlice), big.NewInt(status.MaxSupply), big.NewInt(status.Supply), ws.action.ContractAccount, common.NameToIndex(string(status.Issuer[:])))
	if err != nil{
		return -2
	}

	// generate trx receipt
	byte, err := tokenInfo.Serialize()
	if err != nil {
		return -3
	}
	ws.context.Tc.Trx.Receipt.NewToken = byte

	return 0
}

// C API: ABA_getTokenInfo(char *name, int nameLen, int maxSupply, int maxSupplyLen, int supply, int supplyLen, char *issuer, int issuerLen)
func (ws *WasmService) getTokenInfo(proc *exec.Process, name, nameLen int32, maxSupply, maxSupplyLen, supply, supplyLen, issuer, issuerLen int32) int32 {
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

	// Get token info
	tokenInfo, err := ws.state.GetTokenInfo(string(nameSlice))
	if err != nil {
		return -2
	}

	num := tokenInfo.MaxSupply.Int64()
	Len := unsafe.Sizeof(num)
	bytes := &SliceMock{
		addr: uintptr(unsafe.Pointer(&num)),
		cap:  int(Len),
		len:  int(Len),
	}
	// convert TokenStatus to []byte
	data := *(*[]byte)(unsafe.Pointer(bytes))
	err = proc.WriteAt(data, int(maxSupply), int(maxSupplyLen))
	if err != nil{
		return -1
	}


	num = tokenInfo.Supply.Int64()
	Len = unsafe.Sizeof(num)
	bytes = &SliceMock{
		addr: uintptr(unsafe.Pointer(&num)),
		cap:  int(Len),
		len:  int(Len),
	}
	// convert TokenStatus to []byte
	data = *(*[]byte)(unsafe.Pointer(bytes))
	err = proc.WriteAt(data, int(supply), int(supplyLen))
	if err != nil{
		return -1
	}

	account := common.IndexToName(tokenInfo.Issuer)
	data = []byte(account)
	if int(issuerLen) > len(data) {
		issuerLen = int32(len(data))
	}
	err = proc.WriteAt(data, int(issuer), int(issuerLen))
	if err != nil{
		return -1
	}

	return 0
}
// C API: ABA_putTokenInfo(char *name, int nameLen, int maxSupply, int maxSupplyLen, int supply, int supplyLen, char *issuer, int issuerLen)
func (ws *WasmService) putTokenInfo(proc *exec.Process, name, nameLen int32, maxSupply, supply int64, issuer, issuerLen int32) int32 {
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

	if ws.state.TokenExisted(string(nameSlice)) {
		// Get token creator
		tokenInfo, err := ws.state.GetTokenInfo(string(nameSlice))
		if err == nil {
			// only token creator can modify token info
			if tokenInfo.Creator != ws.action.ContractAccount {
				return -5
			}
		}

		//// if declared permission actor had permission of token issuer
		//if tokenInfo.Creator != ws.action.ContractAccount {
		//	if err := ws.state.CheckAccountPermission(tokenInfo.Creator, ws.action.Permission.Actor, "active"); err != nil {
		//		return -5
		//	}
		//}
	}

	token, err := ws.state.SetTokenInfo(string(nameSlice), big.NewInt(maxSupply), big.NewInt(supply), ws.action.ContractAccount, common.NameToIndex(string(issuerSlice)))
	if err != nil{
		return -2
	}

	// generate trx receipt
	data, err := token.Serialize()
	if err != nil {
		return -3
	}
	ws.context.Tc.Trx.Receipt.NewToken = data

	return 0
}

// C API: ABA_getAccountBalance(char *account, int accountLen, char *name, int nameLen)
func (ws *WasmService)getAccountBalance(proc *exec.Process, account, accountLen, name, nameLen int32) int64{
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


	account_msg := make([]byte, accountLen)
	err = proc.ReadAt(account_msg, int(account), int(accountLen))
	if err != nil{
		return -1
	}

	Length = len(account_msg)
	var accountSlice []byte = account_msg[:Length - 1]
	if account_msg[Length - 1] != 0 {
		accountSlice = append(accountSlice, account_msg[Length - 1])
	}

	bal, err := ws.state.AccountGetBalance(common.NameToIndex(string(accountSlice)), string(nameSlice))
	if err != nil {
		return -2
	}

	return bal.Int64()
}
// C API: ABA_addAccountBalance(char *account, int accountLen, char *name, int nameLen, long long int amount)
func (ws *WasmService)addAccountBalance(proc *exec.Process, account, accountLen, name, nameLen int32, amount int64) int32{
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


	account_msg := make([]byte, accountLen)
	err = proc.ReadAt(account_msg, int(account), int(accountLen))
	if err != nil{
		return -1
	}

	Length = len(account_msg)
	var accountSlice []byte = account_msg[:Length - 1]
	if account_msg[Length - 1] != 0 {
		accountSlice = append(accountSlice, account_msg[Length - 1])
	}

	// Get token creator
	tokenInfo, err := ws.state.GetTokenInfo(string(nameSlice))
	if err != nil {
		return -4
	}

	// this api must invoke by token's creator contract
	if ws.action.ContractAccount != tokenInfo.Creator {
		return -5
	}

	if err := ws.state.AccountAddBalance(common.NameToIndex(string(accountSlice)), string(nameSlice), big.NewInt(amount)); err != nil {
		return -3
	}

	// generate trx receipt
	var acc state.Account
	var deltaByte []byte
	// if one contract action invoke this api some times to modify the same account's balance
	delta, ok := ws.context.Tc.AccountDelta[string(accountSlice)]
	if ok {	// some time
		err = acc.Deserialize(delta)
		if err != nil {
			return -4
		}
		// if accountDelta had existed, check if token existed in accountDelta
		balance, ok := acc.Tokens[string(nameSlice)]
		if ok {
			balance.Balance = new(big.Int).Add(balance.Balance, big.NewInt(amount))
			acc.Tokens[string(nameSlice)] = balance
		} else {
			newBalance := state.Token{
				Name:		string(nameSlice),
				Balance:	big.NewInt(amount),
			}
			acc.Tokens[string(nameSlice)] = newBalance
		}
	} else {	// first times
		acc = state.Account{
			Tokens:			make(map[string]state.Token),
			Index:			common.NameToIndex(string(accountSlice)),
		}

		balance := state.Token{
			Name:		string(nameSlice),
			Balance:	big.NewInt(amount),
		}
		acc.Tokens[string(nameSlice)] = balance
	}

	deltaByte, err = acc.Serialize()
	if err != nil {
		return -4
	}

	ws.context.Tc.AccountDelta[string(accountSlice)] = deltaByte

	var flag int = 0
	for _, accName := range ws.context.Tc.Accounts {
		if accName == string(accountSlice) {
			flag = 1
		}
	}

	if flag == 0 {
		ws.context.Tc.Accounts = append(ws.context.Tc.Accounts, string(accountSlice))
	}

	return 0
}
// C API: ABA_subAccountBalance(char *account, int accountLen, char *name, int nameLen, long long int amount)
func (ws *WasmService)subAccountBalance(proc *exec.Process, account, accountLen, name, nameLen int32, amount int64) int32{
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


	account_msg := make([]byte, accountLen)
	err = proc.ReadAt(account_msg, int(account), int(accountLen))
	if err != nil{
		return -1
	}

	Length = len(account_msg)
	var accountSlice []byte = account_msg[:Length - 1]
	if account_msg[Length - 1] != 0 {
		accountSlice = append(accountSlice, account_msg[Length - 1])
	}

	// Get token creator
	tokenInfo, err := ws.state.GetTokenInfo(string(nameSlice))
	if err != nil {
		return -4
	}

	// this api must invoke by token's creator contract
	if ws.action.ContractAccount != tokenInfo.Creator {
		return -5
	}

	if err := ws.state.AccountSubBalance(common.NameToIndex(string(accountSlice)), string(nameSlice), big.NewInt(int64(amount))); err != nil {
		return -2
	}

	// generate trx receipt
	var acc state.Account
	var deltaByte []byte
	num := new(big.Int).Sub(big.NewInt(0), big.NewInt(amount))
	// if one contract action invoke this api some times to modify the same account's balance
	delta, ok := ws.context.Tc.AccountDelta[string(accountSlice)]
	if ok {	// some time
		err = acc.Deserialize(delta)
		if err != nil {
			return -4
		}

		// if accountDelta had existed, check if token existed in accountDelta
		balance, ok := acc.Tokens[string(nameSlice)]
		if ok {
			balance.Balance = new(big.Int).Add(balance.Balance, num)
			acc.Tokens[string(nameSlice)] = balance
		} else {	// if may happen in ICO
			newBalance := state.Token{
				Name:		string(nameSlice),
				Balance:	big.NewInt(amount),
			}
			acc.Tokens[string(nameSlice)] = newBalance
		}
	} else {	// first times
		acc = state.Account{
			Tokens:			make(map[string]state.Token),
			Index:			common.NameToIndex(string(accountSlice)),
		}

		balance := state.Token{
			Name:		string(nameSlice),
			Balance:	num,
		}
		acc.Tokens[string(nameSlice)] = balance
	}

	deltaByte, err = acc.Serialize()
	if err != nil {
		return -4
	}

	ws.context.Tc.AccountDelta[string(accountSlice)] = deltaByte
	var flag int = 0;
	for _, accName := range ws.context.Tc.Accounts {
		if accName == string(accountSlice) {
			flag = 1
		}
	}

	if flag == 0 {
		ws.context.Tc.Accounts = append(ws.context.Tc.Accounts, string(accountSlice))
	}

	return 0
}
