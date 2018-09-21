package api

import (
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/core/corerepo"
	"context"
	"io"
	"gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit/files"
	"path"
	"path/filepath"
	"os"
	"bytes"
	//"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	opt "github.com/ipfs/go-ipfs/core/coreapi/interface/options"
)

var dsnIpfsApi coreiface.CoreAPI
var dsnIpfsNode *core.IpfsNode

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

func IpfsAddEraFile(ctx context.Context, fpath string, era uint8) (string, error) {
	adder, err := NewEcoAdder(ctx)
	if err != nil {
		return "", err
	}
	adder.SetRedundancy(era)
	fpath = filepath.ToSlash(filepath.Clean(fpath))
	stat, err := os.Lstat(fpath)
	if err != nil {
		return "", err
	}
	af, err := files.NewSerialFile(path.Base(fpath), fpath, false, stat)
	if err != nil {
		return "", err
	}
	adder.AddFile(af)
	dagnode, err := adder.Finalize()
	return dagnode.String(), err
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