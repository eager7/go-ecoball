package commands

import (
	"bytes"
	"strconv"
	"fmt"
	"encoding/json"
	"reflect"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
)


//func buildWasmContract(params []interface{}) ([]byte, error) {
//	args, err := buildWasmContractParam(params)
//	if err != nil {
//		return nil, err
//	}
//
//	bs, err := json.Marshal(Args{args})
//	if err != nil {
//		return nil, err
//	}
//
//	return bs, nil
//}
//
//
////for wasm vm
////build param bytes for wasm contract
//func buildWasmContractParam(params []interface{}) ([]Param, error) {
//	args := make([]Param, len(params))
//
//	for i, param := range params {
//		switch v := param.(type) {
//		case string:
//			arg := Param{Ptype: "string", Pval: param.(string)}
//			args[i] = arg
//		case int:
//			arg := Param{Ptype: "int", Pval: strconv.Itoa(param.(int))}
//			args[i] = arg
//		case int64:
//			arg := Param{Ptype: "int64", Pval: strconv.FormatInt(param.(int64), 10)}
//			args[i] = arg
//		case []int:
//			bf := bytes.NewBuffer(nil)
//			array := param.([]int)
//			for i, tmp := range array {
//				bf.WriteString(strconv.Itoa(tmp))
//				if i != len(array)-1 {
//					bf.WriteString(",")
//				}
//			}
//			arg := Param{Ptype: "int_array", Pval: bf.String()}
//			args[i] = arg
//		case []int64:
//			bf := bytes.NewBuffer(nil)
//			array := param.([]int64)
//			for i, tmp := range array {
//				bf.WriteString(strconv.FormatInt(tmp, 10))
//				if i != len(array)-1 {
//					bf.WriteString(",")
//				}
//			}
//			arg := Param{Ptype: "int_array", Pval: bf.String()}
//			args[i] = arg
//		default:
//			object := reflect.ValueOf(v)
//			kind := object.Kind().String()
//			if kind == "ptr" {
//				object = object.Elem()
//				kind = object.Kind().String()
//			}
//			switch kind {
//			case "slice":
//				ps := make([]interface{}, 0)
//				for i := 0; i < object.Len(); i++ {
//					ps = append(ps, object.Index(i).Interface())
//				}
//				argbytes, err := buildWasmContractParam(ps)
//				if err != nil {
//					return nil, err
//				}
//
//				bf := bytes.NewBuffer(nil)
//				for i, tmp := range argbytes {
//					bf.WriteString("type:" + tmp.Ptype + "," + "value:" + tmp.Pval)
//					if i != len(argbytes)-1 {
//						bf.WriteString(",")
//					}
//				}
//				arg := Param{Ptype: "int_array", Pval: bf.String()}
//				args[i] = arg
//
//				bs, err := json.Marshal(Args{argbytes})
//				if err != nil {
//					return nil, err
//				}
//				fmt.Printf("%s\n", bs)
//			//case "struct":
//			//	builder.EmitPushInteger(big.NewInt(0))
//			//	builder.Emit(neovm.NEWSTRUCT)
//			//	builder.Emit(neovm.TOALTSTACK)
//			//	for i := 0; i < object.NumField(); i++ {
//			//		field := object.Field(i)
//			//		builder.Emit(neovm.DUPFROMALTSTACK)
//			//		err := BuildNeoVMParam(builder, []interface{}{field.Interface()})
//			//		if err != nil {
//			//			return nil, err
//			//		}
//			//		builder.Emit(neovm.APPEND)
//			//	}
//			//	builder.Emit(neovm.FROMALTSTACK)
//			default:
//				return nil, fmt.Errorf("unsupported param:%s", v)
//			}
//		}
//	}
//
//	//bs, err := json.Marshal(Args{args})
//	//if err != nil {
//	//	return nil, err
//	//}
//	//return bs, nil
//	return args, nil
//}

func Int32ToString(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

func BuildWasmContractParam(params []interface{}) ([]byte, error) {
	args := make([]wasmservice.ParamTV, len(params))

	for i, param := range params {
		switch v := param.(type) {
		case string:
			arg := wasmservice.ParamTV{Ptype: "string", Pval: param.(string)}
			args[i] = arg
		case int:
			arg := wasmservice.ParamTV{Ptype: "int", Pval: strconv.Itoa(param.(int))}
			args[i] = arg
		case int64:
			arg := wasmservice.ParamTV{Ptype: "int64", Pval: strconv.FormatInt(param.(int64), 10)}
			args[i] = arg
		case []int:
			bf := bytes.NewBuffer(nil)
			array := param.([]int)
			for i, tmp := range array {
				bf.WriteString(strconv.Itoa(tmp))
				if i != len(array)-1 {
					bf.WriteString(",")
				}
			}
			arg := wasmservice.ParamTV{Ptype: "int32_array", Pval: bf.String()}
			args[i] = arg
		case []int64:
			bf := bytes.NewBuffer(nil)
			array := param.([]int64)
			for i, tmp := range array {
				bf.WriteString(strconv.FormatInt(tmp, 10))
				if i != len(array)-1 {
					bf.WriteString(",")
				}
			}
			arg := wasmservice.ParamTV{Ptype: "int64_array", Pval: bf.String()}
			args[i] = arg
		default:
			object := reflect.ValueOf(v)
			kind := object.Kind().String()
			if kind == "ptr" {
				object = object.Elem()
				kind = object.Kind().String()
			}
			switch kind {
				case "slice":
					ps := make([]interface{}, 0)
					for i := 0; i < object.Len(); i++ {
						ps = append(ps, object.Index(i).Interface())
					}
					argbytes, err := BuildWasmContractParam(ps)
					if err != nil {
						return nil, err
					}

					arg := wasmservice.ParamTV{Ptype: "slice", Pval: string(argbytes[:])}
					args[i] = arg
				//case "struct":
				//	ps := make([]interface{}, 0)
				//	for j := 0; j < object.NumField(); j++ {
				//		ps = append(ps, object.Index(i).Interface())
				//	}
				//
				//	argbytes, err := buildWasmContractParam([]interface{}{field.Interface()})
				//	if err != nil {
				//		return nil, err
				//	}
				//
				//	arg := wasmservice.Param{Ptype: "slice", Pval: string(argbytes[:])}
				//	args[i] = arg
			default:
				return nil, fmt.Errorf("unsupported param:%s", v)
			}
		}
	}

	bs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	return bs, nil
}