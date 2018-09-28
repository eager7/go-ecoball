package api

import (
	"context"
	"io/ioutil"

	chunker "gx/ipfs/QmVDjhUMtkRskBFAVNwyXuLSKbeAya7JKPnzAxMKDaK4x4/go-ipfs-chunker"
	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	files "gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit/files"
	"bytes"
	"github.com/ecoball/go-ecoball/dsn/erasure"
	importer "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/importer"
	"github.com/ecoball/go-ecoball/common/elog"
	"github.com/ipfs/go-ipfs/core/coreunix"
	dag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
)

var ecolog  = elog.NewLogger("dsn-api", elog.DebugLog)

type EraAdder struct {
	*coreunix.Adder
}

func NewEraAdder(ctx context.Context) (*EraAdder, error) {
	cdr, err := coreunix.NewAdder(ctx, dsnIpfsNode.Pinning, dsnIpfsNode.Blockstore, dsnIpfsNode.DAG)
	if err != nil {
		return nil, err
	}
	return &EraAdder{
		cdr,
	}, nil
}

func (adder *EraAdder)EraEnCoding(file files.File, fm EraMetaData) (ipld.Node, error) {
	b, err := ioutil.ReadFile(file.FullPath())
	if err != nil {
		return nil, err
	}

	ecolog.Debug("add file: ", file.FullPath(), "size: ", len(b))

	erCoder, err := erasure.NewRSCode(int(fm.DataPiece), int(fm.ParityPiece))
	if err != nil {
		return nil, err
	}
	shards, err := erCoder.Encode(b)
	if err != nil {
		return nil, err
	}
	ecolog.Debug("shard: ", len(shards), "per shard len", len(shards[0]))
	pieces := len(shards)
	p := make([]byte, pieces * len(shards[0]))
	k := 0
	for i := 0; i < pieces; i++ {
		for _, v := range shards[i] {
			p[k] = v
			k++
		}
	}
	erReader := bytes.NewReader(p)
	ecolog.Debug("after era, len: ",len(p))
	nd, err := importer.BuildDagFromReader(dsnIpfsNode.DAG, chunker.NewSizeSplitter(erReader, int64(fm.PieceSize)))
	if err != nil {
		return nil, err
	}
	return nd, nil
}

func AddMetadataTo(ctx context.Context, fileNode ipld.Node, eraNode ipld.Node, m *EraMetaData) (string, error) {
	mdnode := new(dag.ProtoNode)
	mdata := BytesForMetadata(m)

	mdnode.SetData(mdata)
	if err := mdnode.AddNodeLink("file", fileNode); err != nil {
		return "", err
	}

	if err := mdnode.AddNodeLink("era", eraNode); err != nil {
		return "", err
	}

	err := dsnIpfsNode.DAG.Add(ctx, mdnode)
	if err != nil {
		return "", err
	}

	return mdnode.Cid().String(), nil
}

