package host

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"github.com/ecoball/go-ecoball/dsn/host/erasure"
	"github.com/ecoball/go-ecoball/dsn/host/ipfs/api"
	dsncomm "github.com/ecoball/go-ecoball/dsn/common"
)

func RscCoding(req *dsncomm.RscReq) (string, error) {
	ctx := context.Background()
	r, err := api.IpfsCatFile(ctx, req.Cid)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	fm := api.EraMetaData{
		PieceSize: req.Chunk,
		FileSize:  req.FileSize,
	}

	var fileSize int
	fileSize = int(req.FileSize)
	if fileSize%int(fm.PieceSize) == 0 {
		fm.DataPiece = uint64(fileSize / int(fm.PieceSize))
	} else {
		fm.DataPiece = uint64(fileSize/int(fm.PieceSize) + 1)
	}
	fm.ParityPiece = fm.DataPiece * uint64(req.Redundency)

	erCoder, err := erasure.NewRSCode(int(fm.DataPiece), int(fm.ParityPiece))
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	shards, err := erCoder.Encode(b)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	pieces := len(shards)
	p := make([]byte, pieces*len(shards[0]))
	k := 0
	for i := 0; i < pieces; i++ {
		for _, v := range shards[i] {
			p[k] = v
			k++
		}
	}
	erReader := bytes.NewReader(p)
	return api.AddDagFromReader(ctx, erReader, &fm, req.Cid)
}

func RscDecoding(cid string) (io.Reader, error) {
	ctx := context.Background()
	return api.IpfsEraDecoding(ctx, cid)
}
