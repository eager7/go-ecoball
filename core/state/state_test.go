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

package state_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/core/state"
	"github.com/ecoball/go-ecoball/core/store"
	"math/big"
	"testing"
)

func TestStateNew(t *testing.T) {
	root := common.HexToHash("0x797ddb4c4b8cfcb353989acf8e41b5d425003fb0ae3d89e754658fe264e8dab5")
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexAcc := common.NameToIndex("pct")
	indexToken := state.AbaToken
	s, err := state.NewState("/tmp/state", root)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Trie Root:", s.GetHashRoot().HexString())

	balance, err := s.AccountGetBalance(indexAcc, indexToken)
	if err != nil {
		fmt.Println("get balance error:", err)
		if _, err := s.AddAccount(indexAcc, addr); err != nil {
			t.Fatal(err)
		}
	} else {
		fmt.Println("Value From:", balance)
	}
	value := new(big.Int).SetUint64(100)
	if err := s.AccountAddBalance(indexAcc, indexToken, value); err != nil {
		fmt.Println("Update Error:", err)
	}

	fmt.Println("Hash Root:", s.GetHashRoot().HexString())
	s.CommitToDB()
	balance, err = s.AccountGetBalance(indexAcc, indexToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Value:", balance)
}

func TestStateRoot(t *testing.T) {
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexAcc := common.NameToIndex("pct")
	indexToken := state.AbaToken
	s, err := state.NewState("/tmp/state_root", common.HexToHash("cf4bfc19264aa4bbd6898c0ef43ce5465c794fd587e622fccc19980e634cd9f2"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddAccount(indexAcc, addr); err != nil {
		t.Fatal(err)
	}
	if err := s.AccountAddBalance(indexAcc, indexToken, new(big.Int).SetInt64(100)); err != nil {
		t.Fatal(err)
	}
	value, err := s.AccountGetBalance(indexAcc, indexToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value:", value)
	fmt.Println("root:", s.GetHashRoot().HexString())

	if err := s.AccountAddBalance(indexAcc, indexToken, new(big.Int).SetInt64(150)); err != nil {
		t.Fatal(err)
	}
	value, err = s.AccountGetBalance(indexAcc, indexToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value:", value)
	fmt.Println("root:", s.GetHashRoot().HexString())
	s.CommitToDB()
}

func TestHashRoot(t *testing.T) {
	diskDb, _ := store.NewLevelDBStore("/tmp/state_hash", 0, 0)
	Db := state.NewDatabase(diskDb)

	root := common.HexToHash("c9a4c610b1068a32f091a091ee46836b5425d9dfc9dc58c32a70e2b5e5d67a7b")
	fmt.Println("open trie with root:", root.HexString())
	tree, err := Db.OpenTrie(root)
	if err != nil {
		fmt.Println("can't open trie:", err)
		tree, _ = Db.OpenTrie(common.Hash{})
	}
	fmt.Println("Root0:", tree.Hash().HexString())
	value, err := tree.TryGet([]byte("dog"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("dog1 value:", string(value))

	tree.TryUpdate([]byte("doe"), []byte("reindeer"))
	fmt.Println("root1:", tree.Hash().HexString())

	tree.TryUpdate([]byte("dog"), []byte("puppy"))
	fmt.Println("update dog to puppy, root2:", tree.Hash().HexString())

	value, err = tree.TryGet([]byte("dog"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("dog value:", string(value))

	tree.TryUpdate([]byte("dogglesworth"), []byte("cat"))
	fmt.Println("root3:", tree.Hash().HexString())

	tree.TryUpdate([]byte("dogglesworth"), []byte("cat"))
	fmt.Println("root4:", tree.Hash().HexString())

	fmt.Println("Commit DB")
	tree.Commit(nil)
	lDB := Db.TrieDB()
	lDB.Commit(tree.Hash(), true)
	hash := tree.Hash()

	tree.TryUpdate([]byte("dog"), []byte("puppy2"))
	fmt.Println("update dog to puppy2, root5:", tree.Hash().HexString())

	value, err = tree.TryGet([]byte("dog"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("dog value:", string(value))

	fmt.Println("ReOpen trie with hash:", hash.HexString())
	tree, err = Db.OpenTrie(hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("root5:", tree.Hash().HexString())

	value, err = tree.TryGet([]byte("dog"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("dog value:", string(value))
	tree.Commit(nil)
	lDB = Db.TrieDB()
	lDB.Commit(tree.Hash(), true)
	fmt.Println("root6:", tree.Hash().HexString())

}

func TestStateDB(t *testing.T) {
	addr := common.NewAddress(common.FromHex("01ca5cdd56d99a0023166b337ffc7fd0d2c42330"))
	indexAcc := common.NameToIndex("pct")
	indexToken := state.AbaToken
	s, err := state.NewState("/tmp/state_root", common.HexToHash(""))
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(s.GetHashRoot().HexString())
	s.CommitToMemory()
	if _, err := s.AddAccount(indexAcc, addr); err != nil {
		t.Fatal(err)
	}
	if err := s.AccountAddBalance(indexAcc, indexToken, new(big.Int).SetInt64(100)); err != nil {
		t.Fatal(err)
	}
	s.CommitToDB()
	value, err := s.AccountGetBalance(indexAcc, indexToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value1[100]:", value)

	acc, err := s.GetAccountByName(indexAcc)
	if err != nil {
		t.Fatal(err)
	}
	acc.AddBalance(indexToken, new(big.Int).SetUint64(300))
	//s.CommitAccount(acc)

	d, err := acc.Serialize()
	trie := s.DataBase().CopyTrie(s.Trie())
	trie.TryUpdate(common.IndexToBytes(acc.Index), d)
	trie.Commit(nil)
	fmt.Println(trie.Hash().HexString())

	s.CommitToDB()
	value, err = s.AccountGetBalance(indexAcc, indexToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value2[400]:", value)

	//if err := s.Reset(common.HexToHash("0x7aacdf43f49a61ca524a80f81546f9a059ce9c877530dccd624e8895495a5015")); err != nil {
	//	t.Fatal(err)
	//}
	value, err = s.AccountGetBalance(indexAcc, indexToken)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("value3[100]:", value)
}

func TestRootHash(t *testing.T) {
	s, err := state.NewState("/tmp/state_hash", common.HexToHash(""))
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(s.GetHashRoot().HexString())
	s.CommitToMemory()
	if err := s.Trie().TryUpdate([]byte("pct"), []byte("panchangtao")); err != nil {
		t.Fatal(err)
	}
	//if err := s.AccountAddBalance(indexAcc, indexToken, new(big.Int).SetInt64(100)); err != nil {
	//	t.Fatal(err)
	//}
	s.CommitToDB()
}