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
	"encoding/json"
	"errors"
	"io"
)

type Error struct {
	Message string `json:"message"`
}

func (err Error) Error() string {
	return err.Message
}

func readAPIError(r io.Reader) error {
	var apiErr Error
	if err := json.NewDecoder(r).Decode(&apiErr); err != nil {
		return errors.New(err.Error() + " could not read error response")
	}
	return apiErr
}

type SimpleResult struct {
	Result string `json:"result"`
}