package abi

import (
	"encoding/json"
	"strconv"
	"fmt"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/common/config"
	"github.com/ecoball/go-ecoball/common"
)

//type Name string
//type AccountName Name
//type PermissionName Name
//type ActionName Name
//type TableName Name
//type ScopeName Name

//type HexBytes []byte

type Extension struct {
	Type uint16   `json:"type"`
	Data HexBytes `json:"data"`
}
//
//// NOTE: there's also a new ExtendedSymbol (which includes the contract (as AccountName) on which it is)
//type Symbol struct {
//	Precision uint8
//	Symbol    string
//}
//
//// NOTE: there's also ExtendedAsset which is a quantity with the attached contract (AccountName)
//type Asset struct {
//	Amount int64
//	Symbol
//}


// see: libraries/chain/contracts/abi_serializer.cpp:53...
// see: libraries/chain/include/eosio/chain/contracts/types.hpp:100
type ABI struct {
	Version          string            `json:"version"`
	Types            []ABIType         `json:"types,omitempty"`
	Structs          []StructDef       `json:"structs,omitempty"`
	Actions          []ActionDef       `json:"actions,omitempty"`
	Tables           []TableDef        `json:"tables,omitempty"`
	RicardianClauses []ClausePair      `json:"ricardian_clauses,omitempty"`
	ErrorMessages    []ABIErrorMessage `json:"error_messages,omitempty"`
	Extensions       []*Extension      `json:"abi_extensions,omitempty"`
}

type ABIType struct {
	NewTypeName string `json:"new_type_name"`
	Type        string `json:"type"`
}

type StructDef struct {
	Name   string     `json:"name"`
	Base   string     `json:"base"`
	Fields []FieldDef `json:"fields,omitempty"`
}

type FieldDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ActionDef struct {
	Name              ActionName `json:"name"`
	Type              string     `json:"type"`
	RicardianContract string     `json:"ricardian_contract"`
}

// TableDef defines a table. See libraries/chain/include/eosio/chain/contracts/types.hpp:78
type TableDef struct {
	Name      TableName `json:"name"`
	IndexType string    `json:"index_type"`
	KeyNames  []string  `json:"key_names,omitempty"`
	KeyTypes  []string  `json:"key_types,omitempty"`
	Type      string    `json:"type"`
}

// ClausePair represents clauses, related to Ricardian Contracts.
type ClausePair struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

type ABIErrorMessage struct {
	Code    uint64 `json:"error_code"`
	Message string `json:"error_msg"`
}

type Param struct {
	Ptype string `json:"type"`
	Pval  string `json:"value"`
}

var log = elog.NewLogger("commands", elog.DebugLog)

