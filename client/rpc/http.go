// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball.
//
// The go-ecoball is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball. If not, see <http://www.gnu.org/licenses/>.

package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ecoball/go-ecoball/client/common"
)

//new request
func newRequest(method, resource, address string, body io.Reader) (req *http.Request, err error) {
	url := address + resource

	req, err = http.NewRequest(method, url, body)
	return
}

//post raw data
func postRawResponse(resource, address string, data interface{}) ([]byte, error) {

	s, _ := json.Marshal(data)
	b := bytes.NewBuffer(s)
	req, err := newRequest("POST", resource, address, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if nil != err {
		return nil, err
	}
	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	if 404 == res.StatusCode {
		return nil, errors.New(res.Status)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, readAPIError(res.Body)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("API call not recognized: " + resource)
	}

	if res.StatusCode == http.StatusNoContent {
		// no reason to read the response
		return []byte{}, nil
	}
	return ioutil.ReadAll(res.Body)
}

//post method
func post(resource, address string, data, obj interface{}) error {
	body, err := postRawResponse(resource, address, data)
	if nil != err {
		return err
	}

	if obj == nil {
		return nil
	}

	// Decode response
	buf := bytes.NewBuffer(body)
	err = json.NewDecoder(buf).Decode(obj)
	return err
}

//get raw data
func getRawResponse(resource, address string) ([]byte, error) {
	req, err := newRequest("GET", resource, address, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if nil != err {
		return nil, err
	}
	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	if 404 == res.StatusCode {
		return nil, errors.New(res.Status)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, readAPIError(res.Body)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("API call not recognized: " + resource)
	}

	if res.StatusCode == http.StatusNoContent {
		// no reason to read the response
		return []byte{}, nil
	}
	return ioutil.ReadAll(res.Body)
}

//get method
func get(resource, address string, obj interface{}) error {
	data, err := getRawResponse(resource, address)
	if err != nil {
		return err
	}
	if obj == nil {
		// No need to decode response
		return nil
	}

	// Decode response
	buf := bytes.NewBuffer(data)
	err = json.NewDecoder(buf).Decode(obj)
	return err
}

//post
func NodePost(resource string, data, obj interface{}) error {
	return post(resource, common.RpcAddress(), data, obj)
}

func WalletPost(resource string, data, obj interface{}) error {
	return post(resource, common.WalletRpcAddress(), data, obj)
}

//get
func NodeGet(resource string, obj interface{}) error {
	return get(resource, common.RpcAddress(), obj)
}

func WalletGet(resource string, obj interface{}) error {
	return get(resource, common.WalletRpcAddress(), obj)
}
