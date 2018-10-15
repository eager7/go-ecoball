package renter

type RscReq struct {
	Cid         string  `json:"cid"`
	Redundency  int     `json:"redundency"`
	IsDir       bool    `json:"dir"`
	Chunk       string  `json:"chunk"`
}

