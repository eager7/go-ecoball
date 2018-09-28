package api

import (
	"github.com/ecoball/go-ecoball/dsn/erasure"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	"io"
	"context"
	"path"
	"bytes"
)

type EraCater struct {
	ctx context.Context
}

func NewEraCater(ctx context.Context) *EraCater {
	return &EraCater{
		ctx:ctx,
	}
}

func (cat *EraCater) CatFile(cid string) (io.Reader, error) {
	p, err := coreiface.ParsePath(path.Join(cid, "file"))
	if err != nil {
		return nil, err
	}
	fileRoot, err := dsnIpfsApi.Dag().Get(cat.ctx, p)
	if err != nil {
		return nil, err
	}
	fp, err := coreiface.ParsePath(fileRoot.Cid().String())
	if err != nil {
		return nil, err
	}
	r, err := dsnIpfsApi.Unixfs().Cat(cat.ctx, fp)
	if err == nil {
		return r, nil
	}

	return cat.eraDecoding(cid)
}

func (cat *EraCater)eraDecoding(cid string) (io.Reader, error) {
	meta, err := Metadata(cat.ctx, cid)
	if err != nil {
		return nil, err
	}

	ep, err := coreiface.ParsePath(path.Join(cid, "era"))
	if err != nil {
		return nil, err
	}
	eraNode, err := dsnIpfsApi.Dag().Get(cat.ctx, ep)
	if err != nil {
		return nil, err
	}
	links := eraNode.Links()

	var datas [][]byte
	for i, l := range links {
		p, _ := coreiface.ParsePath(l.Cid.String())
		pnode, err := dsnIpfsApi.Dag().Get(cat.ctx, p)
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