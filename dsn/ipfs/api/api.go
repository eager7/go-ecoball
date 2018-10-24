package api

import (
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/core/corerepo"
	"context"
	"io"
	"path"
	"bytes"
	opt "github.com/ipfs/go-ipfs/core/coreapi/interface/options"
	chunker "gx/ipfs/QmVDjhUMtkRskBFAVNwyXuLSKbeAya7JKPnzAxMKDaK4x4/go-ipfs-chunker"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	dag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	importer "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/importer"
	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	"github.com/ecoball/go-ecoball/dsn/erasure"
)

var dsnIpfsApi coreiface.CoreAPI
var dsnIpfsNode *core.IpfsNode

type EraMetaData struct {
	FileSize    uint64
	DataPiece   uint64
	ParityPiece uint64
	PieceSize   uint64
}


func StartDsnIpfsService(node *core.IpfsNode)  {
	dsnIpfsNode = node
	dsnIpfsApi = coreapi.NewCoreAPI(node)
}

func IpfsRepoStat(ctx context.Context) (corerepo.Stat, error) {
	return corerepo.RepoStat(ctx, dsnIpfsNode)
}

func IpfsBlockAllKey(ctx context.Context) ([]string, error) {
	var cids []string
	baseBlockService := dsnIpfsNode.BaseBlocks
	allCids, err := baseBlockService.AllKeysChan(ctx)
	if err != nil {
		return nil, err
	}
	for cid := range allCids {
		cids = append(cids, cid.String())
	}
	return cids, nil
}

func IpfsBlockGet(ctx context.Context, p string) (io.Reader, error) {
	path, err := coreiface.ParsePath(p)
	if err != nil {
		return nil, err
	}
	return dsnIpfsApi.Object().Data(ctx, path)
}

func IpfsBlockDel(ctx context.Context, p string) error {
	path, err := coreiface.ParsePath(p)
	if err != nil {
		return err
	}
	return dsnIpfsApi.Block().Rm(ctx, path)
}

func IpfsAbaBlkPut(ctx context.Context, blk []byte) (string, error) {
	r := bytes.NewReader(blk)
	//rp, err := dsnIpfsApi.Dag().Put(ctx, r, opt.Dag.InputEnc("raw"), opt.Dag.Codec(cid.EcoballRawData))
	rp, err := dsnIpfsApi.Object().Put(ctx, r, opt.Object.InputEnc("protobuf"))
	if err != nil {
		return "", err
	}
	cidValue := rp.Root().String()
	return cidValue, nil
}

func BytesForMetadata(m *EraMetaData) []byte {
	return encoding.Marshal(m)
}

func MetadataFromBytes(b []byte) (*EraMetaData, error) {
	m := new(EraMetaData)
	err := encoding.Unmarshal(b, m)
	return m, err
}

func Metadata(ctx context.Context, skey string) (*EraMetaData, error) {
	c, err := cid.Decode(skey)
	if err != nil {
		return nil, err
	}

	nd, err := dsnIpfsNode.DAG.Get(ctx, c)
	if err != nil {
		return nil, err
	}

	pbnd, ok := nd.(*dag.ProtoNode)
	if !ok {
		return nil, dag.ErrNotProtobuf
	}

	return MetadataFromBytes(pbnd.Data())
}

func IpfsCatFile(ctx context.Context, cid string) (io.Reader, error) {
	p, err := coreiface.ParsePath(cid)
	if err != nil {
		return nil, err
	}
	return dsnIpfsApi.Unixfs().Cat(ctx, p)
}

func AddDagFromReader(ctx context.Context, r io.Reader, fm *EraMetaData, fc string) (string, error) {
	p, err := coreiface.ParsePath(fc)
	if err != nil {
		return "", err
	}
	fileNode, err := dsnIpfsApi.Dag().Get(ctx, p)
	if err != nil {
		return "", err
	}
	eraNode, err := importer.BuildDagFromReader(dsnIpfsNode.DAG, chunker.NewSizeSplitter(r, int64(fm.PieceSize)))
	if err != nil {
		return "", err
	}
	return AddMetadataTo(ctx, fileNode, eraNode, fm)
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

func IpfsEraDecoding(ctx context.Context, cid string) (io.Reader, error) {
	meta, err := Metadata(ctx, cid)
	if err != nil {
		return nil, err
	}

	ep, err := coreiface.ParsePath(path.Join(cid, "era"))
	if err != nil {
		return nil, err
	}
	eraNode, err := dsnIpfsApi.Dag().Get(ctx, ep)
	if err != nil {
		return nil, err
	}
	links := eraNode.Links()

	var datas [][]byte
	for i, l := range links {
		p, _ := coreiface.ParsePath(l.Cid.String())
		pnode, err := dsnIpfsApi.Dag().Get(ctx, p)
		if err == nil {
			datas[i] = pnode.RawData()
		}
	}
	buff := new(bytes.Buffer)
	ec, err := erasure.NewRSCode(int(meta.DataPiece), int(meta.ParityPiece))
	if err != nil {
		return nil, err
	}
	err = ec.Recover(datas, meta.FileSize, buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}