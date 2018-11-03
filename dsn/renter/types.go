package renter

import (
	"github.com/ecoball/go-ecoball/dsn/renter/pb"
)
type RscReq struct {
	Cid         string  `json:"cid"`
	Redundency  int     `json:"redundency"`
	IsDir       bool    `json:"dir"`
	Chunk       uint64  `json:"chunk"`
	FileSize    uint64  `json:"filesize"`
}


type AccountStakeRsp struct {
	Result string `json:"result"`
	Stake  uint64 `json:"stake"`
}


type FileContract struct {
	PublicKey   []byte
	Cid         string
	LocalPath   string
	FileSize    uint64
	Redundancy  uint8
	Funds       []byte
	StartAt     uint64
	Expiration  uint64
	AccountName   string
}
func (fc *FileContract) Serialize() ([]byte, error) {
	var pfc pb.FcMessage
	pfc.PublicKey = fc.PublicKey
	pfc.Cid = fc.Cid
	pfc.LocalPath = fc.LocalPath
	pfc.FileSize = fc.FileSize
	pfc.Redundancy = uint32(fc.Redundancy)
	//pfc.Funds, _ = fc.Funds.GobEncode()
	pfc.Funds = fc.Funds
	pfc.StartAt = fc.StartAt
	pfc.Expiration = fc.Expiration
	return pfc.Marshal()
}

func (fc *FileContract) Deserialize(data []byte) error {
	var pfc pb.FcMessage
	err := pfc.Unmarshal(data)
	if err != nil {
		return err
	}
	fc.PublicKey = pfc.PublicKey
	fc.Cid = pfc.Cid
	fc.LocalPath = pfc.LocalPath
	fc.FileSize = pfc.FileSize
	fc.Redundancy = uint8(pfc.Redundancy)
	//err = fc.Funds.GobDecode(pfc.Funds)
	//if err != nil {
	//	return err
	//}
	fc.Funds = pfc.Funds
	fc.StartAt = pfc.StartAt
	fc.Expiration = pfc.Expiration
	return nil
}


