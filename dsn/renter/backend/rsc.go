package backend

import (
	"context"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/dsn/ipfs/api"
	"github.com/ecoball/go-ecoball/dsn/erasure"
	"bytes"
	"io/ioutil"
	"fmt"
)

func EraCoding(req *renter.RscReq) (string, error) {
	fmt.Println("------>era coding.....")
	fmt.Println(*req)
	ctx := context.Background()
	r, err := api.IpfsCatFile(ctx, req.Cid)
	if err != nil {
		return "", err
	}

	fm := api.EraMetaData{
		PieceSize: req.Chunk,
		FileSize: req.FileSize,
	}

	var fileSize int
	if fileSize % int(fm.PieceSize) == 0 {
		fm.DataPiece = uint64(fileSize / int(fm.PieceSize))
	} else {
		fm.DataPiece = uint64(fileSize / int(fm.PieceSize) + 1)
	}
	fm.ParityPiece = fm.DataPiece * uint64(req.Redundency)

	erCoder, err := erasure.NewRSCode(int(fm.DataPiece), int(fm.ParityPiece))
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	shards, err := erCoder.Encode(b)
	if err != nil {
		return "", err
	}

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
	return  api.AddDagFromReader(ctx, erReader, &fm, req.Cid)
}

func EraDecoding()  {
	
}