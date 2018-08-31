package erasure

import (
	"io"
	chunker "gx/ipfs/QmVDjhUMtkRskBFAVNwyXuLSKbeAya7JKPnzAxMKDaK4x4/go-ipfs-chunker"
	"io/ioutil"
)

type erasureSplitter struct {
	r    io.Reader
	size uint32
	err  error
	shards [][]byte
	readedPos int
	esCoder ErasureCoder
}

// NewSizeSplitter returns a new size-based Splitter with the given block size.
func NewErasureSplitter(r io.Reader, size int64, eraCoder ErasureCoder) chunker.Splitter {
	return &erasureSplitter{
		r:    r,
		size: uint32(size),
		esCoder: eraCoder,
	}
}

// NextBytes produces a new chunk.
func (ss *erasureSplitter) NextBytes() ([]byte, error) {
	if ss.err != nil {
		return nil, ss.err
	}

	if ss.readedPos == 0 {
		datas, err := ioutil.ReadAll(ss.r)
		if err != nil {
			return nil, err
		}
		ss.shards, err = ss.esCoder.Encode(datas)
		if err != nil {
			return nil, err
		}
	}

	defer func() {
		ss.readedPos = ss.readedPos + 1
	}()

	if ss.readedPos != ss.esCoder.NumPieces() {
		return ss.shards[ss.readedPos], nil
	}
	return nil, nil
}

// Reader returns the io.Reader associated to this Splitter.
func (ss *erasureSplitter) Reader() io.Reader {
	return ss.r
}