func CheckParam(abiDef ABI, method string, arg []byte) ([]byte, error){
	var fields []FieldDef
	for _, action := range abiDef.Actions {
		// first: find method
		if string(action.Name) == method {
			//fmt.Println("find ", method)
			for _, struction := range abiDef.Structs {
				// second: find struct
				if struction.Name == action.Type {
					fields = struction.Fields
				}
			}
			break
		}
	}

	if fields == nil {
		return nil, errors.New("can not find method " + method)
	}

	args := make([]Param, len(fields))
	if(arg[0] == '{') {		// Key-Value structure
		var f interface{}

		if err := json.Unmarshal(arg, &f); err != nil {
			return nil, err
		}

		m := f.(map[string]interface{})

		if len(fields) != len(m) {
			return nil, errors.New("args size error" )
		}

		for i, field := range fields {
			v := m[field.Name]
			if v != nil {
				args[i].Ptype = field.Type
				switch vv := v.(type) {
				case string:
					switch field.Type {
					case "string","account_name","asset":
						args[i].Pval = vv
					case "int8":
						const INT8_MAX = int8(^uint8(0) >> 1)
						const INT8_MIN = ^INT8_MAX
						a, err := strconv.ParseInt(vv, 10, 8)
						if err != nil {
							return nil, errors.New(fmt.Sprintln(vv, "is out of int8 range"))
						}
						if a >= int64(INT8_MIN) && a <= int64(INT8_MAX) {
							args[i].Pval = vv
						} else {
							return nil, errors.New(fmt.Sprintln(vv, "is out of int8 range"))
						}
					case "int16":
						const INT16_MAX = int16(^uint16(0) >> 1)
						const INT16_MIN = ^INT16_MAX
						a, err := strconv.ParseInt(vv, 10, 16)
						if err != nil {
							return nil, errors.New(fmt.Sprintln(vv, "is out of int16 range"))
						}
						if a >= int64(INT16_MIN) && a <= int64(INT16_MAX) {
							args[i].Pval = vv
						} else {
							return nil, errors.New(fmt.Sprintln(vv, "is out of int16 range"))
						}
					case "int32":
						const INT32_MAX = int32(^uint32(0) >> 1)
						const INT32_MIN = ^INT32_MAX
						a, err := strconv.ParseInt(vv, 10, 32)
						if err != nil {
							return nil, errors.New(fmt.Sprintln(vv, "is out of int32 range"))
						}
						if a >= int64(INT32_MIN) && a <= int64(INT32_MAX) {
							args[i].Pval = vv
						} else {
							return nil, errors.New(fmt.Sprintln(vv, "is out of int32 range"))
						}
					case "int64":
						const INT64_MAX = int64(^uint64(0) >> 1)
						const INT64_MIN = ^INT64_MAX
						a, err := strconv.ParseInt(vv, 10, 64)
						if err != nil {
							return nil, errors.New(fmt.Sprintln(vv, "is out of int64 range"))
						}
						if a >= INT64_MIN && a <= INT64_MAX {
							args[i].Pval = vv
						} else {
							return nil, errors.New(fmt.Sprintln(vv, "is out of int64 range"))
						}

					case "uint8":
						const UINT8_MIN uint8 = 0
						const UINT8_MAX = ^uint8(0)
						a, err := strconv.ParseUint(vv, 10, 8)
						if err != nil {
							return nil, errors.New(fmt.Sprintln(vv, "is out of uint8 range"))
						}
						if a >= uint64(UINT8_MIN) && a <= uint64(UINT8_MAX) {
							args[i].Pval = vv
						} else {
							return nil, errors.New(fmt.Sprintln(vv, "is out of uint8 range"))
						}
					case "uint16":
						const UINT16_MIN uint16 = 0
						const UINT16_MAX = ^uint16(0)
						a, err := strconv.ParseUint(vv, 10, 16)
						if err != nil {
							return nil, errors.New(fmt.Sprintln(vv, "is out of uint16 range"))
						}
						if a >= uint64(UINT16_MIN) && a <= uint64(UINT16_MAX) {
							args[i].Pval = vv
						} else {
							return nil, errors.New(fmt.Sprintln(vv, "is out of uint16 range"))
						}
					case "uint32":
						const UINT32_MIN uint32 = 0
						const UINT32_MAX = ^uint32(0)
						a, err := strconv.ParseUint(vv, 10, 32)
						if err != nil {
							return nil, errors.New(fmt.Sprintln(vv, "is out of uint32 range"))
						}
						if a >= uint64(UINT32_MIN) && a <= uint64(UINT32_MAX) {
							args[i].Pval = vv
						} else {
							return nil, errors.New(fmt.Sprintln(vv, "is out of uint32 range"))
						}
					case "uint64":
						const UINT64_MIN uint64 = 0
						const UINT64_MAX = ^uint64(0)
						a, err := strconv.ParseUint(vv, 10, 64)
						if err != nil {
							return nil, errors.New(fmt.Sprintln(vv, "is out of uint64 range"))
						}
						if a >= UINT64_MIN && a <= UINT64_MAX {
							args[i].Pval = vv
						} else {
							return nil, errors.New(fmt.Sprintln(vv, "is out of uint64 range"))
						}

					default:
						return nil, errors.New(fmt.Sprintln("can't match abi struct field type ", field.Type))
					}

					fmt.Println(field.Name, "is ", field.Type, "", vv)

				default:
					return nil, errors.New(fmt.Sprintln("can't match abi struct field type: ", v))
				}
			} else {
				return nil, errors.New("can't match abi struct field name:  " + field.Name)
			}

		}
	} else {		// Only Value structure
		var f []string

		if err := json.Unmarshal(arg, &f); err != nil {
			return nil, err
		}

		if len(fields) != len(f) {
			return nil, errors.New("args size error" )
		}

		for i, field := range fields {
			args[i].Ptype = field.Type
			vv := f[i]
			switch field.Type {
			case "string","account_name","asset":
				args[i].Pval = vv
			case "int8":
				const INT8_MAX = int8(^uint8(0) >> 1)
				const INT8_MIN = ^INT8_MAX
				a, err := strconv.ParseInt(vv, 10, 8)
				if err != nil {
					return nil, errors.New(fmt.Sprintln(vv, "is out of int8 range"))
				}
				if a >= int64(INT8_MIN) && a <= int64(INT8_MAX) {
					args[i].Pval = vv
				} else {
					return nil, errors.New(fmt.Sprintln(vv, "is out of int8 range"))
				}
			case "int16":
				const INT16_MAX = int16(^uint16(0) >> 1)
				const INT16_MIN = ^INT16_MAX
				a, err := strconv.ParseInt(vv, 10, 16)
				if err != nil {
					return nil, errors.New(fmt.Sprintln(vv, "is out of int16 range"))
				}
				if a >= int64(INT16_MIN) && a <= int64(INT16_MAX) {
					args[i].Pval = vv
				} else {
					return nil, errors.New(fmt.Sprintln(vv, "is out of int16 range"))
				}
			case "int32":
				const INT32_MAX = int32(^uint32(0) >> 1)
				const INT32_MIN = ^INT32_MAX
				a, err := strconv.ParseInt(vv, 10, 32)
				if err != nil {
					return nil, errors.New(fmt.Sprintln(vv, "is out of int32 range"))
				}
				if a >= int64(INT32_MIN) && a <= int64(INT32_MAX) {
					args[i].Pval = vv
				} else {
					return nil, errors.New(fmt.Sprintln(vv, "is out of int32 range"))
				}
			case "int64":
				const INT64_MAX = int64(^uint64(0) >> 1)
				const INT64_MIN = ^INT64_MAX
				a, err := strconv.ParseInt(vv, 10, 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintln(vv, "is out of int64 range"))
				}
				if a >= INT64_MIN && a <= INT64_MAX {
					args[i].Pval = vv
				} else {
					return nil, errors.New(fmt.Sprintln(vv, "is out of int64 range"))
				}

			case "uint8":
				const UINT8_MIN uint8 = 0
				const UINT8_MAX = ^uint8(0)
				a, err := strconv.ParseUint(vv, 10, 8)
				if err != nil {
					return nil, errors.New(fmt.Sprintln(vv, "is out of uint8 range"))
				}
				if a >= uint64(UINT8_MIN) && a <= uint64(UINT8_MAX) {
					args[i].Pval = vv
				} else {
					return nil, errors.New(fmt.Sprintln(vv, "is out of uint8 range"))
				}
			case "uint16":
				const UINT16_MIN uint16 = 0
				const UINT16_MAX = ^uint16(0)
				a, err := strconv.ParseUint(vv, 10, 16)
				if err != nil {
					return nil, errors.New(fmt.Sprintln(vv, "is out of uint16 range"))
				}
				if a >= uint64(UINT16_MIN) && a <= uint64(UINT16_MAX) {
					args[i].Pval = vv
				} else {
					return nil, errors.New(fmt.Sprintln(vv, "is out of uint16 range"))
				}
			case "uint32":
				const UINT32_MIN uint32 = 0
				const UINT32_MAX = ^uint32(0)
				a, err := strconv.ParseUint(vv, 10, 32)
				if err != nil {
					return nil, errors.New(fmt.Sprintln(vv, "is out of uint32 range"))
				}
				if a >= uint64(UINT32_MIN) && a <= uint64(UINT32_MAX) {
					args[i].Pval = vv
				} else {
					return nil, errors.New(fmt.Sprintln(vv, "is out of uint32 range"))
				}
			case "uint64":
				const UINT64_MIN uint64 = 0
				const UINT64_MAX = ^uint64(0)
				a, err := strconv.ParseUint(vv, 10, 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintln(vv, "is out of uint64 range"))
				}
				if a >= UINT64_MIN && a <= UINT64_MAX {
					args[i].Pval = vv
				} else {
					return nil, errors.New(fmt.Sprintln(vv, "is out of uint64 range"))
				}

			default:
				return nil, errors.New(fmt.Sprintln("can't match abi struct field type ", field.Type))
			}

			fmt.Println(field.Name, "is ", field.Type, "", vv)
		}
	}

	bs, err := json.Marshal(args)
	if err != nil {
		return nil, errors.New("json.Marshal failed")
	}
	return bs, nil
}

