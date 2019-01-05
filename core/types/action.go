package types

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
)

type PermissionLevel struct {
	Actor      common.AccountName `json:"actor"`
	Permission string             `json:"permission"`
}

type Action struct {
	ContractAccount common.AccountName `json:"account"`
	Permission      PermissionLevel    `json:"permission"`
	Payload         EcoMessage         `json:"payload"`
	Console         string             `json:"console"`
}

func NewAction(tx *Transaction) *Action {
	return &Action{
		ContractAccount: tx.Addr,
		Permission:      PermissionLevel{tx.From, tx.Permission},
		Payload:         tx.Payload,
	}
}

func NewSimpleAction(contract string, permission PermissionLevel, payload EcoMessage) (*Action, error) {
	action := &Action{
		ContractAccount: common.NameToIndex(contract),
		Permission:      permission,
		Payload:         payload,
	}

	return action, nil
}

func (act *Action) Print() {
	fmt.Println("account: ", act.ContractAccount)

	invoke, ok := act.Payload.GetInstance().(*InvokeInfo)
	if !ok {
		fmt.Println("transaction type error[invoke]")
		return
	}

	fmt.Println("action: ", string(invoke.Method))
	fmt.Println("param: ", invoke.Param)
}
