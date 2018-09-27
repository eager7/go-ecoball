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
	"github.com/ecoball/go-ecoball/dsn/common"
	chunker "gx/ipfs/QmVDjhUMtkRskBFAVNwyXuLSKbeAya7JKPnzAxMKDaK4x4/go-ipfs-chunker"
	"fmt"
	"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	dag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
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

func IpfsAddEraFile(ctx context.Context, fpath string, era uint8) (string, error) {
	adder, err := NewEraAdder(ctx)
	if err != nil {
		return "", err
	}
	fpath = filepath.ToSlash(filepath.Clean(fpath))
	stat, err := os.Lstat(fpath)
	if err != nil {
		return "", err
	}

	var fm EraMetaData
	if era > 0 {
		if stat.Size() < common.EraDataPiece * chunker.DefaultBlockSize {
			fm.PieceSize = uint64(stat.Size() / common.EraDataPiece)
		} else {
			fm.PieceSize = uint64(chunker.DefaultBlockSize)
		}
	}
	if stat.Size() < common.EraDataPiece * chunker.DefaultBlockSize {
		adder.Chunker = fmt.Sprintf("size-%d", fm.PieceSize)
	}
	af, err := files.NewSerialFile(path.Base(fpath), fpath, false, stat)
	if err != nil {
		return "", err
	}

	adder.AddFile(af)
	fileRoot, err := adder.Finalize()
	adder.PinRoot()

	fm.FileSize = uint64(stat.Size())
	fileSize := int(fm.FileSize)
	if fileSize % int(fm.PieceSize) == 0 {
		fm.DataPiece = uint64(fileSize / int(fm.PieceSize))
	} else {
		fm.DataPiece = uint64(fileSize / int(fm.PieceSize) + 1)
	}
	fm.ParityPiece = fm.DataPiece * uint64(era)

	eraRoot, err := adder.EraEnCoding(af, fm)

	return AddMetadataTo(ctx, fileRoot, eraRoot, &fm)
}

func IpfsCatErafile(ctx context.Context, cid string) (io.Reader, error) {
	cater := NewEraCater(ctx)
	return cater.EraDecoding(cid)
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