package trie_test

import (
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/trie"
	"os"
	"testing"
)

func TestNewMptTrie(t *testing.T) {
	_ = os.RemoveAll("/tmp/tree")
	tree, err := trie.NewMptTrie("/tmp/tree", common.Hash{})
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tree.Put([]byte("key"), []byte("value")))
	value, err := tree.Get([]byte("key"))
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(string(value) == "value")
	errors.CheckErrorPanic(tree.Commit())
	root := tree.Hash()
	t.Log(root.String())
	errors.CheckErrorPanic(tree.Close())

	tree, err = trie.NewMptTrie("/tmp/tree", root)
	errors.CheckErrorPanic(err)
	value, err = tree.Get([]byte("key"))
	errors.CheckErrorPanic(err)
	t.Log(string(value))
}

func TestMpt_RollBack(t *testing.T) {
	_ = os.RemoveAll("/tmp/tree")
	tree, err := trie.NewMptTrie("/tmp/tree", common.Hash{})
	errors.CheckErrorPanic(err)
	errors.CheckErrorPanic(tree.Put([]byte("key"), []byte("value")))
	value, err := tree.Get([]byte("key"))
	errors.CheckErrorPanic(err)
	errors.CheckEqualPanic(string(value) == "value")
	errors.CheckErrorPanic(tree.Commit())
	root := tree.Hash()
	t.Log(root.String())

	errors.CheckErrorPanic(tree.Put([]byte("key"), []byte("value update")))
	value, err = tree.Get([]byte("key"))
	errors.CheckErrorPanic(err)
	if string(value) == "value" {
		t.Fatal("must be value update")
	}
	errors.CheckErrorPanic(tree.Commit())
	t.Log(tree.Hash().String())

	errors.CheckErrorPanic(tree.RollBack(root))
	value, err = tree.Get([]byte("key"))
	errors.CheckErrorPanic(err)
	if string(value) != "value" {
		t.Fatal("must be value")
	}
}

func BenchmarkNewMptTrie(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree, err := trie.NewMptTrie("/tmp/tree_b", common.Hash{})
		errors.CheckErrorPanic(err)
		errors.CheckErrorPanic(tree.Put([]byte("key"), []byte("value")))
		errors.CheckErrorPanic(tree.Close())
	}
}

func BenchmarkMpt_RollBack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree, err := trie.NewMptTrie("/tmp/tree_ben", common.Hash{})
		errors.CheckErrorPanic(err)
		errors.CheckErrorPanic(tree.Put([]byte("key"), []byte("value")))
		value, err := tree.Get([]byte("key"))
		errors.CheckErrorPanic(err)
		errors.CheckEqualPanic(string(value) == "value")
		errors.CheckErrorPanic(tree.Commit())
		root := tree.Hash()

		errors.CheckErrorPanic(tree.Put([]byte("key"), []byte("value update")))
		value, err = tree.Get([]byte("key"))
		errors.CheckErrorPanic(err)
		if string(value) == "value" {
			b.Fatal("must be value update")
		}
		errors.CheckErrorPanic(tree.Commit())

		errors.CheckErrorPanic(tree.RollBack(root))
		value, err = tree.Get([]byte("key"))
		errors.CheckErrorPanic(err)
		if string(value) != "value" {
			b.Fatal("must be value")
		}
		errors.CheckErrorPanic(tree.Close())
	}
}