func GetContractTable(contractName string, accountName string, abiDef ABI, tableName string) ([]byte, error){

	var fields []FieldDef
	for _, table := range abiDef.Tables {
		if string(table.Name) == tableName {
			for _, struction := range abiDef.Structs {
				if struction.Name == table.Type {
					fields = struction.Fields
				}
			}
		}
	}

	if fields == nil {
		return nil, errors.New("can not find struct of table  " + tableName)
	}

	type TokenStat struct {
		Supply 		int 		`json:"supply"`
		MaxSupply 	int 		`json:"max_supply"`
		Issuer 		[16]byte	`json:"issuer"`
		Token 		[8]byte		`json:"token"`
	}

	type Account struct {
		Balance 	int			`json:"balance"`
		Acc			[16]byte	`json:"account"`
		Token 		[8]byte		`json:"token"`
	}

	table := make(map[string]string, len(fields))

	for i, _ := range fields {
		key := []byte(fields[i].Name)
		if fields[i].Name == "balance" {	// only for token contract, because KV struct can't support
			key = []byte(accountName)
		} else {
			key = append(key, 0)		// C lang string end with 0
		}
		storage, err := ledger.L.StoreGet(config.ChainHash, common.NameToIndex(contractName), key)
		if err != nil {
			return nil, errors.New("can not get store " + fields[i].Name)
		}

		fmt.Println(fields[i].Name + ": " + string(storage))
		table[fields[i].Name] = string(storage)
	}

	js, _ := json.Marshal(table)
	fmt.Println("json format: ", string(js))

	return nil, nil
}