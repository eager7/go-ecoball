package common_test

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"testing"
)

func TestNewIndex(t *testing.T) {
	var char = []byte("abcdefghigklmnopqrstuvwxyz")
	for i := 0; i < 25; i ++ {
		name := common.NameToIndex(fmt.Sprintf("tester%c", char[i]))
		fmt.Println(name, fmt.Sprintf("%d", name), "shard:", uint64(name)%999, uint64(name)%999%8+1)
	}
	name := common.NameToIndex("root")
	fmt.Println(name, fmt.Sprintf("%d", name), "shard:", uint64(name)%999, uint64(name)%999%2+1)
	name = common.NameToIndex("tester")
	fmt.Println(name, fmt.Sprintf("%d", name), "shard:", uint64(name)%999, uint64(name)%999%2+1)

}
