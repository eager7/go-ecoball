package datasync

import "github.com/ecoball/go-ecoball/sharding/cell"

type Sync struct {
	syncType int
	cell     *cell.Cell
}

func MakeSync(c *cell.Cell) *Sync {
	return &Sync{cell: c}
}

func (sync *Sync) SyncRequest(blockType int8, fromHeight int64) {

}
