/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"github.com/ecoball/go-ecoball/smartcontract/wasmservice"
)

func TestParseRawParamsArray(t *testing.T) {
	rawParamStr := "string:foo,[int:0,[bool:true,string:bar],bool:false]"
	res, size, err := parseRawParamsString(rawParamStr)
	if err != nil {
		t.Errorf("TestParseArrayParams error:%s", err)
		return
	}
	if size != len(rawParamStr) {
		t.Errorf("TestParseArrayParams size:%d != %d", size, len(rawParamStr))
		return
	}
	expect := []interface{}{"string:foo", []interface{}{"int:0", []interface{}{"bool:true", "string:bar"}, "bool:false"}}
	ok, err := arrayEqual(res, expect)
	if err != nil {
		t.Errorf("TestParseArrayParams error:%s", err)
		return
	}
	if !ok {
		t.Errorf("TestParseArrayParams faild, res:%s != %s", res, expect)
		return
	}
}

func TestParseParams(t *testing.T) {
	testByteArray := []byte("HelloWorld")
	testByteArrayParam := hex.EncodeToString(testByteArray)
	rawParamStr := "bytearray:" + testByteArrayParam + ",string:foo,[int:0,[bool:true,string:bar],bool:false]"
	params, err := ParseParams(rawParamStr)
	if err != nil {
		t.Errorf("TestParseParams error:%s", err)
		return
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal error:%s", err)
		return
	}
	fmt.Printf("%s\n", data)

	expect := []interface{}{testByteArray, "foo", []interface{}{0, []interface{}{true, "bar"}, false}}
	ok, err := arrayEqual(params, expect)
	if err != nil {
		t.Errorf("TestParseParams error:%s", err)
		return
	}
	if !ok {
		t.Errorf("TestParseParams faild, res:%s != %s", params, expect)
		return
	}
}

func TestParseWasmParams(t *testing.T) {

	//rawParamStr := "string:foo,[int:0,[int:1,string:bar],int:0]"
	//rawParamStr := "string:foo,[int:1,int:0]"
	rawParamStr := "string:foo"
	params, err := ParseParams(rawParamStr)
	if err != nil {
		t.Errorf("TestParseParams error:%s", err)
		return
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal error:%s", err)
		return
	}
	fmt.Printf("%s\n", data)

	//expect := []interface{}{"foo", []interface{}{0, []interface{}{1, "bar"}, 0}}
	//expect := []interface{}{"foo", []interface{}{1, 0}}
	expect := []interface{}{"foo"}
	ok, err := arrayEqual(params, expect)
	if err != nil {
		t.Errorf("TestParseParams error:%s", err)
		return
	}
	if !ok {
		t.Errorf("TestParseParams faild, res:%s != %s", params, expect)
		return
	}

	argbytes, err := BuildWasmContractParam(params)
	if err != nil {
		t.Errorf("build wasm contract param failed:%s", err)
		return
	}
	fmt.Printf("%s\n", argbytes)

	//args := make([]Param, len(params))
	var args []wasmservice.ParamTV
	err1 := json.Unmarshal(argbytes, &args)
	if err1 != nil {
		t.Errorf("TestParseParams error:%s", err1)
		return
	}
	fmt.Printf("%+v\n", args)

	//js, err := simplejson.NewJson(argbytes)
	//fmt.Printf("%s\n", *js)


	argbytes1, err := BuildWasmContractParam(params)
	if err != nil {
		t.Errorf("build wasm contract param failed:%s", err)
		return
	}
	fmt.Printf("%s\n", argbytes1)

	return
}

//func TestReflect(t *testing.T) {
//	var perm []interface{}{"threshold": 1,"keys": [{"key": "EOS5ANFfT6y5ubcCzPWGUYGUR6U13KKS5T3XwE32Uo7j8Ruarfi1x","weight": 1}],"accounts": [{"permission":{"actor":"hellozhongxh","permission":"active"},"weight":1}]}
//
//	t := reflect.TypeOf(perm)
//	v := reflect.ValueOf(perm)
//	for i := 0; i < v.NumField(); i++ {
//		if v.Field(i).CanInterface() {  //判断是否为可导出字段
//
//			//判断是否是嵌套结构
//			if v.Field(i).Type().Kind() == reflect.Struct{
//				structField := v.Field(i).Type()
//				for j :=0 ; j< structField.NumField(); j++ {
//					fmt.Printf("%s %s = %v -tag:%s \n",
//						structField.Field(j).Name,
//						structField.Field(j).Type,
//						v.Field(i).Field(j).Interface(),
//						structField.Field(j).Tag)
//				}
//				continue
//			}
//
//			fmt.Printf("%s %s = %v -tag:%s \n",
//				t.Field(i).Name,
//				t.Field(i).Type,
//				v.Field(i).Interface(),
//				t.Field(i).Tag)
//		}
//
//	}
//
//}

func arrayEqual(a1, a2 []interface{}) (bool, error) {
	data1, err := json.Marshal(a1)
	if err != nil {
		return false, fmt.Errorf("json.Marshal:%s error:%s", a1, err)
	}
	data2, err := json.Marshal(a2)
	if err != nil {
		return false, fmt.Errorf("json.Marshal:%s error:%s", a2, err)
	}
	return string(data1) == string(data2), nil
}
