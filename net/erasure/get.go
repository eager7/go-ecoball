package erasure

import (
	uio "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/io"
	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	ft "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs"
	mdag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
	upb "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/pb"
	"io"
	"context"
	"errors"
	//"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	"io/ioutil"
	"os"
)

type ErasureDagReader struct {
	pbRreader uio.PBDagReader
	//TODO
}

type ErasureWriter struct {
	Dag  ipld.DAGService
	Writer *io.Writer

	ctx context.Context
}

func ErasureDagRecover() error {
	//TODO
	return nil
}

func ErasureRecoverFile(dag  ipld.DAGService, nd *mdag.ProtoNode, fpath string) error {
	ctx := context.Background()
	links := nd.Links()
	var eraLink *ipld.Link
	for _, link := range links {
		if link.Name == "erasure" {
			eraLink = link
			break
		}
	}
	if eraLink == nil {
		err := errors.New("unErasured node")
		return err
	}

	eraLode, err := dag.Get(ctx, eraLink.Cid)
	if err != nil {
		return err
	}

	eraNd := eraLode.(*mdag.ProtoNode)
	fsEraNode, err := ft.FSNodeFromBytes(eraNd.Data())
	if err != nil {
		return err
	}
	stat, err := eraLode.Stat()
	if err != nil {
		return err
	}

	dataSize := stat.DataSize
	data := make([]byte, 0, dataSize)

	eraDagr := NewErasureFileReader(ctx, eraNd, fsEraNode, dag)
	err = eraDagr.ErasureWriteTo(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fpath, data, os.ModePerm)
	return  err
}

func ErasureGetRecover(dag  ipld.DAGService, node ipld.Node, fpath string) error {
	switch nd := node.(type) {
	case *mdag.ProtoNode:
		fsNode, err := ft.FSNodeFromBytes(nd.Data())
		if err != nil {
			return err
		}
		switch fsNode.Type() {
		case upb.Data_File:
			ErasureRecoverFile(dag, nd, fpath)
		}
	}
	return nil
}