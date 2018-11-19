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
