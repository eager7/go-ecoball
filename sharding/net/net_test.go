package net

import "testing"

func TestNet_CalcCrossShardIndex(t *testing.T) {
	bSend, begin, count := CalcCrossShardIndex(0, 1, 1)
	if !bSend || begin != 0 || count != 1 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 2, 1)
	if !bSend || begin != 0 || count != 1 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 2, 2)
	if !bSend || begin != 0 || count != 1 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 2, 1)
	if bSend {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 2, 2)
	if !bSend || begin != 1 || count != 1 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 2, 3)
	if !bSend || begin != 0 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 2, 3)
	if !bSend || begin != 2 || count != 1 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 2, 4)
	if !bSend || begin != 0 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 2, 4)
	if !bSend || begin != 2 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 2, 5)
	if !bSend || begin != 0 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 2, 5)
	if !bSend || begin != 3 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 2, 6)
	if !bSend || begin != 0 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 2, 6)
	if !bSend || begin != 3 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 3, 6)
	if !bSend || begin != 0 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 3, 6)
	if !bSend || begin != 2 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(2, 3, 6)
	if !bSend || begin != 4 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 3, 7)
	if !bSend || begin != 0 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 3, 7)
	if !bSend || begin != 3 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(2, 3, 7)
	if !bSend || begin != 5 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 3, 8)
	if !bSend || begin != 0 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 3, 8)
	if !bSend || begin != 3 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(2, 3, 8)
	if !bSend || begin != 6 || count != 2 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 3, 9)
	if !bSend || begin != 0 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 3, 9)
	if !bSend || begin != 3 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(2, 3, 9)
	if !bSend || begin != 6 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 3, 10)
	if !bSend || begin != 0 || count != 4 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 3, 10)
	if !bSend || begin != 4 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(2, 3, 10)
	if !bSend || begin != 7 || count != 3 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(0, 3, 2)
	if !bSend || begin != 0 || count != 1 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(1, 3, 2)
	if !bSend || begin != 1 || count != 1 {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(2, 3, 2)
	if bSend {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(8, 9, 2)
	if bSend {
		t.Fatal(bSend, begin, count)
	}

	bSend, begin, count = CalcCrossShardIndex(7, 9, 2)
	if bSend {
		t.Fatal(bSend, begin, count)
	}

	return
}

func TestCalcGossipIndex(t *testing.T) {
	indexs := CalcGossipIndex(1, 0)
	if indexs != nil {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(1, 1)
	if indexs != nil {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(2, 0)
	if len(indexs) != 1 || indexs[0] != 1 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(2, 1)
	if len(indexs) != 1 || indexs[0] != 0 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(3, 0)
	if len(indexs) != 1 || indexs[0] != 1 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(3, 1)
	if len(indexs) != 1 || indexs[0] != 2 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(3, 2)
	if len(indexs) != 1 || indexs[0] != 0 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(4, 0)
	if len(indexs) != 2 || indexs[0] != 1 || indexs[1] != 2 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(4, 1)
	if len(indexs) != 2 || indexs[0] != 2 || indexs[1] != 3 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(4, 2)
	if len(indexs) != 2 || indexs[0] != 3 || indexs[1] != 0 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(4, 3)
	if len(indexs) != 2 || indexs[0] != 0 || indexs[1] != 1 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(5, 0)
	if len(indexs) != 2 || indexs[0] != 1 || indexs[1] != 2 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(5, 1)
	if len(indexs) != 2 || indexs[0] != 2 || indexs[1] != 3 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(5, 2)
	if len(indexs) != 2 || indexs[0] != 3 || indexs[1] != 4 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(5, 3)
	if len(indexs) != 2 || indexs[0] != 4 || indexs[1] != 0 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(5, 4)
	if len(indexs) != 2 || indexs[0] != 0 || indexs[1] != 1 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(5, 5)
	if indexs != nil {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(8, 0)
	if len(indexs) != 2 || indexs[0] != 1 || indexs[1] != 2 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(8, 3)
	if len(indexs) != 2 || indexs[0] != 4 || indexs[1] != 5 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(8, 6)
	if len(indexs) != 2 || indexs[0] != 7 || indexs[1] != 0 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(8, 7)
	if len(indexs) != 2 || indexs[0] != 0 || indexs[1] != 1 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(9, 0)
	if len(indexs) != 3 || indexs[0] != 1 || indexs[1] != 2 || indexs[2] != 3 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(9, 5)
	if len(indexs) != 3 || indexs[0] != 6 || indexs[1] != 7 || indexs[2] != 8 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(9, 8)
	if len(indexs) != 3 || indexs[0] != 0 || indexs[1] != 1 || indexs[2] != 2 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(10, 8)
	if len(indexs) != 3 || indexs[0] != 9 || indexs[1] != 0 || indexs[2] != 1 {
		t.Fatal(indexs)
	}

	indexs = CalcGossipIndex(25, 8)
	if len(indexs) != 5 || indexs[0] != 9 || indexs[1] != 10 || indexs[2] != 11 || indexs[3] != 12 || indexs[4] != 13 {
		t.Fatal(indexs)
	}
}
