package client

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

