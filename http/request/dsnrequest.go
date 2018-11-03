package request



type DsnAddFileReq struct {

	Cid         string  `json:"cid"`
	Redundency  int     `json:"redundency"`
	IsDir       bool    `json:"dir"`
	Chunk       uint64  `json:"chunk"`
	FileSize    uint64  `json:"filesize"`

}