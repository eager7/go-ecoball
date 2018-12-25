package gossip_test

import (
	"testing"
	"github.com/ecoball/go-ecoball/test/net/gossippull"
	"context"
)

func TestGossip(t *testing.T) {
	gossippull.StartBlockPuller(context.Background())

}
