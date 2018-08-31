package erasure

import (
	//"io"
	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
	chunker "gx/ipfs/QmVDjhUMtkRskBFAVNwyXuLSKbeAya7JKPnzAxMKDaK4x4/go-ipfs-chunker"
	importer "gx/ipfs/QmSaz8Qg77gGqvDvLKeSAY7ivDEnramSWF6T7TcRwFpHtP/go-unixfs/importer"
	dag "gx/ipfs/QmRy4Qk9hbgFX9NGJRm8rBThrA8PZhNCitMgeRYyZ67s59/go-merkledag"
	"bytes"
	"gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit/files"
	"errors"
	"fmt"
	//"os"
	"io/ioutil"
)

var ErrDataInvalid  = errors.New("Invalid data size")
var ErrOpenfile  = errors.New("failed to open file")


func AddFileWithErasure(ds ipld.DAGService, node ipld.Node, file files.File) error {
	//var reader io.Reader = file
	//var size int64
	//if fi, ok := file.(files.FileInfo); ok {
	//	size = fi.Stat().Size()
	//} else {
	//	return ErrDataInvalid
	//}
	fmt.Printf("file name: %s, fullpath: %s\n", file.FileName(), file.FullPath())

	//fi, err := os.OpenFile(file.FullPath(), os.O_RDONLY, 0600)
	//if err != nil {
	//	return ErrOpenfile
	//}

	//fil, err := fi.Stat()
	//if err != nil {
	//	return ErrOpenfile
	//}

	//size := fil.Size()

	//var datas [][]byte
	//var dataPieces int
	//for {
	//	var data [256]byte
		//n, err := io.ReadFull(reader, data[:])
	//	n, err := fi.Read(data[:])
	//	if n==0 || err == io.ErrUnexpectedEOF {
	//		break
	//	}
	//	datas = append(datas, data[:])
	//	dataPieces = dataPieces + 1
	//}

	//if dataPieces == 0 {
	//	return ErrDataInvalid
	//}
	//if n == 0 || err != nil {
	//	return ErrDataInvalid
	//}


	//size := len(datas)


	//dataPieces := size/int(chunker.DefaultBlockSize)
	//if size%1024 == 0 {
	//	dataPieces = int(size/1024)
	//} else {
	//	dataPieces = int(size/1024 + 1)
	//}
	//This is just for test
	//parityPieces := dataPieces
	//erCoder, err := NewRSCode(dataPieces, parityPieces)
	//if err != nil {
	//	return err
	//}

	b, err := ioutil.ReadFile(file.FullPath())
	if err != nil {
		return ErrDataInvalid
	}

	var dataPieces int
	size := len(b)
	if size%int(DefaultPieceSize) == 0 {
		dataPieces = int(size/int(DefaultPieceSize))
	} else {
		dataPieces = int(size/int(DefaultPieceSize) + 1)
	}
	parityPieces := dataPieces
	erCoder, err := NewRSCode(dataPieces, parityPieces)
	if err != nil {
		return err
	}

	//shards, err := erCoder.enc.Split(b)
	shards, err := erCoder.Encode(b)
	if err != nil {
		return err
	}

	p := make([]byte, (dataPieces + parityPieces) * int(DefaultPieceSize))
	//var parity [parityPieces][1024]byte
	//parity := make([][]byte, 1024)
	//copy(parity, shards[dataPieces:])
	//parity := erCoder.Parity()
	//i := dataPieces
	k := 0
	for i := 0; i < dataPieces + parityPieces; i++ {
		for _, v := range shards[i] {
			p[k] = v
			k++
		}
		//for j := 0; j < 1024; j++ {
			//parity[k][j] = shards[i][j]
			//p = append(p, shards[i][j])

		//}
		//k++
	}
	//for _, v := range parity {
	//	for _, j := range v {
	//		p = append(p, j)
	//	}
	//}
	fmt.Printf("total len: %d, parity len: %d\n", len(shards), len(p))

	erReader := bytes.NewReader(p)
	nd, err := importer.BuildDagFromReader(ds, chunker.DefaultSplitter(erReader))
	if err != nil {
		return err
	}

	pnode := node.(*dag.ProtoNode)
	pnode.AddNodeLink("erasure", nd)

	return nil
}