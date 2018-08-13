package erasure


// NodeStat is a statistics object for a Node. Mostly sizes.
type ErasureNodeStat struct {
	Hash           string
	NumLinks       int // number of links in link table
	BlockSize      int // size of the raw, encoded data
	LinksSize      int // size of the links segment
	DataSize       int // size of the data segment
	CumulativeSize int // cumulative size of object and its references
	DataPieces     int
	ParityPieces   int
}
