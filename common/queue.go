// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package common

import "container/list"

type Queue struct {
	l *list.List
}

func NewQueue() *Queue {
	return &Queue{l: list.New()}
}

func (q *Queue) Pop() interface{} {
	if q.l.Back() != nil {
		return q.l.Remove(q.l.Back())
	}

	return nil
}

func (q *Queue) Push(data interface{}) {
	q.l.PushFront(data)
}

func (q *Queue) GetHeadValue() interface{} {
	if q.l.Front() != nil {
		return q.l.Front().Value
	}
	return nil
}

func (q *Queue) GetHead() interface{} {
	return q.l.Front()
}

func (q *Queue) Length() int {
	return q.l.Len()
}
